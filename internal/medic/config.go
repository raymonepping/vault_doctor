package medic

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

func LoadConfigFromEnv() Config {
	return Config{
		Addr:       strings.TrimSpace(os.Getenv("VAULT_ADDR")),
		Token:      strings.TrimSpace(os.Getenv("VAULT_TOKEN")),
		RoleID:     strings.TrimSpace(os.Getenv("VAULT_ROLE_ID")),
		SecretID:   strings.TrimSpace(os.Getenv("VAULT_SECRET_ID")),
		Namespace:  strings.TrimSpace(os.Getenv("VAULT_NAMESPACE")),
		SkipVerify: strings.EqualFold(strings.TrimSpace(os.Getenv("VAULT_SKIP_VERIFY")), "true"),
	}
}

func loadDotEnvIfPresent(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		kv := strings.SplitN(line, "=", 2)
		if len(kv) != 2 {
			continue
		}
		k := strings.TrimSpace(kv[0])
		v := strings.TrimSpace(kv[1])
		if os.Getenv(k) == "" {
			_ = os.Setenv(k, v)
		}
	}
}

func NewHTTPClient(skipVerify bool) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
	}
	return &http.Client{Transport: tr, Timeout: 10 * time.Second}
}

func NewRequestJSON(method, url string, body []byte) (*http.Request, error) {
	req, err := http.NewRequest(method, url, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func withVaultHeaders(req *http.Request, cfg Config) {
	if cfg.Token != "" {
		req.Header.Set("X-Vault-Token", cfg.Token)
	}
	if cfg.Namespace != "" {
		req.Header.Set("X-Vault-Namespace", cfg.Namespace)
	}
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(fmt.Errorf("unexpected: %w", err))
	}
	return v
}
