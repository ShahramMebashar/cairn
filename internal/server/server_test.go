package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"cairn/internal/mcp"
	"cairn/internal/repo"
	"cairn/internal/session"
)

// webIDRe matches the time-ordered ids minted by the store (prefix + 16 base32 chars).
var webIDRe = regexp.MustCompile(`^WEB-[0-9a-z]{10}$`)

type runResp struct {
	Runs []struct {
		File     string `json:"file"`
		At       string `json:"at"`
		Cmd      string `json:"cmd"`
		Cwd      string `json:"cwd"`
		Exit     int    `json:"exit"`
		TimedOut bool   `json:"timedout"`
		Duration string `json:"duration"`
		Output   string `json:"output"`
	} `json:"runs"`
}

func TestRunsEndpointParsesLogsNewestFirst(t *testing.T) {
	s, h := newServer(t)
	var st statusResp
	call(t, h, "POST", "/api/init", `{"prefix":"WEB"}`, &st)

	runsDir := filepath.Join(s.defaultRoot, ".cairn", "runs")
	if err := os.MkdirAll(runsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(runsDir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("WEB-001-20260621-185736.523.log",
		"cmd: echo old\ncwd: /repo\nexit: 1  timedout: false  duration: 500ms\n----\nold output\n")
	write("WEB-001-20260621-190000.000.log",
		"cmd: echo new\ncwd: /repo\nexit: 0  timedout: false  duration: 1.2s\n----\nnew output\n")
	// A different task's log must not leak into WEB-001's runs.
	write("WEB-002-20260621-190001.000.log",
		"cmd: nope\ncwd: /repo\nexit: 0  timedout: false  duration: 1s\n----\n")

	var resp runResp
	call(t, h, "GET", "/api/tasks/WEB-001/runs", "", &resp)
	if len(resp.Runs) != 2 {
		t.Fatalf("want 2 runs, got %d: %+v", len(resp.Runs), resp.Runs)
	}
	newest := resp.Runs[0]
	if newest.Cmd != "echo new" || newest.Exit != 0 || newest.Output != "new output\n" {
		t.Fatalf("newest run: %+v", newest)
	}
	if newest.At != "2026-06-21T19:00:00Z" {
		t.Fatalf("newest At: %q", newest.At)
	}
	older := resp.Runs[1]
	if older.Cmd != "echo old" || older.Exit != 1 || older.TimedOut {
		t.Fatalf("older run: %+v", older)
	}
}

func TestRunsEndpointEmptyWhenNoRuns(t *testing.T) {
	_, h := newServer(t)
	var st statusResp
	call(t, h, "POST", "/api/init", `{"prefix":"WEB"}`, &st)
	var resp runResp
	call(t, h, "GET", "/api/tasks/WEB-001/runs", "", &resp)
	if len(resp.Runs) != 0 {
		t.Fatalf("want 0 runs, got %d", len(resp.Runs))
	}
}

func newServer(t *testing.T) (*Server, http.Handler) {
	t.Helper()
	s := New(t.TempDir(), "human:test")
	return s, s.Handler()
}

func TestStatusThenInit(t *testing.T) {
	s, h := newServer(t)

	var st statusResp
	call(t, h, "GET", "/api/status", "", &st)
	if st.Initialized {
		t.Fatal("should not be initialized")
	}
	if st.SuggestedPrefix == "" {
		t.Fatal("expected a suggested prefix")
	}

	call(t, h, "POST", "/api/init", `{"prefix":"WEB"}`, &st)
	if !st.Initialized || st.Prefix != "WEB" {
		t.Fatalf("after init: %+v", st)
	}
	if !repo.IsInitialized(s.defaultRoot) {
		t.Fatal("workspace not created on disk")
	}
	// Status now carries the config states for the board.
	call(t, h, "GET", "/api/status", "", &st)
	if len(st.States) == 0 || st.Initial != "backlog" {
		t.Fatalf("status missing config: %+v", st)
	}
}

func TestTaskLifecycleOverHTTP(t *testing.T) {
	_, h := newServer(t)
	var st statusResp
	call(t, h, "POST", "/api/init", `{"prefix":"WEB"}`, &st)

	// Create with a passing check.
	var created taskDTO
	call(t, h, "POST", "/api/tasks", `{"title":"ship it","checks":[{"desc":"t","cmd":"exit 0"}]}`, &created)
	if !webIDRe.MatchString(created.ID) || created.Status != "backlog" {
		t.Fatalf("created: %+v", created)
	}

	// List shows it.
	var list struct {
		Tasks []taskDTO `json:"tasks"`
	}
	call(t, h, "GET", "/api/tasks", "", &list)
	if len(list.Tasks) != 1 {
		t.Fatalf("want 1 task, got %d", len(list.Tasks))
	}

	// Claim.
	var claimed taskDTO
	call(t, h, "POST", "/api/tasks/"+created.ID+"/claim", "", &claimed)
	if claimed.Assignee != "human:test" {
		t.Fatalf("assignee: %q", claimed.Assignee)
	}

	// Transition to done auto-runs the passing check and closes.
	var done taskDTO
	call(t, h, "POST", "/api/tasks/"+created.ID+"/transition", `{"to":"done"}`, &done)
	if done.Status != "done" || done.Checks[0].Result != "pass" {
		t.Fatalf("transition: %+v", done)
	}
}

func TestSessionLifecycleOverHTTP(t *testing.T) {
	_, h := newServer(t)
	var st statusResp
	call(t, h, "POST", "/api/init", `{"prefix":"WEB"}`, &st)
	var created taskDTO
	call(t, h, "POST", "/api/tasks", `{"title":"observable"}`, &created)

	var identity mcp.Identity
	call(t, h, "GET", "/api/identity", "", &identity)
	if identity.Actor != "human:test" {
		t.Fatalf("identity = %+v", identity)
	}

	var begun mcp.SessionView
	call(t, h, "POST", "/api/tasks/"+created.ID+"/sessions/begin",
		`{"expectedActor":"human:test","client":"test","idempotencyKey":"begin-1"}`, &begun)
	if begun.TaskID != created.ID || begun.Status != session.StatusActive {
		t.Fatalf("begun = %+v", begun)
	}

	var heartbeat mcp.SessionView
	call(t, h, "POST", "/api/sessions/"+begun.ID+"/heartbeat", `{"progress":"running tests"}`, &heartbeat)
	if heartbeat.Live == nil || heartbeat.Live.Progress != "running tests" {
		t.Fatalf("heartbeat = %+v", heartbeat)
	}

	var finished mcp.SessionView
	call(t, h, "POST", "/api/sessions/"+begun.ID+"/finish", `{"summary":"implemented"}`, &finished)
	if finished.Status != session.StatusFinished {
		t.Fatalf("finished = %+v", finished)
	}

	var sessions struct {
		Sessions []mcp.SessionView `json:"sessions"`
	}
	call(t, h, "GET", "/api/tasks/"+created.ID+"/sessions", "", &sessions)
	if len(sessions.Sessions) != 1 || sessions.Sessions[0].ID != begun.ID {
		t.Fatalf("sessions = %+v", sessions.Sessions)
	}

	var list struct {
		Tasks []taskDTO `json:"tasks"`
	}
	call(t, h, "GET", "/api/tasks?execution=awaiting_review", "", &list)
	if len(list.Tasks) != 1 || list.Tasks[0].SessionID != begun.ID || list.Tasks[0].ExecutionState != mcp.ExecutionAwaitingReview {
		t.Fatalf("awaiting-review tasks = %+v", list.Tasks)
	}
}

func TestSessionIdentityMismatchIs422(t *testing.T) {
	_, h := newServer(t)
	var st statusResp
	call(t, h, "POST", "/api/init", `{"prefix":"WEB"}`, &st)
	var created taskDTO
	call(t, h, "POST", "/api/tasks", `{"title":"observable"}`, &created)

	code, body := raw(h, "POST", "/api/tasks/"+created.ID+"/sessions/begin",
		`{"expectedActor":"agent:codex","idempotencyKey":"begin-1"}`)
	if code != http.StatusUnprocessableEntity || !strings.Contains(body, "bound") {
		t.Fatalf("identity mismatch = %d %s", code, body)
	}
}

func TestTransitionRefusalIs422(t *testing.T) {
	_, h := newServer(t)
	var st statusResp
	call(t, h, "POST", "/api/init", `{"prefix":"WEB"}`, &st)
	var created taskDTO
	call(t, h, "POST", "/api/tasks", `{"title":"x","checks":[{"desc":"bad","cmd":"exit 1"}]}`, &created)

	code, body := raw(h, "POST", "/api/tasks/"+created.ID+"/transition", `{"to":"done"}`)
	if code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422; body=%s", code, body)
	}
	if !strings.Contains(body, "checks") {
		t.Fatalf("expected gate reason in body, got %s", body)
	}
}

func TestGetMissingIs404(t *testing.T) {
	_, h := newServer(t)
	var st statusResp
	call(t, h, "POST", "/api/init", "", &st)
	if code, _ := raw(h, "GET", "/api/tasks/NOPE-999", ""); code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", code)
	}
}

