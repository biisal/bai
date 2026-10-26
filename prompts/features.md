# bai web UI: feature parity

Follow-up to `impl.md` and `ui-improvement.md`. Bring the web UI up to the TUI's feature set and past the basics of T3 Code, DeepSeek Harness and OpenCode web. Add a "Features" section to `task.md` and work from it. Obey the same rules as `impl.md` (no new Go dependency, stdlib first, token/Host/Origin guard on every route, never serialize config structs, one active run, fewest files, one runnable check per non-trivial piece). New Go code goes in `internal/web`; new UI follows the `ui-improvement.md` look.

Lift the `impl.md` "Later" and "Do not" limits only for the items below. Everything else there still holds.

## Features

### 1. `@file` mentions
- `GET /api/files?q=<query>`: reuse `internal/files` (`IfFileFinding` and the finder the TUI `commands/files.go` uses), scoped to the working directory, honoring the same ignore rules. Return at most 20 relative paths, `{path, isDir}`. Reject `..` and absolute paths; never read file contents here.
- Prompt input: typing `@` opens a `command`/popover list under the caret, arrow keys + Enter/Tab to pick, Esc to close, debounced fetch (about 150 ms). The picked path is inserted the same way `ReplaceFileQuery` does, so the model receives the same text the TUI sends. Mentions render as a small chip in the sent user message.

### 2. Slash commands
- Typing `/` at the start of the input opens a menu: `/new`, `/models`, `/sessions`, `/themes` mapped to UI actions (new chat, model picker, focus the sidebar search, theme picker). No server route for these.

### 3. Sessions
- Sidebar search (client-side filter over the loaded list).
- `DELETE /api/conversations/{id}` and `PATCH /api/conversations/{id}` `{title}` (rename). Refuse with 409 while a run is active on that conversation. Confirm delete with a `dialog`.
- Group the list by Today / Yesterday / Earlier.

### 4. Settings dialog
A gear in the sidebar footer opens one `dialog` with tabs. Hand-built response structs only; secrets never leave the server.
- **Models:** providers and models (read-only list from config: id, context, max output), active model, switch. API keys shown only as `configured: true|false`. No key editing in v1.
- **Appearance:** theme from `internal/config/themes.go` (same names the TUI offers), applied as CSS variables, plus light/dark/system. Persist in `localStorage`.
- **Skills and plugins:** `GET /api/skills` and `GET /api/plugins` list name, description and source path from the skills loader and plugins manifest, read-only.
- **Git:** shadow-repo status (`initialized`, `dirty`) from `internal/git`, read-only. If not initialized, show a one-line hint and the command to run.
- **About:** version, working directory, config path, log path, update availability if `internal/update` exposes it without side effects.

### 5. Changes view
- `GET /api/conversations/{id}/diff`: run read-only `git` against the shadow repo (`RunGitCommandContext`, fixed args, no user-supplied arguments) and return per-file unified diffs for the last run, capped at 1 MB.
- UI: a "Changes" button on the header opens a right panel (`Sheet` on small screens) with a file list and a diff view. Use a tiny hand-rolled unified-diff renderer (green/red line tint, no library) unless AI Elements has a fitting component that stays within the bundle budget.

### 6. Run feedback
- Show per-run token usage and elapsed time under the last assistant message if Fantasy exposes usage in the finish callback; emit it as a transient `data-usage` part. If it does not, skip it and say so.
- Copy button on messages and on code blocks; regenerate-last-response button that re-sends the last user message (new run, never edits history).
- Sound/desktop notification when a run finishes in a background tab, using the browser `Notification` API behind an explicit opt-in toggle in Settings. Do not reuse `internal/notifier`.

### 7. Shortcuts
- Cmd/Ctrl+K: command palette (new chat, switch session, switch model, open settings, open changes). Cmd/Ctrl+B sidebar (already there). `?` opens a shortcut cheat sheet. Esc stops a running response only when the input is focused and empty.

## Constraints

- New routes all pass through the existing guard, return JSON, and use `http.MaxBytesReader`. `DELETE`/`PATCH` also require `Content-Type: application/json` where there is a body and an `Origin` match.
- `/api/files` and `/api/*/diff` are the only filesystem-adjacent routes allowed; both are read-only, cwd-scoped, size-capped, and tested with a path-traversal case.
- Do not touch the TUI, DB schema or migrations. If delete/rename needs a new sqlc query, stop and ask first.
- Do not add settings that write config, provider/key editing, approvals, worktrees, terminal, tunnels or multi-project switching.
- Initial JS grows at most about 40 KB gzipped; diff and settings panels are lazy loaded.

## Acceptance

- `go test -race ./...` passes apart from failures recorded at the start. New tests: `/api/files` (match, ignore rules, traversal rejected, cap), delete/rename (409 while running, not found), diff (fixed args only, size cap), settings structs contain no key or base URL (marshal and grep the JSON).
- `pnpm lint`, `pnpm format:check`, `tsc --noEmit` pass.
- `curl` with and without token on each new route: 401 without, 403 on wrong Host/Origin.
- In the browser: `@` picker inserts a path that the model reads; slash menu works; search, rename and delete a session; change theme; open Changes after a run that edited a file; Cmd/Ctrl+K palette; notification opt-in. Describe what you saw.

## Report

Per feature done/skipped/deviated, new routes with their guard result, bundle size before/after, anything the TUI can do that the web UI still cannot.
