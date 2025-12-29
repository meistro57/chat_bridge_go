# AGENT GUIDE

This repository builds **Chat Bridge (Go Edition)**: a Cobra-based CLI that connects two AI providers in a streaming conversation using the retro terminal UI from the original Python project.

## Quick-start & build commands
- `make build`: compiles `./...` into `bin/chat-bridge`, embedding `internal/version.Version` via `-ldflags` (`git describe --tags` fallback). Expects `GOBIN`/`GOPATH` under `$HOME` and exports them during the build.
- `make build-all`: cross-compiles linux/amd64, darwin/amd64, darwin/arm64, and windows/amd64 outputs into `bin/`.
- `make run`: builds and executes `bin/chat-bridge` in one step.
- `make demo`: builds and runs `chat-bridge start --provider-a openai --provider-b openai --max-rounds 3` (requires `OPENAI_API_KEY`).
- `make test`: runs `go test -v ./...` so every package must compile.
- `make test-coverage`: runs tests with `-cover` and writes `coverage.out` before opening a browser via `go tool cover -html=coverage.out`.
- `make install`: copies the built binary into `$GOBIN` after building.
- `make clean`: removes `bin/` and `coverage.out`.
- `make fmt`: executes `go fmt ./...`.
- `make deps`: runs `go get` for lipgloss/cobra/godotenv and `go mod tidy` so dependencies match go.mod.

## Environment & configuration hints
- Copy `.env.example` to `.env` or otherwise export environment variables before running commands that contact AI providers.
- Required API keys: `OPENAI_API_KEY`, `ANTHROPIC_API_KEY`, `GEMINI_API_KEY`, `DEEPSEEK_API_KEY`, `OPENROUTER_API_KEY`. `pkg/config.Config.Validate` requires at least one of these.
- `pkg/config.Load` uses `github.com/joho/godotenv` so `.env` is loaded automatically but missing `.env` is tolerated.
- Optional overrides exist for base URLs (e.g., `OPENAI_BASE_URL`, `OLLAMA_HOST`, `LMSTUDIO_BASE_URL`) and default models per provider. `BRIDGE_PROVIDER_A`/`BRIDGE_PROVIDER_B` define the CLI’s default pair when `--provider-*` flags are not supplied.
- `pkg/config.Config` exposes helpers (`GetAPIKey`, `GetDefaultModel`, `GetProviderBaseURL`) so the CLI can route each provider-specific configuration into `providers.ProviderConfig` when instantiating a provider.

## Core layout
- `main.go`: short entry point that calls `cmd.Execute()`.
- `cmd/`: Cobra commands. `root.go` shows the retro banner/help and handles `--version` using `internal/version`. `start.go` defines `chat-bridge start` with flags for providers, models, temperatures, starter prompt, and round limits, orchestrates configuration loading, and runs the multi-round streaming loop while using `providers.NewProvider` and the provider registry so any registered factory becomes available to the CLI without editing a switch statement.
- `pkg/config/`: environment-first configuration with defaults, validation, and helpers (`GetAPIKey`, `GetDefaultModel`, `GetProviderBaseURL`, `getEnvOrDefault`).
- `pkg/providers/`: defines the `Provider` interface, shared errors, provider registry management (`RegisterProvider`, `RegisterProviderFactory`, `GetProviderSpec`, `ListProviders`, `NewProvider`), and the OpenAI implementation (registers spec + factory in `init()`, hits `/models` for health checks, streams SSE chunks via `StreamChat`, and exposes a base URL override through `ProviderConfig`).
- `pkg/ui/`: lipgloss-based styling helpers (`PrintBanner`, `PrintSectionHeader`, `PrintError`, `PrintInfo`, `PrintSuccess`, `Colorize`, rainbow text utilities) plus the retro color constants used throughout the command and logging.
- `internal/version/`: version metadata (default `1.0.0`, `dev`, `unknown`) that gets overridden via `-ldflags` during builds.
- `pkg/conversation/`: currently empty but reserved space for shared conversation helpers or history/state management.

