package main

import (
	"runtime/debug"
	"strings"
)

func buildinfo() (revision string, modified, ok bool) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "", false, false
	}

	settings := make(map[string]string)
	for _, s := range info.Settings {
		settings[s.Key] = s.Value
	}

	if rev, ok := settings["vcs.revision"]; ok {
		return rev, settings["vcs.modified"] == "true", true
	}

	// info.Main.Version can be something like: v0.1.6-0.20231208225832-9ba5a2aace9a
	if idx := strings.LastIndexByte(info.Main.Version, '-'); idx > -1 {
		return info.Main.Version[idx+1:], false, true
	}

	return "<no version>", false, false
}
