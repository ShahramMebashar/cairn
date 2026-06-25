package gitctx

import "testing"

func TestParseNameStatusHandlesSpacesAndRenames(t *testing.T) {
	raw := []byte("M\x00web/src/App.tsx\x00A\x00docs/new guide.md\x00R100\x00old name.go\x00new name.go\x00")

	got := parseNameStatus(raw)
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3: %+v", len(got), got)
	}
	if got[0] != (ChangedFile{Status: "M", Path: "web/src/App.tsx"}) {
		t.Fatalf("first = %+v", got[0])
	}
	if got[1] != (ChangedFile{Status: "A", Path: "docs/new guide.md"}) {
		t.Fatalf("second = %+v", got[1])
	}
	if got[2] != (ChangedFile{Status: "R100", OldPath: "old name.go", Path: "new name.go"}) {
		t.Fatalf("rename = %+v", got[2])
	}
}

func TestParseStatusRenameOrientation(t *testing.T) {
	// porcelain -z emits renames as "XY NEW\0OLD" — the reverse of diff --name-status.
	raw := []byte("R  renamed.txt\x00orig.txt\x00 M web/src/App.tsx\x00")

	got := parseStatus(raw)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2: %+v", len(got), got)
	}
	if got[0] != (ChangedFile{Status: "R", OldPath: "orig.txt", Path: "renamed.txt"}) {
		t.Fatalf("rename = %+v", got[0])
	}
	if got[1] != (ChangedFile{Status: "M", Path: "web/src/App.tsx"}) {
		t.Fatalf("modified = %+v", got[1])
	}
}

func TestSplitNULDropsTrailingEmpty(t *testing.T) {
	got := splitNUL([]byte("a\x00b\x00"))
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("split = %#v", got)
	}
}