## Patterns & conventions
- Cobra is the CLI framework; `cmd/root.go` and `cmd/start.go` register commands/flags in `init()` functions, and `cmd/start.go` centralizes conversation orchestration (prompt history, streaming, agent switching, colored output).
- Providers implement the `Provider` interface (name, list of models, streaming API, health check, default model) and register their metadata via `RegisterProvider` plus a factory via `RegisterProviderFactory`. The CLI calls `providers.NewProvider`, which uses the registry map, so no new switch statement is required when adding providers.
- Each provider gets instantiated with `ProviderConfig` that carries the API key, optional base URL (`cfg.GetProviderBaseURL`), model, and temperature. `cmd/start.go` tracks per-agent temperatures (`tempA`/`tempB`) and ensures the currently speaking agent’s temperature is passed to every `StreamChat` request.
- Streaming is handled by `Provider.StreamChat`, which returns `<-chan string` and `<-chan error`. `cmd/start.go` selects on text, errors, and a 30-second timeout per chunk, accumulates the response in a `strings.Builder`, and only adds the assistant message to history once the stream closes.
- UI helpers keep the CLI output consistent: colored agent names, success/error/warning/info methods, section headers, and the banner are all centralized in `pkg/ui/colors.go`.
- Configuration values are read from env/`.env` and are not stored globally beyond `cmd/start.go` and `pkg/config`. Passing them explicitly keeps the CLI thread-safe and simplifies future concurrency.

## Testing & formatting
- Tests are Go-native; run `make test` or `go test -v ./...`. The suite now includes `pkg/config/config_test.go` and `pkg/providers/provider_test.go`, so add new tests next to the package they cover.
- Use `make fmt` or `go fmt ./...` before committing changes; the style is idiomatic Go with tabs for indentation.
- Running `make test-coverage` generates `coverage.out` if line coverage details are needed.

## Observed gotchas
- Only OpenAI has an implementation and factory registered by default, so other providers must implement the interface and register both a spec and a factory (`RegisterProvider` + `RegisterProviderFactory`) before `providers.NewProvider` can instantiate them.
- `make demo` assumes `OPENAI_API_KEY` is set; the CLI aborts fast if the OpenAI key is missing because the OpenAI health check hits `/models` to validate credentials.
- Health checks call `/models` for each provider before starting the conversation; missing or invalid keys surface immediately.
- Providers stream SSE data line by line and skip malformed chunks, so new providers should mirror that retry-tolerant loop and close the channels appropriately when the context is done or the stream ends.

## Containerized workflow
- A multi-stage `Dockerfile` lives at the repository root. It downloads modules, builds the binary for `linux/amd64` with `CGO_ENABLED=0`, and copies the static executable into a Scratch-based final image with `ENTRYPOINT ["/chat-bridge"]`.
- Keep a `.dockerignore` next to it for faster builds (`bin`, `.git`, `.github`, coverage artifacts, etc.).
- Build once with `docker build -t chat-bridge-go .` and run with `docker run --rm -it -v "$PWD/.env:/app/.env" -e OPENAI_API_KEY=sk-... chat-bridge-go start`. Mounting the `.env` file into `/app` ensures the CLI sees it via `godotenv` and environment overrides stay outside the image.
- The container uses `/app` as its working directory, so mount directories or files relative to that path (e.g., conversation logs) when you need persistence.

## Useful references
- `README.md` for onboarding instructions, features, and the roadmap.
- `Makefile` for the supported command set, dependency commands, and how `GOBIN`/`GOPATH` are managed during builds/tests.
- `pkg/config/config.go` for `Config` structure, validation, and helpers (`GetProviderBaseURL`, `GetDefaultModel`, etc.).
- `pkg/providers/provider.go` for the shared interface, factory registry (`RegisterProviderFactory`, `NewProvider`), and common errors.
- `pkg/providers/openai.go` for an example provider that registers spec + factory and streams from OpenAI.
- `pkg/ui/colors.go` for consistent retro theming and helper printing functions.
- `pkg/config/config_test.go` and `pkg/providers/provider_test.go` for examples of how to structure tests covering configuration loading and provider registration.

Agents touching this repo should scope their changes with these conventions in mind, keep streaming/state logic centralized in `cmd/start.go`, and reuse the helper packages rather than reinventing cross-cutting concerns.