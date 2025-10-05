package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/raymonepping/vault_doctor/internal/medic"
	"github.com/raymonepping/vault_doctor/internal/version"
)

// Fallback injected by ldflags if internal/version.Version fails to bind
var buildVersion = "" // injected via: -X main.buildVersion=...

func resolvedVersion() string {
	v := strings.TrimSpace(version.Version)
	if v == "" || v == "v" || v == "dev" {
		bv := strings.TrimSpace(buildVersion)
		if bv != "" && bv != "v" {
			v = bv
		}
	}
	if v == "" || v == "v" {
		v = "dev"
	}
	return v
}

func main() {
	if len(os.Args) == 1 {
		fmt.Print(medic.Doc(resolvedVersion()))
		return
	}

	switch os.Args[1] {
	case "-h", "--help", "help":
		fmt.Print(medic.Doc(resolvedVersion()))
		return

	case "-V", "--version", "--version=true", "--version=false":
		// Always print a non-empty semver (no leading 'v')
		fmt.Println(strings.TrimPrefix(resolvedVersion(), "v"))
		return

	case "--version-debug":
		// Debug: show raw vars and the resolved outcome (helps diagnose ldflags issues)
		fmt.Println("version.Version =", quote(version.Version))
		fmt.Println("main.buildVersion =", quote(buildVersion))
		fmt.Println("resolved =", quote(resolvedVersion()), "-> printed =", quote(strings.TrimPrefix(resolvedVersion(), "v")))
		return

	case "medic":
		runMedicCmd()
		return

	case "completion":
		runCompletionCmd()
		return

	default:
		fmt.Print(medic.Doc(resolvedVersion()))
		return
	}
}

func quote(s string) string {
	if s == "" {
		return `""`
	}
	return `"` + s + `"`
}

func runMedicCmd() {
	fs := flag.NewFlagSet("medic", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "Output JSON")
	quiet := fs.Bool("quiet", false, "Quiet mode")
	noColor := fs.Bool("no-color", false, "Disable colors")
	_ = fs.Parse(os.Args[2:])

	opt := medic.Options{
		Version: resolvedVersion(),
		Quiet:   *quiet,
		JSON:    *jsonOut,
		NoColor: *noColor,
	}

	code := medic.Run(opt)
	os.Exit(code)
}

func runCompletionCmd() {
	args := os.Args[2:]
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: vault_doctor completion [bash|zsh|fish]")
		os.Exit(1)
	}
	os.Exit(medic.PrintCompletion(strings.ToLower(args[0])))
}
