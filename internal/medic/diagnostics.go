package medic

import (
	"fmt"
	"net/http"
	"strings"
)

func runDiagnostics(client *http.Client, cfg Config, health *healthResp) []check {
	diagnostics := []check{}

	// 0) Version/latency/HA markers
	if health != nil {
		if health.Version != "" {
			v := health.Version
			if health.Enterprise {
				v += " (ent)"
			}
			diagnostics = append(diagnostics, check{"Vault version", true, v})
		}
		if health.EchoDurationMS != nil {
			diagnostics = append(diagnostics, check{"Health latency", true, fmt.Sprintf("%dms", *health.EchoDurationMS)})
		}
		if health.HAConnHealthy != nil && (health.Standby != nil && *health.Standby) {
			diagnostics = append(diagnostics, check{"HA link healthy", *health.HAConnHealthy, fmt.Sprintf("%v", *health.HAConnHealthy)})
		}
		if health.RemovedFromCL != nil && *health.RemovedFromCL {
			diagnostics = append(diagnostics, check{"Removed from cluster", false, "true"})
		}
		if health.ReplicationDR != "" && health.ReplicationDR != "disabled" {
			diagnostics = append(diagnostics, check{"DR mode", true, health.ReplicationDR})
		}
		if health.ReplicationPerf != "" && health.ReplicationPerf != "disabled" {
			diagnostics = append(diagnostics, check{"Performance mode", true, health.ReplicationPerf})
		}
		if health.ReplicationDRLegacy != nil && health.ReplicationDRLegacy.Mode != "" {
			diagnostics = append(diagnostics, check{"DR mode", true, health.ReplicationDRLegacy.Mode})
		}
		if health.ReplicationPerfLegacy != nil && health.ReplicationPerfLegacy.Mode != "" {
			diagnostics = append(diagnostics, check{"Performance mode", true, health.ReplicationPerfLegacy.Mode})
		}
	}

	// 1) Leader info
	type leaderResp struct {
		HAEnabled bool   `json:"ha_enabled"`
		IsSelf    *bool  `json:"is_self,omitempty"`
		Leader    string `json:"leader_address"`
	}
	var lr leaderResp
	if code, err := doGET(client, cfg, "/v1/sys/leader", &lr); err == nil && code == 200 {
		addr := strings.TrimSpace(lr.Leader)
		if addr == "" {
			addr = cfg.Addr
		}
		diagnostics = append(diagnostics, check{"Leader address", true, addr})

		isSelf := false
		if lr.IsSelf != nil {
			isSelf = *lr.IsSelf
		}
		if health != nil && health.Initialized && !health.Sealed && (health.Standby == nil || !*health.Standby) {
			isSelf = true
		} else if sameAddress(lr.Leader, cfg.Addr) || strings.TrimSpace(lr.Leader) == "" {
			isSelf = true
		}
		diagnostics = append(diagnostics, check{"Leader is self", true, fmt.Sprintf("%v", isSelf)})
	} else if code == 403 {
		diagnostics = append(diagnostics, check{"Leader info", true, "forbidden (insufficient perms)"})
	}

	// 2) Seal status
	type sealStatus struct {
		Type      string `json:"type"`
		Threshold int    `json:"t"`
		N         int    `json:"n"`
		Progress  int    `json:"progress"`
	}
	var ss sealStatus
	if code, err := doGET(client, cfg, "/v1/sys/seal-status", &ss); err == nil && code == 200 {
		jsonSealType = ss.Type
		autoUnseal := ss.Threshold == 0 && ss.N == 0
		if autoUnseal {
			jsonSealThresh = ""
			jsonSealProg = nil
			diagnostics = append(diagnostics, check{"Seal type", true, ss.Type})
		} else {
			jsonSealThresh = fmt.Sprintf("%d/%d", ss.Threshold, ss.N)
			jsonSealProg = &ss.Progress
			diagnostics = append(diagnostics, check{"Seal type", true, fmt.Sprintf("%s (threshold %s, progress %d)", ss.Type, jsonSealThresh, ss.Progress)})
		}
	}

	// 3) Secret engines + KV flavors
	type mounts struct {
		Data map[string]struct {
			Type    string         `json:"type"`
			Options map[string]any `json:"options"`
		} `json:"data"`
	}
	var m mounts
	if code, err := doGET(client, cfg, "/v1/sys/mounts", &m); err == nil && (code == 200 || code == 204) {
		total := 0
		kvTotal := 0
		kvV2 := 0
		for path, mount := range m.Data {
			if path == "" {
				continue
			}
			total++
			if mount.Type == "kv" || mount.Type == "generic" {
				kvTotal++
				if mount.Options != nil {
					if verRaw, ok := mount.Options["version"]; ok {
						if fmt.Sprintf("%v", verRaw) == "2" {
							kvV2++
						}
					}
				}
			}
		}
		diagnostics = append(diagnostics, check{"Secret engines", true, fmt.Sprintf("%d", total)})
		kvV1 := kvTotal - kvV2
		diagnostics = append(diagnostics, check{"KV engines", true, fmt.Sprintf("total=%d (v2=%d, v1=%d)", kvTotal, kvV2, kvV1)})
	} else if code == 403 {
		diagnostics = append(diagnostics, check{"Secret engines", true, "forbidden (insufficient perms)"})
	}

	// 4) Auth methods
	type auths struct {
		Data map[string]struct {
			Type string `json:"type"`
		} `json:"data"`
	}
	var a auths
	if code, err := doGET(client, cfg, "/v1/sys/auth", &a); err == nil && code == 200 {
		cnt := 0
		for p := range a.Data {
			if p != "" {
				cnt++
			}
		}
		diagnostics = append(diagnostics, check{"Auth methods", true, fmt.Sprintf("%d", cnt)})
	} else if code == 403 {
		diagnostics = append(diagnostics, check{"Auth methods", true, "forbidden (insufficient perms)"})
	}

	// 5) Token introspection
	type tokenSelf struct {
		Data struct {
			Policies  []string `json:"policies"`
			TTL       int64    `json:"ttl"`
			Renewable bool     `json:"renewable"`
			Orphan    bool     `json:"orphan"`
		} `json:"data"`
	}
	var ts tokenSelf
	if code, err := doGET(client, cfg, "/v1/auth/token/lookup-self", &ts); err == nil && code == 200 {
		diagnostics = append(diagnostics, check{"Token policies", true, strings.Join(ts.Data.Policies, ",")})

		ttlStr := humanTTL(ts.Data.TTL)
		if ts.Data.TTL <= 0 {
			jsonTokenTTL = "infinite"
		} else {
			jsonTokenTTL = ttlStr
		}
		jsonTokenRen = &ts.Data.Renewable
		jsonTokenOrph = &ts.Data.Orphan

		if ts.Data.TTL <= 0 {
			diagnostics = append(diagnostics, check{"Token TTL", true,
				fmt.Sprintf("%s (renewable=%v, orphan=%v) — non-expiring", ttlStr, ts.Data.Renewable, ts.Data.Orphan)})
		} else {
			diagnostics = append(diagnostics, check{"Token TTL", true,
				fmt.Sprintf("%s (renewable=%v, orphan=%v)", ttlStr, ts.Data.Renewable, ts.Data.Orphan)})
		}
	} else if code == 403 {
		diagnostics = append(diagnostics, check{"Token policies", true, "forbidden (insufficient perms)"})
	}

	return diagnostics
}

// JSON accumulators filled during diagnostics
var (
	jsonSealType   string
	jsonSealThresh string
	jsonSealProg   *int

	jsonTokenTTL  string
	jsonTokenRen  *bool
	jsonTokenOrph *bool
)
