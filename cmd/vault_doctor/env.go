package main

import (
	"bufio"
	"os"
	"strings"
)

// Very small .env loader, no external deps.
// Supports KEY=VALUE lines, ignores comments and blanks. No export, no quotes parsing beyond trimming.
func loadDotEnvIfPresent(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// naive split on first '='
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		k := strings.TrimSpace(parts[0])
		v := strings.TrimSpace(parts[1])
		// strip surrounding quotes if any
		v = strings.Trim(v, `"'`)
		if k != "" {
			// only set if not already in environment
			if _, exists := os.LookupEnv(k); !exists {
				_ = os.Setenv(k, v)
			}
		}
	}
}
