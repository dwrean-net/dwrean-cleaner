//go:build !windows

package main

import "os"

func isLinkOrReparse(info os.FileInfo) bool {
	return info.Mode()&os.ModeSymlink != 0
}
