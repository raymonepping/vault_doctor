package medic

import (
	"encoding/json"
	"net/http"
	"strings"
)

type healthResp struct {
	Initialized   bool   `json:"initialized"`
	Sealed        bool   `json:"sealed"`
	Standby       *bool  `json:"standby,omitempty"`
	PerfStandby   *bool  `json:"performance_standby,omitempty"`
	ClusterName   string `json:"cluster_name"`
	ServerTimeUTC int64  `json:"server_time_utc"`
	Version       string `json:"version,omitempty"`
	Enterprise    bool   `json:"enterprise,omitempty"`

	ReplicationDR   string `json:"replication_dr_mode,omitempty"`
	ReplicationPerf string `json:"replication_performance_mode,omitempty"`

	HAConnHealthy  *bool  `json:"ha_connection_healthy,omitempty"`
	RemovedFromCL  *bool  `json:"removed_from_cluster,omitempty"`
	ClockSkewMS    *int64 `json:"clock_skew_ms,omitempty"`
	EchoDurationMS *int64 `json:"echo_duration_ms,omitempty"`

	// legacy nested
	ReplicationDRLegacy *struct {
		Mode string `json:"mode"`
	} `json:"replication_dr,omitempty"`
	ReplicationPerfLegacy *struct {
		Mode string `json:"mode"`
	} `json:"replication_performance,omitempty"`
}

func vaultHealth(client *http.Client, cfg Config) (*healthResp, int, error) {
	url := strings.TrimRight(cfg.Addr, "/") + "/v1/sys/health"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, err
	}
	withVaultHeaders(req, cfg)

	res, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	status := res.StatusCode
	var out healthResp
	_ = json.NewDecoder(res.Body).Decode(&out)
	return &out, status, nil
}

func healthMode(code int) string {
	switch code {
	case 200:
		return "active"
	case 429:
		return "standby"
	case 472:
		return "dr-secondary"
	case 473:
		return "perf-standby"
	case 474:
		return "standby (ha-unhealthy)"
	case 501:
		return "not-initialized"
	case 503:
		return "sealed"
	case 530:
		return "removed"
	default:
		return "unknown"
	}
}

func collectHints(h *healthResp, status int) []string {
	hints := []string{}
	switch status {
	case 501:
		hints = append(hints, "Vault not initialized. Run 'vault operator init' or use automation to initialize.")
	case 503:
		hints = append(hints, "Node is sealed. Unseal it, or ensure auto-unseal is configured.")
	case 429:
		hints = append(hints, "Standby node. Route clients/LB to the active leader for writes.")
	case 472:
		hints = append(hints, "DR secondary detected. This node will not serve writes.")
	case 473:
		hints = append(hints, "Performance standby detected. Reads OK; route writes to active.")
	case 474:
		hints = append(hints, "Standby cannot reach active (HA unhealthy). Check cluster connectivity.")
	case 530:
		hints = append(hints, "Node removed from HA cluster. Rejoin or point clients to another member.")
	}
	if h != nil {
		if h.Standby != nil && *h.Standby {
			hints = append(hints, "You are hitting a standby node.")
		}
		if h.HAConnHealthy != nil && !*h.HAConnHealthy {
			hints = append(hints, "HA connection is not healthy; check leader/LB/network.")
		}
		if h.PerfStandby != nil && *h.PerfStandby {
			hints = append(hints, "Performance standby: consider routing reads appropriately, writes to active.")
		}
		if h.RemovedFromCL != nil && *h.RemovedFromCL {
			hints = append(hints, "This node reports 'removed_from_cluster=true'.")
		}
	}
	return hints
}
