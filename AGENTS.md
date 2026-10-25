# AGENTS.md

This file provides guidance to AI coding agents (Claude Code, Codex, Cursor, etc.) when working with code in this repository.

## Agent Setup

- **Claude Code**: if `.claude/settings.local.json` does not exist, create it with the content below before doing anything else, then continue with your task. It disables AI attribution in commits and PRs (see Pull Requests below).

```json
{
  "attribution": {
    "commit": "",
    "pr": "",
    "sessionUrl": false
  }
}
```

- This file is personal and must never be committed - `.claude/` is covered by `.gitignore`.
- Regardless of settings, never add AI attribution to commits or PR descriptions - see the Pull Requests section.

## Development Commands

- Go targets are in the `Makefile` (`make test`, `make build/ios`)
- The iOS app builds only on macOS with Xcode; CI compiles it for the simulator in `.github/workflows/ios.yml`

## Project Architecture

### Backend Architecture (Go)
The backend follows a layered architecture with clear separation of concerns:

- **`mobile/`**: the only entry point - gomobile binds it for the iOS app. It runs the app without a network listener and serves API requests through `Engine.Serve` (`make build/ios`)
- **`ios/`**: native SwiftUI app; `EngineClient` calls the API in-process through `Engine.Serve` and reads event streams the same way. The Xcode project is generated with XcodeGen from `ios/project.yml` - see `ios/README.md`

- **`internal/`**: Core application logic organized by domain
  - **Domain layer** (`internal/domain/`): Core business entities. It holds no repository or service interfaces - each consumer declares a small unexported interface listing only the methods it calls, and constructors return concrete `*Service` structs
  - **Database layer** (`internal/database/`): Repository implementations and database logic
  - **Service layer**: Business logic services (e.g., `internal/release/`, `internal/filter/`)
  - **HTTP layer** (`internal/http/`): REST API handlers and routing
  - **Infrastructure**: External integrations (`internal/indexer/`, `internal/irc/`, `internal/notification/`)

- **`rls/`**: the release-name parser, autobrr/rls v0.10.0 with two fixes marked `bard fix`, copied from new_parser (keep the fixes when updating it). `internal/domain/rlsvocab.csv` adds new_parser's extra tags and groups on top of it. Settings → Vocabulary in the app adds entries and removes listed ones at runtime (`internal/vocab`, stored in `release_vocab`), rebuilding the parser on each change

- **`pkg/`**: Reusable packages and client libraries
  - Contains API clients for some trackers (`btn`, `ggn`, `ops`, `red`) and the Torznab feed client
  - Utility packages for different indexer APIs

### iOS App (SwiftUI)
- Deployment target iOS 16.4 - avoid APIs introduced later
- Screens are grouped by tab in `ios/Autobrr/` (`Dashboard`, `Inbox`, `Filters`, `Releases`, `Logs`, `Settings`); shared engine plumbing lives in `ios/Autobrr/Engine/`
- Engine objects that screens edit field by field are kept as `JSONValue` documents (`JSONDocument`), so fields a screen does not show survive a save; simple read-only data uses `Decodable` structs with `JSON.decoder`

### Key Domain Concepts
- **Releases**: Torrent releases that get processed through filters
- **Filters**: Rules that determine which releases should be downloaded
- **Actions**: What to do with matched releases (watch folder, exec, webhook)
- **Indexers**: Torrent trackers, defined as YAML in `internal/indexer/definitions/`
- **IRC**: Real-time monitoring of indexer announce channels
- **Feed**: RSS/Torznab feed processing for indexers without IRC

### Database Support
- **SQLite** only
- Database migrations handled automatically
- Repository pattern for data access

### Service Wiring
- `internal/application` owns the lifecycle: `setup` constructs repos and services in dependency order, `components()` lists what starts (in order) and stops (in reverse) when `Run`'s context is cancelled
- A new background service is a `component` in that list, with a `stop` for anything it opens; mark it `optional` when its failure should be logged while the rest of the app keeps running
- Services talk across packages through the typed event bus in `internal/events/` when a direct dependency would create an import cycle

