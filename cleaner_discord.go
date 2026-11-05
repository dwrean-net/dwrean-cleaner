package main

import (
	"path/filepath"
	"strings"
	"time"
)

// Discord's official Windows cache lives in AppData/Roaming/discord/Cache.
// It is never scanned if Discord is running or a provided path is unexpected.
// This function intentionally excludes account settings, messages, credentials,
// Local Storage, and every other Discord directory.
func discordCacheTargets(local, roaming string, running map[string]bool) []CleanTarget {
	if running["discord"] || len(targetsForLocal(local)) == 0 || roaming == "" || !filepath.IsAbs(roaming) {
		return nil
	}
	local = filepath.Clean(local)
	roaming = filepath.Clean(roaming)
	if !strings.EqualFold(filepath.Base(roaming), "Roaming") ||
		!strings.EqualFold(filepath.Base(filepath.Dir(roaming)), "AppData") ||
		!strings.EqualFold(filepath.Dir(local), filepath.Dir(roaming)) {
		return nil
	}

	cache := filepath.Join(roaming, "discord", "Cache")
	if !approvedRoot(cache) {
		return nil
	}
	return []CleanTarget{{
		Name: "Παλιά cache Discord",
		Path: cache,
		OlderThan: 30 * 24 * time.Hour,
	}}
}

// Retains the old explicit AppData/Local allowlist and adds Discord's
// one verified cache root. The recycle bin is deliberately NOT in this list.
func fullTargetsForUser(local, roaming string, running map[string]bool) []CleanTarget {
	targets := fullTargetsForLocal(local, running)
	return append(targets, discordCacheTargets(local, roaming, running)...)
}