func TestAttestEndpointUnblocksClose(t *testing.T) {
	_, h := newServer(t)
	var st statusResp
	call(t, h, "POST", "/api/init", `{"prefix":"WEB"}`, &st)

	var created taskDTO
	call(t, h, "POST", "/api/tasks", `{"title":"review me","checks":[{"desc":"human review","type":"manual"}]}`, &created)

	// Closing is refused while the manual check is pending.
	if code, _ := raw(h, "POST", "/api/tasks/"+created.ID+"/transition", `{"to":"done"}`); code != http.StatusUnprocessableEntity {
		t.Fatalf("close before attest = %d, want 422", code)
	}

	var attested taskDTO
	call(t, h, "POST", "/api/tasks/"+created.ID+"/attest", `{"index":0,"pass":true}`, &attested)
	if attested.Checks[0].Result != "pass" {
		t.Fatalf("after attest: %+v", attested.Checks[0])
	}

	var done taskDTO
	call(t, h, "POST", "/api/tasks/"+created.ID+"/transition", `{"to":"done"}`, &done)
	if done.Status != "done" {
		t.Fatalf("close after attest: %+v", done)
	}
}

func TestListIncludesUpdatedAt(t *testing.T) {
	_, h := newServer(t)
	var st statusResp
	call(t, h, "POST", "/api/init", `{"prefix":"WEB"}`, &st)
	var created taskDTO
	call(t, h, "POST", "/api/tasks", `{"title":"x"}`, &created)

	var list struct {
		Tasks []taskDTO `json:"tasks"`
	}
	call(t, h, "GET", "/api/tasks", "", &list)
	if len(list.Tasks) != 1 || list.Tasks[0].UpdatedAt == "" {
		t.Fatalf("expected updatedAt on list, got %+v", list.Tasks)
	}
	first := list.Tasks[0].UpdatedAt

	// A note appends a newer provenance entry; updatedAt must not go backwards.
	call(t, h, "POST", "/api/tasks/"+created.ID+"/note", `{"text":"hi"}`, &created)
	call(t, h, "GET", "/api/tasks", "", &list)
	if list.Tasks[0].UpdatedAt < first {
		t.Fatalf("updatedAt regressed: %q < %q", list.Tasks[0].UpdatedAt, first)
	}
}