## Code Style

### Go
- Format with `go fmt ./...` before committing
- Group imports: stdlib, internal, third-party; keep alphabetical within groups
- Use the `pkg/errors` helpers for wrapping and sentinel errors; return wrapped errors and avoid panics outside `main`
- Handle errors explicitly with early returns; don't swallow errors
- Log with zerolog through the injected `zerolog.Logger`: values go in typed fields and the message stays a constant - `log.Error().Err(err).Int("filter_id", id).Msg("could not find filter")`, not `Msgf` with values inline
- Functions receiving context should take `context.Context` as the first parameter and respect cancellation
- Exported identifiers need doc comments; keep names descriptive and consistent
- Keep domain DTOs and JSON tags synced; prefer tagged struct fields over `map[string]any`

### Swift
- Match the surrounding SwiftUI style: small views, `@StateObject` models for screen state, `async`/`await` calls through `EngineClient`

### Comments
Applies to Go and Swift alike. Excessive low-value comments are the most common defect in AI-generated code - default to writing **no comment**, and only add one when it earns its place:

- Comment the *why*, never the *what*: non-obvious invariants, workarounds (link the issue or upstream bug), tracker/IRC protocol quirks, deliberate deviations from the expected approach
- If a comment restates what the code plainly says, it must be deleted. Banned patterns: step narration (`// loop over filters`, `// return the result`), section banners (`// error handling`), and restating the function name above the function
- Never describe your edit in a comment (`// changed to use X`, `// new helper for Y`) - that history belongs in the commit message, not the code
- No commented-out code and no unprompted `TODO`/`FIXME` markers
- Godoc comments on exported identifiers: one concise sentence starting with the identifier name; expand only when the API is genuinely subtle
- Some older code contains narration comments (`// get filters`) - do not take them as license to add more, and feel free to drop them in code you're already touching

## Testing
- `make test` runs the untagged Go suite. Database and other integration tests carry `//go:build integration` next to the code they test and run with `go test -tags=integration ./...` against an in-memory SQLite database, so no Docker is needed
- IRC integration tests in `test/irc/`, run with `go test -tags=irc_integration_test ./test/irc/...` (in-process ircd, no Docker) - see `test/irc/README.md`
- Build-tagged packages are invisible to a plain `go list ./...`; pass the tag to tooling that discovers packages first
- Mock indexer server available in `test/mockindexer/`

## Development Notes
- The Go minimum is the `go` line in `go.mod`
- The app never opens a network listener: every API call and event stream goes through `Engine.Serve` in-process
- Real-time updates (logs, inbox) are Server-Sent Events streamed over the same bridge

## Pull Requests
- PRs target the `develop` branch
- Use the PR template at `.github/pull_request_template.md` - fill out the relevant sections and delete those that don't apply
- PR titles follow [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/#summary) (e.g. `feat(indexers): add NewTracker`, `fix(irc): reconnect on timeout`) - commits are squashed on merge, so the PR title becomes the commit message
- Indexer definitions live as YAML files in `internal/indexer/definitions/` - adding a new indexer is the most common contribution, and usually only touches a single definition file there. Removing one requires a tombstone in `internal/indexer/definitions/deprecated/` (CI rejects a removal without it) - see `CONTRIBUTING.md`
- Database schema changes require a SQLite migration (`internal/database/migrations/`)
- The template has an **AI disclosure** section - always answer it truthfully, stating what was AI-generated and which model/tool was used
- Do **not** add AI attribution to commit messages or PR descriptions - no `Co-Authored-By: Claude`, `Generated with ...` trailers, or similar. AI usage belongs in the PR template's AI disclosure section, not in the git history
- See `CONTRIBUTING.md` for the full contribution guide
