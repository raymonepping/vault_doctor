package medic

import "fmt"

// Doc returns the CLI usage & env docs.
func Doc(version string) string {
	return fmt.Sprintf(`vault_doctor

Usage:
  vault_doctor medic [--json] [--quiet] [--no-color]
  vault_doctor -V|--version
  vault_doctor -h|--help

Version:
  %s

Flags (medic):
  --json       Output machine-readable JSON (no banner, no prompts).
  --quiet      Suppress pretty output and prompts (exit code reflects status).
  --no-color   Disable ANSI colors (NO_COLOR=1 also works).

Environment variables (read directly and via .env if present):
  VAULT_ADDR         https://<host>:8200
  VAULT_TOKEN        <token>
  VAULT_ROLE_ID      <role_id>
  VAULT_SECRET_ID    <secret_id>
  VAULT_NAMESPACE    <namespace>
  VAULT_SKIP_VERIFY  true|false
`, version)
}
