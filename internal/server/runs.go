package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"cairn/internal/store"
)

// stampLayout matches check.Runner's run-log filename stamp (<id>-<stamp>.log).
const stampLayout = "20060102-150405.000"

// runDTO is one parsed check-run log. Output is the tail captured by the check
// runner; the header fields are parsed from the log's leading lines (SPEC §6).
type runDTO struct {
	File     string `json:"file"`
	At       string `json:"at,omitempty"`
	Cmd      string `json:"cmd,omitempty"`
	Cwd      string `json:"cwd,omitempty"`
	Exit     int    `json:"exit"`
	TimedOut bool   `json:"timedout"`
	Duration string `json:"duration,omitempty"`
	Output   string `json:"output,omitempty"`
}

// handleRuns returns the run logs for a task, newest-first. The task file stores
// only pass/fail (SPEC §149); full output lives in gitignored .cairn/runs logs,
// which this endpoint reads and parses. A missing runs dir yields an empty list.
func (s *Server) handleRuns(w http.ResponseWriter, r *http.Request) {
	_, root, ok := s.svcFor(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	runsDir := store.New(root).RunsDir()

	matches, err := filepath.Glob(filepath.Join(runsDir, id+"-*.log"))
	if err != nil {
		writeErr(w, err)
		return
	}
	// Filenames embed the stamp, so lexical-descending == newest-first.
	sort.Sort(sort.Reverse(sort.StringSlice(matches)))

	runs := make([]runDTO, 0, len(matches))
	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			continue // log vanished mid-list; skip it rather than fail the request
		}
		runs = append(runs, parseRun(id, filepath.Base(path), string(data)))
	}
	writeJSON(w, http.StatusOK, map[string]any{"runs": runs})
}

// parseRun turns a run-log file into a runDTO. The header format is fixed by
// check.Runner.writeLog; an unparseable header degrades to the raw body so output
// is never lost.
func parseRun(id, file, content string) runDTO {
	run := runDTO{File: file}

	// Timestamp comes from the filename: <id>-<stamp>.log.
	stamp := strings.TrimSuffix(strings.TrimPrefix(file, id+"-"), ".log")
	if at, err := time.ParseInLocation(stampLayout, stamp, time.UTC); err == nil {
		run.At = at.UTC().Format(time.RFC3339)
	}

	head, body, found := strings.Cut(content, "\n----\n")
	if !found {
		run.Output = content // no recognizable header; keep everything as output
		return run
	}
	run.Output = body

	for line := range strings.SplitSeq(head, "\n") {
		switch {
		case strings.HasPrefix(line, "cmd: "):
			run.Cmd = strings.TrimPrefix(line, "cmd: ")
		case strings.HasPrefix(line, "cwd: "):
			run.Cwd = strings.TrimPrefix(line, "cwd: ")
		case strings.HasPrefix(line, "exit: "):
			fmt.Sscanf(line, "exit: %d  timedout: %t  duration: %s",
				&run.Exit, &run.TimedOut, &run.Duration)
		}
	}
	return run
}
