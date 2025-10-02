package medic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func approleLogin(client *http.Client, cfg Config) (string, error) {
	type req struct {
		RoleID   string `json:"role_id"`
		SecretID string `json:"secret_id"`
	}
	type auth struct {
		ClientToken string `json:"client_token"`
	}
	type resp struct {
		Auth *auth `json:"auth"`
	}

	url := strings.TrimRight(cfg.Addr, "/") + "/v1/auth/approle/login"
	body, _ := json.Marshal(req{RoleID: cfg.RoleID, SecretID: cfg.SecretID})
	httpReq := must(NewRequestJSON(http.MethodPost, url, body))
	withVaultHeaders(httpReq, cfg)

	res, err := client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return "", fmt.Errorf("approle login failed: HTTP %d", res.StatusCode)
	}
	var out resp
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.Auth == nil || out.Auth.ClientToken == "" {
		return "", fmt.Errorf("approle login response missing client_token")
	}
	return out.Auth.ClientToken, nil
}
