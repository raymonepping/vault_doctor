# -------- Config --------
APP := $(shell basename $(CURDIR))

# Resolve module path robustly (fallback to canonical)
MODULE := $(shell go list -m 2>/dev/null)
ifeq ($(strip $(MODULE)),)
MODULE := github.com/raymonepping/vault_doctor
endif

# Version: default from latest tag (vX.Y.Z). Override with VERSION=...
VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo dev)
STRIPPED_VERSION := $(shell printf "%s" "$(VERSION)" | sed 's/^v//')

# Commit/Date with safe fallbacks
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE   := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

GOFLAGS := -trimpath

# Inject BOTH internal/version.Version and main.buildVersion (belt & suspenders)
LDFLAGS := -s -w \
  -X $(MODULE)/internal/version.Version=$(VERSION) \
  -X $(MODULE)/internal/version.Commit=$(COMMIT) \
  -X $(MODULE)/internal/version.Date=$(DATE) \
  -X main.buildVersion=$(VERSION)

# Cross targets we support
TARGETS := darwin/amd64 darwin/arm64 linux/amd64 linux/arm64

# Homebrew tap setup
TAP_FORMULA := homebrew-tap/Formula/vault_doctor.rb
GH_OWNER    := raymonepping
REPO_NAME   := vault_doctor
TARBALL_URL := https://github.com/$(GH_OWNER)/$(REPO_NAME)/archive/refs/tags/$(VERSION).tar.gz
TARBALL     := $(CURDIR)/v$(VERSION).tar.gz

# -------- Phony targets --------
.PHONY: all build dev fmt vet lint test clean release publish completions install-completions check-tag check-clean print-version gorelease brew-bump brew-push check-bins

all: build

print-version:
	@echo "APP=$(APP)"
	@echo "MODULE=$(MODULE)"
	@echo "VERSION=$(VERSION)"
	@echo "STRIPPED_VERSION=$(STRIPPED_VERSION)"
	@echo "COMMIT=$(COMMIT)"
	@echo "DATE=$(DATE)"
	@echo "LDFLAGS=$(LDFLAGS)"

fmt:
	@go fmt ./...

vet:
	@go vet ./...

build: fmt vet
	@mkdir -p bin
	@echo ">> go build -ldflags '$(LDFLAGS)'"
	@go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o bin/$(APP) ./cmd/$(APP)

dev:
	@echo ">> go run -ldflags '$(LDFLAGS)'"
	@go run -ldflags "$(LDFLAGS)" ./cmd/$(APP)

test:
	@go test ./...

clean:
	@rm -rf bin dist

# Ensure working tree is clean before tagging/releasing
check-clean:
	@if ! git diff --quiet || ! git diff --cached --quiet; then \
		echo "⚠ Working tree is dirty. Commit or stash changes before releasing."; \
		exit 1; \
	fi

# Ensure the VERSION exists as a tag (if it looks like a tag)
check-tag:
	@if printf "%s" "$(VERSION)" | grep -Eq '^v[0-9]'; then \
		git rev-parse -q --verify "refs/tags/$(VERSION)" >/dev/null || { \
			echo "❌ Tag $(VERSION) not found. Create and push it: git tag -a $(VERSION) -m '$(APP) $(VERSION)'; git push origin $(VERSION)"; \
			exit 1; \
		}; \
	fi

# Cross-platform release builds (local dist/ artifacts)
release: clean fmt vet
	@mkdir -p dist
	@set -e; \
	for t in $(TARGETS); do \
		OS=$${t%/*}; ARCH=$${t##*/}; \
		OUT="dist/$(APP)_$(VERSION)_$${OS}_$${ARCH}"; \
		echo "Building $$OUT"; \
		GOOS=$$OS GOARCH=$$ARCH go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o "$$OUT" ./cmd/$(APP); \
	done
	@echo "Built release artifacts in ./dist:"; ls -1 dist

