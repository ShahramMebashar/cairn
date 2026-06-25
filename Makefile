BIN := bin/cairn

.PHONY: build test fmt vet check run dev up web web-dev web-build tidy clean desktop desktop-dev desktop-up sidecar

build: ## Build the cairn binary into bin/
	go build -o $(BIN) ./cmd/cairn

test: ## Run all tests
	go test ./...

fmt: ## Format all code
	gofmt -w cmd internal

vet: ## Run go vet
	go vet ./...

check: fmt vet test ## Format, vet, and test

run: build ## Run the MCP server (ACTOR=agent:claude-1 REPO=.)
	$(BIN) serve --actor $(or $(ACTOR),agent:claude-1) --repo $(or $(REPO),.)

dev: ## Live-reload the MCP server with air (rebuilds on save)
	@command -v air >/dev/null || go install github.com/air-verse/air@latest
	air

up: build ## Run backend + web dev server together (REPO=. ADDR=:8080)
	@echo "backend on $(or $(ADDR),:8080) + vite dev — Ctrl-C stops both"
	@trap 'kill 0' EXIT INT TERM; \
		$(BIN) web --repo $(or $(REPO),.) --addr $(or $(ADDR),:8080) & \
		(cd web && pnpm dev) & \
		wait

web: build ## Run the web/HTTP server (REPO=. ADDR=:8080)
	$(BIN) web --repo $(or $(REPO),.) --addr $(or $(ADDR),:8080)

web-dev: ## Run the Vite dev server (proxies /api to :8080 — run `make web` too)
	cd web && pnpm dev

web-build: ## Build the web UI (web/dist)
	cd web && pnpm build

desktop-dev: ## Run the Tauri desktop app in dev (run `make web` in another shell)
	cd desktop && pnpm tauri:dev

desktop-up: build ## Run everything: Go server + Vite + desktop window (REPO=. ADDR=:8080)
	@echo "backend on $(or $(ADDR),:8080) + tauri dev (starts vite) — Ctrl-C stops all"
	@trap 'kill 0' EXIT INT TERM; \
		$(BIN) web --repo $(or $(REPO),.) --addr $(or $(ADDR),:8080) & \
		(cd desktop && pnpm tauri:dev) & \
		wait

desktop: ## Build the Tauri desktop installer for this OS (embeds UI + sidecar)
	cd web && pnpm install
	cd desktop && pnpm install && pnpm tauri:build

sidecar: ## Build the cairn binary as a Tauri sidecar (desktop/src-tauri/binaries)
	node scripts/build-sidecar.mjs

tidy: ## Tidy go.mod/go.sum
	go mod tidy

clean: ## Remove build artifacts
	rm -rf bin

help: ## List targets
	@grep -hE '^[a-z-]+:.*##' $(MAKEFILE_LIST) | sed 's/:.*## /\t/'
