package medic

type Config struct {
	Addr       string
	Token      string
	RoleID     string
	SecretID   string
	Namespace  string
	SkipVerify bool
}

type check struct {
	name   string
	ok     bool
	detail string
}

// For JSON mode
type jsonCheck struct {
	Name   string `json:"name"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail,omitempty"`
}
type jsonDiag = jsonCheck

type jsonResult struct {
	Version        string `json:"version"`
	Timestamp      int64  `json:"timestamp"`
	Mode           string `json:"mode,omitempty"`
	HTTPStatus     int    `json:"http_status,omitempty"`
	ClusterName    string `json:"cluster_name,omitempty"`
	LeaderAddress  string `json:"leader_address,omitempty"`
	LeaderIsSelf   *bool  `json:"leader_is_self,omitempty"`
	SealType       string `json:"seal_type,omitempty"`
	SealThreshold  string `json:"seal_threshold,omitempty"`
	SealProgress   *int   `json:"seal_progress,omitempty"`
	TokenTTL       string `json:"token_ttl,omitempty"`
	TokenRenewable *bool  `json:"token_renewable,omitempty"`
	TokenOrphan    *bool  `json:"token_orphan,omitempty"`
	// (we keep KV counts inside diagnostics; promote later if desired)
	Checks      []jsonCheck `json:"checks"`
	Diagnostics []jsonDiag  `json:"diagnostics,omitempty"`
	Hints       []string    `json:"hints,omitempty"`
	Failures    int         `json:"failures"`
}

// CLI options passed from main
type Options struct {
	Version string
	Quiet   bool
	JSON    bool
	NoColor bool
}
