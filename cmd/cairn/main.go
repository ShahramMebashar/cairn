// Command cairn is the single cairn binary. It serves the repo's task graph to two
// front-ends over the same rule-set: agents via MCP (stdio) and the web UI via HTTP.
//
// Usage:
//
//	cairn init  [--prefix PROJ] [--repo .]          # scaffold .cairn/ in a project
//	cairn serve [--actor agent:claude-1] [--client claude] [--repo .] # MCP server over stdio
//	cairn web   [--addr :8080] [--repo .]           # HTTP server for the web UI
//
// Identity for MCP writes is injected once via --actor and stamped onto every write as
// provenance (SPEC §7); it is never passed per call.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"cairn/internal/mcp"
	"cairn/internal/repo"
	"cairn/internal/server"
	"cairn/internal/store"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "cairn:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return usage()
	}
	cmd, rest := args[0], args[1:]
	switch cmd {
	case "init":
		return runInit(rest)
	case "serve":
		return runServe(rest)
	case "web":
		return runWeb(rest)
	case "-h", "--help", "help":
		return usage()
	default:
		return fmt.Errorf("unknown command %q (want init, serve, or web)", cmd)
	}
}

func usage() error {
	fmt.Fprint(os.Stderr, `cairn — cairn task tool

  cairn init  [--prefix PROJ] [--repo .]            scaffold .cairn/ in a project
  cairn serve [--actor agent:claude-1] [--client claude] [--repo .]
                                                    MCP server over stdio (auto-inits)
  cairn web   [--addr :8080] [--repo .]             HTTP server for the web UI
`)
	return nil
}

func runInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	prefix := fs.String("prefix", "", "id prefix (default: derived from the project folder name)")
	repoRoot := fs.String("repo", ".", "repo root to initialize")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := repo.Init(*repoRoot, *prefix); err != nil {
		return err
	}
	fmt.Printf("initialized cairn workspace in %s (prefix %s)\n", *repoRoot, repo.CurrentPrefix(*repoRoot))
	return nil
}

func runServe(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	actor := fs.String("actor", "", "identity for writes, e.g. agent:claude-1 or human:shah")
	client := fs.String("client", "", "agent client identity, e.g. codex or claude")
	repoRoot := fs.String("repo", ".", "repo root containing .cairn/")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *actor == "" {
		return fmt.Errorf("--actor is required (e.g. agent:claude-1)")
	}
	// Auto-init so a freshly opened project just works.
	if err := repo.Init(*repoRoot, ""); err != nil {
		return err
	}

	svc := mcp.NewServiceWithClient(store.New(*repoRoot), *actor, *client, nil)
	srv := mcp.NewServer(svc)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// A client closing the pipe (EOF) or a Ctrl-C is a normal shutdown, not a failure.
	err := srv.Run(ctx, &mcpsdk.StdioTransport{})
	if errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}

func runWeb(args []string) error {
	fs := flag.NewFlagSet("web", flag.ContinueOnError)
	addr := fs.String("addr", ":8080", "address to listen on (a busy port falls back to the next free one)")
	repoRoot := fs.String("repo", ".", "default repo root (the UI can open other folders)")
	actor := fs.String("actor", "human:web", "identity stamped on web-driven writes")
	parentWatch := fs.Bool("parent-watch", false, "shut down when stdin closes (set by the desktop shell)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// As a Tauri sidecar the desktop app holds our stdin open; when it exits the
	// pipe closes (EOF) and we shut down, so the server never outlives the app.
	// Gated so terminal/CI runs (stdin may be /dev/null → instant EOF) are unaffected.
	if *parentWatch || os.Getenv("CAIRN_PARENT_WATCH") != "" {
		go func() {
			io.Copy(io.Discard, os.Stdin)
			cancel()
		}()
	}

	ln, err := listenWithFallback(*addr)
	if err != nil {
		return err
	}

	// One machine-parseable line on stdout for the desktop shell to read (the port
	// may differ from the request after fallback); the human line stays on stderr.
	_, port, _ := net.SplitHostPort(ln.Addr().String())
	url := "http://127.0.0.1:" + port
	fmt.Printf("CAIRN_WEB_URL=%s\n", url)
	os.Stdout.Sync()
	fmt.Fprintf(os.Stderr, "cairn web listening on %s (default repo %s)\n", url, *repoRoot)

	srv := &http.Server{Handler: server.New(*repoRoot, *actor).Handler()}
	go func() {
		<-ctx.Done()
		sctx, c := context.WithTimeout(context.Background(), 5*time.Second)
		defer c()
		srv.Shutdown(sctx)
	}()
	if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// listenWithFallback binds addr; if its port is taken it scans the next 20 ports,
// then falls back to an OS-assigned one — so the desktop app always comes up even
// when the preferred port (7777) is already in use.
func listenWithFallback(addr string) (net.Listener, error) {
	ln, err := net.Listen("tcp", addr)
	if err == nil {
		return ln, nil
	}
	host, portStr, perr := net.SplitHostPort(addr)
	if perr != nil {
		return nil, err
	}
	base, aerr := strconv.Atoi(portStr)
	if aerr != nil || base == 0 {
		return nil, err
	}
	for p := base + 1; p <= base+20; p++ {
		if l, e := net.Listen("tcp", net.JoinHostPort(host, strconv.Itoa(p))); e == nil {
			return l, nil
		}
	}
	return net.Listen("tcp", net.JoinHostPort(host, "0"))
}
