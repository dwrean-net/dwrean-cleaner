//go:build windows

package main

import (
	"strings"
	"syscall"
	"unsafe"
)

// On enumeration failure assume each supported cache-owning app is open.
// Fail closed: never touch the cache of an application we cannot verify is stopped.
func runningCleanupProcesses() map[string]bool {
	names := []string{"chrome", "msedge", "brave", "firefox", "discord"}
	open := make(map[string]bool, len(names))
	snapshot, err := syscall.CreateToolhelp32Snapshot(syscall.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		for _, name := range names { open[name] = true }
		return open
	}
	defer syscall.CloseHandle(snapshot)
	var p syscall.ProcessEntry32
	p.Size = uint32(unsafe.Sizeof(p))
	if err := syscall.Process32First(snapshot, &p); err != nil {
		for _, name := range names { open[name] = true }
		return open
	}
	for {
		process := strings.ToLower(syscall.UTF16ToString(p.ExeFile[:]))
		for _, name := range names {
			if process == name+".exe" { open[name] = true }
		}
		if err := syscall.Process32Next(snapshot, &p); err != nil { break }
	}
	return open
}
