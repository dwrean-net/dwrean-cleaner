package main

import (
	"os"
	"path/filepath"
	"strings"
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

func TestExtraTargetsCacheOnlyAndBrowserRunning(t *testing.T) {
	user := t.TempDir()
	local := filepath.Join(user, "AppData", "Local")
	chrome := filepath.Join(local, "Google", "Chrome", "User Data", "Default")
	cache := filepath.Join(chrome, "Cache", "Cache_Data")
	cookies := filepath.Join(chrome, "Network")
	gpu := filepath.Join(local, "NVIDIA", "DXCache")
	for _, p := range []string{cache, cookies, gpu} {
		if err := os.MkdirAll(p, 0700); err != nil { t.Fatal(err) }
	}
	oldCache := filepath.Join(cache, "old_cache_file")
	oldCookie := filepath.Join(cookies, "Cookies")
	oldGpu := filepath.Join(gpu, "old_shader")
	for _, p := range []string{oldCache, oldCookie, oldGpu} {
		if err := os.WriteFile(p, []byte("data"), 0600); err != nil { t.Fatal(err) }
		past := time.Now().Add(-60 * 24 * time.Hour)
		if err := os.Chtimes(p, past, past); err != nil { t.Fatal(err) }
	}
	r := cleanTargets(fullTargetsForLocal(local, map[string]bool{"chrome":true}), time.Now())
	if r.Deleted != 1 { t.Fatalf("while Chrome runs, expected only NVIDIA deleted: %+v",r) }
	if _, err := os.Stat(oldCache); err != nil { t.Fatal("cache deleted while browser running",err) }
	r = cleanTargets(fullTargetsForLocal(local, map[string]bool{}), time.Now())
	if r.Deleted != 1 { t.Fatalf("expected Chrome cache to be deleted: %+v",r) }
	if _,err:=os.Stat(oldCookie); err != nil { t.Fatal("cookie database must not be deleted",err) }
}

func TestExtraTargetsDoesNotFollowJunctionOrSymlink(t *testing.T) {
	user:=t.TempDir()
	local:=filepath.Join(user,"AppData","Local")
	outside:=filepath.Join(user,"Documents")
	if err:=os.MkdirAll(outside,0700);err!=nil {t.Fatal(err)}
	file:=filepath.Join(outside,"important.txt")
	if err:=os.WriteFile(file,[]byte("keep"),0600);err!=nil {t.Fatal(err)}
	past:=time.Now().Add(-60*24*time.Hour)
	if err:=os.Chtimes(file,past,past);err!=nil {t.Fatal(err)}
	base:=filepath.Join(local,"NVIDIA")
	if err:=os.MkdirAll(base,0700);err!=nil {t.Fatal(err)}
	if err:=os.Symlink(outside,filepath.Join(base,"DXCache"));err!=nil {t.Skip("symlinks not allowed on this host")}
	r:=cleanTargets(fullTargetsForLocal(local,nil),time.Now())
	if r.Deleted!=0 {t.Fatalf("followed a linked cache directory: %+v",r)}
	if _,err:=os.Stat(file);err!=nil {t.Fatal(err)}
}

func TestNoBrowserCredentialsOrHistoryInTargets(t *testing.T) {
	local:=filepath.Join(t.TempDir(),"AppData","Local")
	root:=filepath.Join(local,"Google","Chrome","User Data","Default")
	if err:=os.MkdirAll(root,0700);err!=nil {t.Fatal(err)}
	targets:=fullTargetsForLocal(local,nil)
	for _,target:=range targets {
		base:=strings.ToLower(filepath.Base(target.Path))
		for _,name:=range []string{"cookies","history","login data","local storage","sessions"} {
			if base==name {t.Fatalf("sensitive target: %s",target.Path)}
		}
	}
}
