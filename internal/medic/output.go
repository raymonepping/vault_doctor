package medic

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func normVersion(v string) string {
	v = strings.TrimSpace(v)
	// Treat empty, "dev", or bare "v" (bad inject) as dev
	if v == "" || v == "dev" || v == "v" {
		return "dev"
	}
	// Always return a single leading "v" for banner display
	return "v" + strings.TrimPrefix(v, "v")
}

func printBanner(version string, opt Options) {
	if opt.Quiet || opt.JSON {
		return
	}
	fmt.Printf("%s %s  %s  %s\n",
		cwrap("🩺 vault_doctor", colGreen, opt),
		cwrap("medic", colYellow, opt),
		cwrap("", colReset, opt),
		normVersion(version),
	)
}

func summaryLine(failures int) string {
	if failures == 0 {
		return "Medic finished: all checks passed ✔"
	}
	return fmt.Sprintf("Medic finished: %d check(s) failed ❌", failures)
}

func printResultsPretty(results []check, status int, trailer string, opt Options) {
	if opt.Quiet || opt.JSON {
		return
	}
	if status != 0 {
		mode := healthMode(status)
		fmt.Printf("%s %s (HTTP %d)\n", cwrap("ℹ Mode", colYellow, opt), mode, status)
	}

	nameW := nameColWidth(results)
	for _, r := range results {
		mark := "✅"
		color := colGreen
		if !r.ok {
			mark = "❌"
			color = colRed
		}
		if opt.NoColor || os.Getenv("NO_COLOR") != "" {
			color = ""
		}
		name := r.name
		if len(name) < nameW {
			name = name + strings.Repeat(" ", nameW-len(name))
		}
		if r.detail != "" {
			if color != "" {
				fmt.Printf("%s%s%s %s  %s\n", color, mark, colReset, name, r.detail)
			} else {
				fmt.Printf("%s %s  %s\n", mark, name, r.detail)
			}
		} else {
			if color != "" {
				fmt.Printf("%s%s%s %s\n", color, mark, colReset, name)
			} else {
				fmt.Printf("%s %s\n", mark, name)
			}
		}
	}
	if trailer != "" {
		fmt.Println()
		if strings.Contains(trailer, "failed") {
			fmt.Println(cwrap(trailer, colRed, opt))
		} else {
			fmt.Println(cwrap(trailer, colGreen, opt))
		}
	}
}

func printDiagnostics(diags []check, opt Options) {
	if opt.Quiet || opt.JSON || len(diags) == 0 {
		return
	}
	fmt.Println()
	fmt.Printf("%s%s%s\n", cwrap("Diagnostics", colYellow, opt), "", "")
	nameW := nameColWidth(diags)
	for _, d := range diags {
		name := d.name
		if len(name) < nameW {
			name = name + strings.Repeat(" ", nameW-len(name))
		}
		fmt.Printf("%s %s  %s\n", cwrap("•", colGreen, opt), name, d.detail)
	}
}

// JSON/Quiet finisher
func finish(results []check, status int, health *healthResp, httpStatus int, cfg Config, diags []check, opt Options) int {
	failures := 0
	for _, r := range results {
		if !r.ok {
			failures++
		}
	}
	hints := collectHints(health, status)

	if opt.JSON {
		out := jsonResult{
			Version:        opt.Version,
			Timestamp:      time.Now().Unix(),
			Mode:           healthMode(status),
			HTTPStatus:     httpStatus,
			ClusterName:    health.ClusterName,
			LeaderAddress:  "", // filled in diagnostics
			LeaderIsSelf:   nil,
			SealType:       jsonSealType,
			SealThreshold:  jsonSealThresh,
			SealProgress:   jsonSealProg,
			TokenTTL:       jsonTokenTTL,
			TokenRenewable: jsonTokenRen,
			TokenOrphan:    jsonTokenOrph,
			Checks:         make([]jsonCheck, 0, len(results)),
			Hints:          hints,
			Failures:       failures,
		}
		for _, r := range results {
			out.Checks = append(out.Checks, jsonCheck{Name: r.name, OK: r.ok, Detail: r.detail})
		}
		if len(diags) > 0 {
			out.Diagnostics = make([]jsonDiag, 0, len(diags))
			for _, d := range diags {
				out.Diagnostics = append(out.Diagnostics, jsonDiag{Name: d.name, OK: d.ok, Detail: d.detail})
				// opportunistically lift leader_addr/self if present
				if d.name == "Leader address" && out.LeaderAddress == "" {
					out.LeaderAddress = d.detail
				}
				if d.name == "Leader is self" && out.LeaderIsSelf == nil {
					v := d.detail == "true"
					out.LeaderIsSelf = &v
				}
			}
		}
		enc := mustJSONEncoder()
		_ = enc.Encode(out)
	} else if opt.Quiet {
		if failures > 0 {
			fmt.Println("medic: checks failed")
		}
	} else {
		printResultsPretty(results, status, summaryLine(failures), opt)
		if len(hints) > 0 {
			fmt.Println()
			fmt.Printf("%sNext actions%s\n", cwrap("", colYellow, opt), colReset)
			for _, h := range hints {
				fmt.Printf("  • %s\n", h)
			}
		}
		if len(diags) > 0 {
			printDiagnostics(diags, opt)
		}
	}

	if failures > 0 {
		return 1
	}
	return 0
}
