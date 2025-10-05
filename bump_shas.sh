#!/usr/bin/env bash
set -euo pipefail

VER="${1:?usage: $0 0.1.2}"
FORM="homebrew-tap/Formula/vault_doctor.rb"

declare -A FILES=(
  [darwin_arm64]="dist/vault_doctor_${VER}_darwin_arm64"
  [darwin_amd64]="dist/vault_doctor_${VER}_darwin_amd64"
  [linux_arm64]="dist/vault_doctor_${VER}_linux_arm64"
  [linux_amd64]="dist/vault_doctor_${VER}_linux_amd64"
)

declare -A SHAS
for k in "${!FILES[@]}"; do
  SHAS[$k]=$(shasum -a 256 "${FILES[$k]}" | awk '{print $1}')
done

# in-place replace
sed -i '' \
  -e "s#\(vault_doctor_${VER}_darwin_arm64\"\)\s*sha256 \".*\"#\1\n      sha256 \"${SHAS[darwin_arm64]}\"#" \
  -e "s#\(vault_doctor_${VER}_darwin_amd64\"\)\s*sha256 \".*\"#\1\n      sha256 \"${SHAS[darwin_amd64]}\"#" \
  -e "s#\(vault_doctor_${VER}_linux_arm64\"\)\s*sha256 \".*\"#\1\n      sha256 \"${SHAS[linux_arm64]}\"#" \
  -e "s#\(vault_doctor_${VER}_linux_amd64\"\)\s*sha256 \".*\"#\1\n      sha256 \"${SHAS[linux_amd64]}\"#" \
  "$FORM"

echo "Updated SHAs in $FORM:"
printf "  darwin_arm64 %s\n  darwin_amd64 %s\n  linux_arm64 %s\n  linux_amd64 %s\n" \
  "${SHAS[darwin_arm64]}" "${SHAS[darwin_amd64]}" "${SHAS[linux_arm64]}" "${SHAS[linux_amd64]}"