# Publish to GitHub Releases (requires: tag exists + gh CLI)
publish: check-tag release
	@echo "Uploading dist binaries to GitHub release $(VERSION)…"
	@gh release view "$(VERSION)" >/dev/null 2>&1 || gh release create "$(VERSION)" --title "$(APP) $(VERSION)" --notes "Release $(APP) $(VERSION)"
	@gh release upload "$(VERSION)" dist/$(APP)_$(VERSION)_* --clobber
	@echo "✅ Published $(VERSION) to GitHub Releases."

# Optional: GoReleaser full pipeline
gorelease: check-clean check-tag
	@goreleaser release --clean

# Generate completion scripts into dist/
completions: build
	@mkdir -p dist
	@./bin/$(APP) completion bash > dist/$(APP).bash || true
	@./bin/$(APP) completion zsh  > dist/_$(APP)     || true
	@./bin/$(APP) completion fish > dist/$(APP).fish || true
	@echo "Wrote completions to dist/:"
	@ls -1 dist | grep -E '(^_$(APP)|($(APP)\.(bash|fish)))' || true

install-completions: completions
	# bash
	@mkdir -p $(HOME)/.local/share/bash-completion/completions
	@cp dist/$(APP).bash $(HOME)/.local/share/bash-completion/completions/$(APP) || true
	# zsh
	@mkdir -p $(HOME)/.zsh/completions
	@cp dist/_$(APP) $(HOME)/.zsh/completions/_$(APP) || true
	# fish
	@mkdir -p $(HOME)/.config/fish/completions
	@cp dist/$(APP).fish $(HOME)/.config/fish/completions/$(APP).fish || true
	@echo "Installed completions. Restart your shell (or re-source completion directories)."

# ---- Brew / Tap automation ----
brew-bump: check-tag
	@echo "▶ Downloading $(TARBALL_URL)"
	@curl -sSL -o $(TARBALL) "$(TARBALL_URL)"
	@echo "▶ Calculating sha256"
	@SHA=$$(shasum -a 256 $(TARBALL) | awk '{print $$1}'); \
	echo "sha256=$$SHA"; \
	echo "▶ Writing formula $(TAP_FORMULA)"; \
	printf "%s\n" \
"class VaultDoctor < Formula" \
"  desc \"Medic for HashiCorp Vault: health, caps, KV, transit\"" \
"  homepage \"https://github.com/$(GH_OWNER)/$(REPO_NAME)\"" \
"  version \"$(STRIPPED_VERSION)\"" \
"  url \"https://github.com/$(GH_OWNER)/$(REPO_NAME)/archive/refs/tags/$(VERSION).tar.gz\"" \
"  sha256 \"$$SHA\"" \
"  license \"MPL-2.0\"" \
"" \
"  depends_on \"go\" => :build" \
"" \
"  def install" \
"    mod = Utils.safe_popen_read(\"go\", \"list\", \"-m\").chomp" \
"    ldflags = [" \
"      \"-s -w\"," \
"      \"-X \#{mod}/internal/version.Version=v\#{version}\"," \
"      \"-X main.buildVersion=v\#{version}\"," \
"    ].join(\" \")" \
"    ohai \"Module: \#{mod}\"" \
"    ohai \"ldflags: \#{ldflags}\"" \
"    system \"go\", \"build\", \"-trimpath\", \"-ldflags\", ldflags, \"-o\", bin/\"vault_doctor\", \"./cmd/vault_doctor\"" \
"  end" \
"" \
"  test do" \
"    assert_match version.to_s, shell_output(\"#{bin}/vault_doctor -V\")" \
"  end" \
"end" \
	> $(TAP_FORMULA)
	@echo "✅ Brew formula updated for $(VERSION)"

brew-push:
	@cd homebrew-tap && git add Formula/vault_doctor.rb
	@cd homebrew-tap && git commit -m "vault_doctor: bump to $(VERSION)" || echo "No changes to commit."
	@cd homebrew-tap && git push origin main
	@echo "🚀 Tap update pushed to raymonepping/homebrew-tap"

check-bins:
	@echo "Local bin:"; ./bin/$(APP) -V || true; ./bin/$(APP) medic | sed -n '1p' || true; echo
	@echo "Homebrew bin:"; /opt/homebrew/bin/$(APP) -V || true; /opt/homebrew/bin/$(APP) medic | sed -n '1p' || true
