package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestOnlyAllowlistedAgedFiles(t *testing.T) {
	user := t.TempDir()
	local := filepath.Join(user, "AppData", "Local")
	temp := filepath.Join(local, "Temp")
	outside := filepath.Join(user, "Documents")
	for _, d := range []string{temp, outside} {
		if err := os.MkdirAll(d, 0700); err != nil {
			t.Fatal(err)
		}
	}
	old := filepath.Join(temp, "old.tmp")
	recent := filepath.Join(temp, "new.tmp")
	protected := filepath.Join(outside, "keep.txt")
	for _, p := range []string{old, recent, protected} {
		if err := os.WriteFile(p, []byte("do not remove unless old temporary"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	past := time.Now().Add(-20 * 24 * time.Hour)
	if err := os.Chtimes(old, past, past); err != nil {
		t.Fatal(err)
	}
	// Try a directory link escaping into Documents. Never follow it.
	_ = os.Symlink(outside, filepath.Join(temp, "a-link"))
	r := cleanTargets(targetsForLocal(local), time.Now())
	if r.Deleted != 1 {
		t.Fatalf("deleted=%d, want 1: %+v", r.Deleted, r)
	}
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Fatalf("old temp file not removed: %v", err)
	}
	for _, p := range []string{recent, protected} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("protected file %q touched: %v", p, err)
		}
	}
}

func TestRejectsUnexpectedLocalPath(t *testing.T) {
	for _, x := range []string{"", "/", "/tmp", "/Downloads", "relative/AppData/Local"} {
		if len(targetsForLocal(x)) != 0 {
			t.Fatalf("unexpected accepted path: %s", x)
		}
	}
}

func TestDoesNotRemoveNewCrashDumps(t *testing.T) {
	local := filepath.Join(t.TempDir(), "AppData", "Local")
	dir := filepath.Join(local, "CrashDumps")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	recent := filepath.Join(dir, "new.dmp")
	old := filepath.Join(dir, "old.dmp")
	for _, p := range []string{recent, old} {
		if err := os.WriteFile(p, []byte("data"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	past := time.Now().Add(-32 * 24 * time.Hour)
	if err := os.Chtimes(old, past, past); err != nil {
		t.Fatal(err)
	}
	r := cleanTargets(targetsForLocal(local), time.Now())
	if r.Deleted != 1 {
		t.Fatalf("unexpected deletion: %+v", r)
	}
	if _, err := os.Stat(recent); err != nil {
		t.Fatal(err)
	}
}
