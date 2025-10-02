package medic

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

func mustJSONEncoder() *json.Encoder {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc
}

func Run(opt Options) int {
	printBanner(opt.Version, opt)

	// env
	loadDotEnvIfPresent(".env")
	cfg := LoadConfigFromEnv()

	results := []check{}

	// VAULT_ADDR present
	if cfg.Addr == "" {
		results = append(results, check{"VAULT_ADDR present", false, "not set"})
		return finish(results, 0, nil, 0, cfg, nil, opt)
	}
	results = append(results, check{"VAULT_ADDR present", true, cfg.Addr})

	client := NewHTTPClient(cfg.SkipVerify)

	// Auth: token or AppRole
	if cfg.Token == "" && cfg.RoleID != "" && cfg.SecretID != "" {
		token, err := approleLogin(client, cfg)
		if err != nil {
			results = append(results, check{"AppRole login", false, err.Error()})
			return finish(results, 0, nil, 0, cfg, nil, opt)
		}
		cfg.Token = token
		results = append(results, check{"AppRole login", true, "received client token"})
	} else if cfg.Token != "" {
		results = append(results, check{"VAULT_TOKEN present", true, "token provided"})
	} else {
		results = append(results, check{"Auth configuration", false, "provide VAULT_TOKEN or VAULT_ROLE_ID + VAULT_SECRET_ID"})
		return finish(results, 0, nil, 0, cfg, nil, opt)
	}

	// Health
	health, status, err := vaultHealth(client, cfg)
	if err != nil {
		results = append(results, check{"API reachability", false, fmt.Sprintf("%v", err)})
		return finish(results, status, health, status, cfg, nil, opt)
	}
	results = append(results, check{"API reachability", true, fmt.Sprintf("%s (HTTP %d)", healthMode(status), status)})

	if health != nil {
		results = append(results, check{"Initialized", health.Initialized, fmt.Sprintf("%v", health.Initialized)})
		results = append(results, check{"Sealed", !health.Sealed, fmt.Sprintf("sealed=%v", health.Sealed)})
		if health.Standby != nil {
			results = append(results, check{"Standby mode", !*health.Standby, fmt.Sprintf("standby=%v", *health.Standby)})
		}
		if health.ClusterName != "" {
			results = append(results, check{"Cluster name", true, health.ClusterName})
		}
		if health.ServerTimeUTC != 0 {
			results = append(results, check{"Server time", true, fmt.Sprintf("%d", health.ServerTimeUTC)})
		}
	} else {
		results = append(results, check{"Health payload", false, "no JSON body returned"})
	}

	// Optionally prompt to unseal
	if health != nil && health.Sealed && !opt.JSON && !opt.Quiet {
		if err := promptUnseal(client, cfg, opt); err != nil && !opt.Quiet && !opt.JSON {
			fmt.Printf("%sUnseal attempt failed: %v%s\n", cwrap("", colRed, opt), err, colReset)
		}
		time.Sleep(500 * time.Millisecond)
		newHealth, newStatus, err := vaultHealth(client, cfg)
		if err == nil && newHealth != nil && !newHealth.Sealed {
			diags := runDiagnostics(client, cfg, newHealth)
			return finish(results, newStatus, newHealth, newStatus, cfg, diags, opt)
		}
	}

	// Normal finish with diagnostics (if unsealed)
	var diags []check
	if health != nil && !health.Sealed {
		diags = runDiagnostics(client, cfg, health)
	}
	return finish(results, status, health, status, cfg, diags, opt)
}