func raw(h http.Handler, method, path, body string) (int, string) {
	var r *http.Request
	if body != "" {
		r = httptest.NewRequest(method, path, strings.NewReader(body))
	} else {
		r = httptest.NewRequest(method, path, nil)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w.Code, w.Body.String()
}

func call(t *testing.T, h http.Handler, method, path, body string, out any) {
	t.Helper()
	code, b := raw(h, method, path, body)
	if code != http.StatusOK {
		t.Fatalf("%s %s -> %d: %s", method, path, code, b)
	}
	if err := json.Unmarshal([]byte(b), out); err != nil {
		t.Fatalf("decode: %v", err)
	}
}

func TestUpdateFieldsAndStatusActor(t *testing.T) {
	_, h := newServer(t)
	var st statusResp
	call(t, h, "POST", "/api/init", `{"prefix":"WEB"}`, &st)
	if st.Actor != "human:test" {
		t.Fatalf("status actor = %q, want human:test", st.Actor)
	}
	var created taskDTO
	call(t, h, "POST", "/api/tasks", `{"title":"x"}`, &created)
	var updated taskDTO
	call(t, h, "POST", "/api/tasks/"+created.ID+"/update", `{"priority":"high","labels":["backend"]}`, &updated)
	if updated.Priority != "high" || len(updated.Labels) != 1 || updated.Labels[0] != "backend" {
		t.Fatalf("update not applied: %+v", updated)
	}
}

func TestUpdateValidationAndGetUpdatedAt(t *testing.T) {
	_, h := newServer(t)
	var st statusResp
	call(t, h, "POST", "/api/init", `{"prefix":"WEB"}`, &st)
	var created taskDTO
	call(t, h, "POST", "/api/tasks", `{"title":"x"}`, &created)

	// single-task GET carries updatedAt (regression: dtoFromDoc must set it)
	var got taskDTO
	call(t, h, "GET", "/api/tasks/"+created.ID, "", &got)
	if got.UpdatedAt == "" {
		t.Fatal("GET task missing updatedAt")
	}

	// create with a missing parent -> 422, not 500
	if code, body := raw(h, "POST", "/api/tasks", `{"title":"y","parent":"NOPE-1"}`); code != http.StatusUnprocessableEntity {
		t.Fatalf("missing parent status = %d, want 422; body=%s", code, body)
	}
	// invalid priority -> 422
	if code, _ := raw(h, "POST", "/api/tasks/"+created.ID+"/update", `{"priority":"ASAP"}`); code != http.StatusUnprocessableEntity {
		t.Fatalf("invalid priority status = %d, want 422", code)
	}
}

func TestEditTaskTitleBodyChecks(t *testing.T) {
	_, h := newServer(t)
	var st statusResp
	call(t, h, "POST", "/api/init", `{"prefix":"WEB"}`, &st)
	var created taskDTO
	call(t, h, "POST", "/api/tasks", `{"title":"x","body":"old\n"}`, &created)

	var updated taskDTO
	call(t, h, "POST", "/api/tasks/"+created.ID+"/update",
		`{"title":"renamed","body":"new\n","checks":[{"desc":"tests","cmd":"go test ./..."}]}`, &updated)
	if updated.Title != "renamed" || updated.Body != "new\n" || len(updated.Checks) != 1 {
		t.Fatalf("edit not applied: %+v", updated)
	}

	// empty title -> 422
	if code, _ := raw(h, "POST", "/api/tasks/"+created.ID+"/update", `{"title":"  "}`); code != http.StatusUnprocessableEntity {
		t.Fatalf("empty title status = %d, want 422", code)
	}
}

func TestDeleteTaskEndpoint(t *testing.T) {
	_, h := newServer(t)
	var st statusResp
	call(t, h, "POST", "/api/init", `{"prefix":"WEB"}`, &st)
	var parent, child taskDTO
	call(t, h, "POST", "/api/tasks", `{"title":"parent"}`, &parent)
	call(t, h, "POST", "/api/tasks", `{"title":"child","parent":"`+parent.ID+`"}`, &child)

	// parent blocked by child -> 422
	if code, _ := raw(h, "DELETE", "/api/tasks/"+parent.ID, ""); code != http.StatusUnprocessableEntity {
		t.Fatalf("delete blocked status = %d, want 422", code)
	}
	// child deletes -> 200, then it's gone -> 404
	if code, b := raw(h, "DELETE", "/api/tasks/"+child.ID, ""); code != http.StatusOK {
		t.Fatalf("delete child = %d: %s", code, b)
	}
	if code, _ := raw(h, "GET", "/api/tasks/"+child.ID, ""); code != http.StatusNotFound {
		t.Fatalf("get deleted = %d, want 404", code)
	}
	// deleting a missing task -> 404
	if code, _ := raw(h, "DELETE", "/api/tasks/NOPE-1", ""); code != http.StatusNotFound {
		t.Fatalf("delete missing = %d, want 404", code)
	}
}

func TestEditAndDeleteNoteEndpoint(t *testing.T) {
	_, h := newServer(t)
	var st statusResp
	call(t, h, "POST", "/api/init", `{"prefix":"WEB"}`, &st)
	var created taskDTO
	call(t, h, "POST", "/api/tasks", `{"title":"x"}`, &created)
	var noted taskDTO
	call(t, h, "POST", "/api/tasks/"+created.ID+"/note", `{"text":"first"}`, &noted)
	noteID := noted.Provenance[len(noted.Provenance)-1].ID
	if noteID == "" {
		t.Fatal("note missing id")
	}

	var edited taskDTO
	call(t, h, "PATCH", "/api/tasks/"+created.ID+"/notes/"+noteID, `{"text":"edited"}`, &edited)
	last := edited.Provenance[len(edited.Provenance)-1]
	if last.Text != "edited" || last.EditedAt == "" {
		t.Fatalf("note not edited: %+v", last)
	}

	// editing a system entry (the created entry, index 0) by index -> 422
	if code, _ := raw(h, "PATCH", "/api/tasks/"+created.ID+"/notes/-?index=0", `{"text":"nope"}`); code != http.StatusUnprocessableEntity {
		t.Fatalf("edit system entry status = %d, want 422", code)
	}

	var deleted taskDTO
	call(t, h, "DELETE", "/api/tasks/"+created.ID+"/notes/"+noteID, "", &deleted)
	for _, p := range deleted.Provenance {
		if p.ID == noteID {
			t.Fatal("note not deleted")
		}
	}
	// deleting a missing note -> 404
	if code, _ := raw(h, "DELETE", "/api/tasks/"+created.ID+"/notes/n_missing", ""); code != http.StatusNotFound {
		t.Fatalf("delete missing note = %d, want 404", code)
	}
}

func TestReorderEndpoint(t *testing.T) {
	_, h := newServer(t)
	var st statusResp
	call(t, h, "POST", "/api/init", `{"prefix":"WEB"}`, &st)
	var created taskDTO
	call(t, h, "POST", "/api/tasks", `{"title":"x"}`, &created)
	var got taskDTO
	call(t, h, "POST", "/api/tasks/"+created.ID+"/reorder", `{"rank":2048}`, &got)
	if got.Rank != 2048 {
		t.Fatalf("rank = %v, want 2048", got.Rank)
	}
}

func TestPerRequestActor(t *testing.T) {
	_, h := newServer(t)
	var st statusResp
	call(t, h, "POST", "/api/init", `{"prefix":"WEB"}`, &st)
	if st.SuggestedActor == "" || !strings.HasPrefix(st.SuggestedActor, "human:") {
		t.Fatalf("suggestedActor = %q, want human:*", st.SuggestedActor)
	}

	// A write carrying X-Cairn-Actor is attributed to that actor.
	r := httptest.NewRequest("POST", "/api/tasks", strings.NewReader(`{"title":"x"}`))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-Cairn-Actor", "human:ali")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}
	var created taskDTO
	_ = json.Unmarshal(w.Body.Bytes(), &created)

	var got taskDTO
	call(t, h, "GET", "/api/tasks/"+created.ID, "", &got)
	if got.Provenance[0].Who != "human:ali" {
		t.Fatalf("provenance who = %q, want human:ali", got.Provenance[0].Who)
	}

	// No header -> falls back to the server default (human:test from newServer).
	var created2 taskDTO
	call(t, h, "POST", "/api/tasks", `{"title":"y"}`, &created2)
	var got2 taskDTO
	call(t, h, "GET", "/api/tasks/"+created2.ID, "", &got2)
	if got2.Provenance[0].Who != "human:test" {
		t.Fatalf("fallback who = %q, want human:test", got2.Provenance[0].Who)
	}
}

func TestSanitizeActor(t *testing.T) {
	if got := sanitizeActor("  human:shahram  "); got != "human:shahram" {
		t.Fatalf("trim = %q", got)
	}
	if got := sanitizeActor("human:ali\ninjected: true"); strings.ContainsAny(got, "\n\r") {
		t.Fatalf("newline not stripped: %q", got)
	}
	if got := sanitizeActor("   "); got != "" {
		t.Fatalf("empty = %q, want \"\"", got)
	}
}
