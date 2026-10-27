package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CleanTarget is an explicitly allowlisted, user-owned cache directory.
// Nothing outside this list is considered for removal.
type CleanTarget struct {
	Name      string
	Path      string
	OlderThan time.Duration
}

type CleanResult struct {
	Deleted   int
	Bytes     int64
	Skipped   int
	Scanned   int
	Locations int
}

// targetsForLocal requires the standard LocalAppData suffix so a broken or
// tampered environment variable cannot turn a broad folder into a target.
func targetsForLocal(local string) []CleanTarget {
	if local == "" || !filepath.IsAbs(local) {
		return nil
	}
	local = filepath.Clean(local)
	if !strings.EqualFold(filepath.Base(local), "Local") ||
		!strings.EqualFold(filepath.Base(filepath.Dir(local)), "AppData") {
		return nil
	}
	return []CleanTarget{
		{"Προσωρινά αρχεία χρήστη", filepath.Join(local, "Temp"), 14 * 24 * time.Hour},
		{"Μνήμη DirectX", filepath.Join(local, "D3DSCache"), 14 * 24 * time.Hour},
		{"Παλιά αρχεία σφαλμάτων", filepath.Join(local, "CrashDumps"), 30 * 24 * time.Hour},
		{"Αρχειοθετημένες αναφορές Windows", filepath.Join(local, "Microsoft", "Windows", "WER", "ReportArchive"), 30 * 24 * time.Hour},
	}
}

func approvedRoot(root string) bool {
	// Reject reparse points not just at the root but in every directory
	// under LocalAppData, to prevent junctions from escaping the allowlist.
	cursor := root
	for {
		info, err := os.Lstat(cursor)
		if err != nil || !info.IsDir() || isLinkOrReparse(info) {
			return false
		}
		parent := filepath.Dir(cursor)
		if parent == cursor {
			break
		}
		cursor = parent
	}
	return true
}

// cleanTargets never follows symbolic links or Windows directory junctions.
// It removes aged regular files only; all directories and recent files remain.
func cleanTargets(targets []CleanTarget, now time.Time) CleanResult {
	var result CleanResult
	const maxFiles = 150000
	const maxDepth = 24

	for _, target := range targets {
		if target.OlderThan < 24*time.Hour || !approvedRoot(target.Path) {
			continue
		}
		result.Locations++
		cutoff := now.Add(-target.OlderThan)
		var visit func(string, int)
		visit = func(directory string, depth int) {
			if depth > maxDepth || result.Scanned >= maxFiles {
				result.Skipped++
				return
			}
			entries, err := os.ReadDir(directory)
			if err != nil {
				result.Skipped++
				return
			}
			for _, entry := range entries {
				if result.Scanned >= maxFiles {
					result.Skipped++
					break
				}
				path := filepath.Join(directory, entry.Name())
				info, err := os.Lstat(path)
				if err != nil {
					result.Skipped++
					continue
				}
				if isLinkOrReparse(info) {
					result.Skipped++
					continue
				}
				if info.IsDir() {
					visit(path, depth+1)
					continue
				}
				result.Scanned++
				if !info.Mode().IsRegular() || !info.ModTime().Before(cutoff) {
					continue
				}
				// The file may have changed since enumeration: check again
				// immediately before removal to avoid deleting newer files.
				fresh, err := os.Lstat(path)
				if err != nil || isLinkOrReparse(fresh) || !fresh.Mode().IsRegular() ||
					!fresh.ModTime().Before(cutoff) || fresh.Size() != info.Size() {
					result.Skipped++
					continue
				}
				if err := os.Remove(path); err != nil {
					result.Skipped++ // Busy or protected files are left intact.
					continue
				}
				result.Deleted++
				result.Bytes += fresh.Size()
			}
		}
		visit(target.Path, 0)
	}
	return result
}

func formatBytes(n int64) string {
	if n < 1024 {
		return fmt.Sprintf("%d B", n)
	}
	if n < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(n)/1024)
	}
	if n < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(n)/(1024*1024))
	}
	return fmt.Sprintf("%.2f GB", float64(n)/(1024*1024*1024))
}
