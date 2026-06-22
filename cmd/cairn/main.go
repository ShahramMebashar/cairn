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
	"os"
	"os/signal"

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
	addr := fs.String("addr", ":8080", "address to listen on")
	repoRoot := fs.String("repo", ".", "default repo root (the UI can open other folders)")
	actor := fs.String("actor", "human:web", "identity stamped on web-driven writes")
	if err := fs.Parse(args); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "cairn web listening on %s (default repo %s)\n", *addr, *repoRoot)
	return server.New(*repoRoot, *actor).Run(*addr)
}
