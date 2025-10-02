package medic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"
)

const (
	colReset  = "\033[0m"
	colRed    = "\033[31m"
	colGreen  = "\033[32m"
	colYellow = "\033[33m"
)

func cwrap(s, color string, opt Options) string {
	if opt.NoColor || os.Getenv("NO_COLOR") != "" {
		return s
	}
	return color + s + colReset
}

func termWidth() int {
	if w, _, err := term.GetSize(int(syscall.Stdout)); err == nil && w > 0 {
		return w
	}
	if c := os.Getenv("COLUMNS"); c != "" {
		if n, err := strconv.Atoi(c); err == nil && n > 0 {
			return n
		}
	}
	return 80
}

func nameColWidth(results []check) int {
	maxName := 0
	for _, r := range results {
		if l := len(r.name); l > maxName {
			maxName = l
		}
	}
	w := termWidth()
	width := maxName
	if width < 22 {
		width = 22
	}
	if width > 40 {
		width = 40
	}
	if width+2 > w-24 {
		width = w - 26
		if width < 18 {
			width = 18
		}
	}
	return width
}

func sameAddress(a, b string) bool {
	ax := strings.TrimRight(strings.ToLower(strings.TrimSpace(a)), "/")
	bx := strings.TrimRight(strings.ToLower(strings.TrimSpace(b)), "/")
	return ax != "" && bx != "" && ax == bx
}

func humanTTL(sec int64) string {
	if sec <= 0 {
		return "∞"
	}
	d := time.Duration(sec) * time.Second
	if d%time.Hour == 0 {
		return fmt.Sprintf("%dh", int(d/time.Hour))
	}
	if d%time.Minute == 0 && d >= time.Minute {
		return fmt.Sprintf("%dm", int(d/time.Minute))
	}
	return fmt.Sprintf("%ds", int(sec))
}

// Generic GET JSON helper with headers
func doGET(client *http.Client, cfg Config, path string, out any) (int, error) {
	url := strings.TrimRight(cfg.Addr, "/") + path
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	withVaultHeaders(req, cfg)
	res, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	code := res.StatusCode
	if out != nil && code >= 200 && code <= 299 {
		if err := json.NewDecoder(res.Body).Decode(out); err != nil {
			return code, err
		}
	}
	return code, nil
}
