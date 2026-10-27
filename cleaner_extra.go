package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// All paths are narrowly allowlisted, and active browser profiles are excluded.
// Only files older than the cutoff are removed by cleanTargets.
func extraTargets(local string, browsersRunning map[string]bool) []CleanTarget {
	if len(targetsForLocal(local)) == 0 { return nil }
	local = filepath.Clean(local)
	days := 24*time.Hour
	targets := []CleanTarget{
		{"Παλαιές αναφορές Windows σε αναμονή", filepath.Join(local,"Microsoft","Windows","WER","ReportQueue"), 45*days},
		{"Παλιά προσωρινά σφαλμάτων Windows", filepath.Join(local,"Microsoft","Windows","WER","Temp"), 14*days},
		{"NVIDIA DirectX shader cache", filepath.Join(local,"NVIDIA","DXCache"), 30*days},
		{"NVIDIA OpenGL shader cache", filepath.Join(local,"NVIDIA","GLCache"), 30*days},
		{"AMD DirectX shader cache", filepath.Join(local,"AMD","DxCache"), 30*days},
		{"AMD OpenGL shader cache", filepath.Join(local,"AMD","GLCache"), 30*days},
	}
	browsers := []struct{process, name, subpath string}{
		{"chrome", "Chrome", filepath.Join("Google", "Chrome", "User Data")},
		{"msedge", "Edge", filepath.Join("Microsoft", "Edge", "User Data")},
		{"brave", "Brave", filepath.Join("BraveSoftware", "Brave-Browser", "User Data")},
	}
	for _, b := range browsers {
		// If a browser is running, do not touch any of its disk caches.
		if browsersRunning[b.process] { continue }
		base := filepath.Join(local,b.subpath)
		if !approvedRoot(base) { continue }
		entries,err := os.ReadDir(base)
		if err != nil { continue }
		profiles := 0
		for _,entry := range entries {
			profile := entry.Name()
			if profile != "Default" && !isChromiumProfile(profile) { continue }
			if !entry.IsDir() || entry.Type()&os.ModeSymlink!=0 { continue }
			profileRoot := filepath.Join(base,profile)
			if !approvedRoot(profileRoot) { continue }
			profiles++
			if profiles>40 { break }
			for _,suffix := range []string{
				filepath.Join("Cache","Cache_Data"),
				filepath.Join("Code Cache","js"),
				filepath.Join("Code Cache","wasm"),
				"GPUCache", "GrShaderCache", "ShaderCache",
			} {
				targets=append(targets,CleanTarget{b.name+" cache",filepath.Join(profileRoot,suffix),30*days})
			}
		}
	}
	if !browsersRunning["firefox"] {
		base:=filepath.Join(local,"Mozilla","Firefox","Profiles")
		if approvedRoot(base) {
			entries,err:=os.ReadDir(base)
			if err==nil {
				count:=0
				for _,entry:=range entries {
					if !entry.IsDir() || entry.Type()&os.ModeSymlink!=0 {continue}
					if !isFirefoxProfile(entry.Name()) { continue }
					profile:=filepath.Join(base,entry.Name())
					if !approvedRoot(profile) { continue }
					count++
					if count>40 {break}
					targets=append(targets,CleanTarget{"Firefox cache",filepath.Join(profile,"cache2","entries"),30*days})
				}
			}
		}
	}
	return targets
}

func isChromiumProfile(s string) bool {
	if !strings.HasPrefix(s,"Profile ") {return false}
	suffix:=strings.TrimPrefix(s,"Profile ")
	if len(suffix)==0 || len(suffix)>4 {return false}
	for _,c:=range suffix {if c<'0' || c>'9' {return false}}
	return true
}

func isFirefoxProfile(s string) bool {
	dot:=strings.IndexByte(s,'.')
	if dot!=8 || len(s)>60 {return false}
	for _,c:=range s[:dot] {
		if !(c>='a'&&c<='z'||c>='A'&&c<='Z'||c>='0'&&c<='9') {return false}
	}
	return strings.HasPrefix(s[dot+1:],"default") || strings.HasPrefix(s[dot+1:],"dev-edition-default")
}

func fullTargetsForLocal(local string, browsersRunning map[string]bool) []CleanTarget {
	first:=targetsForLocal(local)
	return append(first, extraTargets(local,browsersRunning)...)
}
