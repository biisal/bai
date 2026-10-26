# bai web UI: implementation prompt for the coding agent

Oct 26, 2026 · @loki

Add a loopback-only web UI to bai as a single embedded binary: Go net/http speaks the Vercel AI SDK UI message stream, and a small Vite + React app built from AI Elements renders it.

## Mission and rules of engagement

You are working in `bai`, a Go terminal coding agent (Bubble Tea TUI, `charm.land/fantasy` agent loop, SQLite through sqlc and goose). Build `bai web`: a browser front end for the same agent, with no Node at runtime.

- Read `AGENTS.md` first and obey it: lazy senior dev mode, stdlib first, no new dependency if avoidable, fewest files, one runnable check for non-trivial logic.
- The user's working tree is dirty with uncommitted work: `.goreleaser.yml`, `internal/agent/gateway.go`, `internal/agent/core/tools/*`, `internal/consts/consts.go`, `internal/tui/chat-builder/*`. Never revert, stash, reformat or commit those. Work on a new branch and stage only files you created or deliberately edited.
- Run `go test -race ./...` before changing anything. Record failures that already exist and report them; do not fix them.
- Positioning: T3 Code is a Node WebSocket server plus Electron and mobile clients; DeepSeek Harness ships a Web UI with plugins. bai's edge is one Go binary, any model, terminal first. The web UI must not make bai heavier to install or run.
- Work milestone by milestone (see Milestones). Stop at each acceptance check and report before continuing.

### First action: create task.md

Before reading more code or writing any code, create `task.md` at the repo root. It is your live to-do list and the source for your final report.

- Write it as `- [ ]` checkboxes grouped under M0 to M4. Give each concrete task from Backend tasks and Frontend tasks its own line, and give each milestone's acceptance check its own line.
- Tick `- [x]` the moment an item is done, not in batches. Mark the item you are working on with `(doing)`, and keep only one item in that state.
- Add items you discover along the way. Add a short "Notes and blockers" section at the bottom with one line per decision, deviation from this prompt, or open question.
- Update it before starting each milestone and after each acceptance check, and record the check's result on its line.
- Leave it untracked and uncommitted. Do not let it grow past one line per item.

## What the repo already has

Verified by reading the code. Each finding creates a constraint you must respect.

| Finding | Consequence |
| --- | --- |
| `internal/pubsub` Broker is one buffered channel (32). `Subscribe()` returns that same channel and `Publish` blocks until there is room or ctx is done. | The TUI is its only reader. With no reader, the agent deadlocks after 32 events. A second reader would steal events. In web mode, drain it in one goroutine and add a test that publishes more than 64 events with no other subscriber. |
| Broker events are flat `{Type, Text, IsComplete}`. Tool events fire before execution, with no call id, no args, no result. | Not enough for AI SDK tool parts. Web must hook Fantasy's own stream callbacks. |
| `Gateway` holds one mutable `conversation` and one agent. `streamChat` reads `g.conversation.ID` live and `AddMessageToDB` lazily creates conversations. | Concurrent runs are unsafe. Never call `SetActiveConversation` while a run is active. Read history by explicit id with `GetMessagesByConversationID`. |
| `streamChat` builds `fantasy.AgentStreamCall` with OnToolCall, OnAgentStart, OnTextDelta(id, text), OnReasoningDelta, OnStepFinish and OnRetry. | charm.land/fantasy v0.45.2 very likely also has text and reasoning start/end, tool-input and tool-result callbacks. Verify with `go doc charm.land/fantasy AgentStreamCall` in your own shell. This was not verified. |
| Messages are saved to SQLite at step finish (plus partial text on error) as JSON `parts`, roles user, assistant, tool. | History must be converted to AI SDK UIMessages. Merge the assistant and tool messages of one run into one assistant message, with a `step-start` part between steps. |
| Conversations are scoped by `files.CurrentDir()`. | One server is one project, the cwd. No project switcher in v1. |
| Tools: `read`, `write_file`, `edit_file`, `bash`, plus dynamic tools from the plugins manifest. There is no approval flow anywhere. | The web UI exposes an LLM that can run bash. The security section is mandatory. Do not add approvals. |
| `config.ProviderConfig` has `APIKey` with a json tag. | Never serialize config structs to the browser. |
| `frontend/` is the marketing landing page (`bai-landing`): React 19, Vite 6, Tailwind 4, shadcn base-nova, several PNGs in `public/`. It is not embedded in the Go binary today. | Do not embed it and do not edit it. Build the app in a new `web/` folder so the binary only carries the app bundle. |
| `//go:embed` is used only for migrations and skill markdown. `.gitignore` ignores `dist`. | You must solve embedding on a clean checkout (see Backend tasks). |
| `utils/tests/assert.go` has test helpers. `make test` runs `go test -race -failfast ./...`. Frontend uses pnpm, oxlint and oxfmt. | Reuse them. No new test frameworks. |
| `make format` and `make lint-fix` run repo-wide gofmt and fixers. | Do not run them. They would rewrite the user's uncommitted files. |
| `.goreleaser.yml` only runs `go mod tidy` as a before-hook, and it has uncommitted edits. | Do not edit it. Report the exact hook lines needed. |
| `github.com/coder/websocket` appears only as an indirect dependency. | Not a reason to use WebSockets. |

## Architecture and decisions

One process, `bai web`, serves a REST API, one SSE chat stream and the embedded SPA on 127.0.0.1. Decisions are made; challenge one only with evidence from the code.

| Decision | Choice | Why |
| --- | --- | --- |
| Router | stdlib `net/http` ServeMux with Go 1.22+ method patterns | About nine routes. The one hard route is the SSE stream, which Huma would also handle through its StreamResponse escape hatch, and the AI SDK wire format cannot be described in OpenAPI. Huma would add dependencies to a go.mod that already lists about 100 indirect ones. If a public API appears later, Huma's stdlib adapter can be layered on top. |
| Transport | HTTP POST in, SSE out | `useChat` speaks SSE natively. DeepSeek-TUI's own `serve --http` also uses HTTP and SSE. WebSocket only pays off for multi-client bidirectional control, which v1 does not need. |
| Frontend | Vite + React 19 SPA, `useChat` from `@ai-sdk/react`, components from AI Elements, no Next.js | AI Elements docs list a Next.js project as a prerequisite, but the components are shadcn-registry source copied into your repo and target React 19 and Tailwind 4. Verify they build under Vite. If one imports `next/*`, patch or skip it. |
| Server-side AI SDK | None | Go emits the protocol. The `ai` package is used only in the browser. Never run `streamText` in a Node server. |
| Location | New `web/` folder with its own package.json, embedded by `web/embed.go` | Keeps landing-page images out of the binary that goreleaser ships and self-update downloads. |
| Concurrency | At most one active run per process. A second POST gets 409. | The agent edits one working tree and commits to one shadow git repo. Two runs corrupt each other. "Scalable" here means many viewers on one run, many stored sessions, and a later registry keyed by working directory. Parallel runs are a future feature through worktrees. |
| Run lifetime | A run owns a context derived from the server context, never the request context | A browser refresh must not kill the run. Stopping is an explicit endpoint. |
| Fan-out | Per run: in-memory buffer of encoded frames plus subscribers | No Redis. The AI SDK resume docs use it; a single local process does not need it. Cap the buffer (about 4 MB). A slow subscriber is dropped and may reconnect; the agent goroutine never waits on a browser. |
| Auth | Random per-launch token exchanged for an HttpOnly cookie, loopback bind | See Security requirements. |

### Scope from the competitor research

| Capability | Seen in | v1 |
| --- | --- | --- |
| Sessions sidebar, streaming markdown, reasoning, tool cards, stop | T3 Code (projects and threads), DeepSeek Harness Web UI | Yes |
| Model picker | T3 Code model selector | Yes |
| Survive page reload mid-run | Needed because runs outlive tabs | Yes, milestone 3 |
| Token auth and responsive layout | T3 Code `--auth-token`, mobile clients | Yes |
| Diff view over the shadow git history | DeepSeek Harness git views, T3 Code diff panel | Later |
| Plan, agent and yolo approval modes | DeepSeek-TUI | Later |
| `@file` mentions | Already in the TUI through `internal/files` | Later |
| Worktrees, PTY terminal, remote relay and tunnels | T3 Code | No |

## API and stream contract

The browser talks to these routes. Hand-write the TypeScript types for the JSON ones; there are too few to justify codegen.

| Route | Purpose |
| --- | --- |
| `GET /` | With `?t=<token>`: set the auth cookie, 302 to `/`. Otherwise serve the SPA. Extensionless unknown paths fall back to `index.html`; a missing `/assets/*` file is a real 404. |
| `GET /api/state` | Working directory base name, version, active provider and model, `running`: `{conversationId}` or null. |
| `GET /api/models` | Providers and model ids from config plus the active one. Never include keys or base URLs. |
| `PUT /api/model` | `{provider, model}`. Validate against config, then call `Gateway.AddOrUpdateProvider`. |
| `GET /api/conversations` | Conversations for the working directory, newest first: id, title, createdAt. |
| `GET /api/conversations/{id}/messages` | `UIMessage[]`. While a run is active on that conversation, stop at the run's user message so replay does not duplicate it. |
| `POST /api/chat` | Body `{id, message}` (the last UI message only). Starts a run and streams SSE. 409 if a run is active. |
| `GET /api/chat/{id}/stream` | Resume: 204 when no run is active, otherwise replay every frame of the run, then follow live. |
| `POST /api/chat/{id}/stop` | Cancel the run. 204. |

The chat `id` is the conversation id as a string. If it does not parse as an existing conversation in this working directory, create a new one and send a transient `data-conversation` part so the client learns the real id.

### SSE rules

- Headers: `Content-Type: text/event-stream`, `Cache-Control: no-cache, no-transform`, `x-vercel-ai-ui-message-stream: v1`, `X-Accel-Buffering: no`.
- Each frame is `data: <json>` followed by a blank line. Flush after every frame with `http.NewResponseController(w).Flush()`.
- Send `: ping` as a comment every 15 seconds. Terminate with `data: [DONE]`.
- Do not set `http.Server.WriteTimeout`. Set `ReadHeaderTimeout`.

### Frame order

| When | Frames |
| --- | --- |
| Run begins | `{"type":"start","messageId":"..."}`; for a new chat also `{"type":"data-conversation","data":{"id":42,"title":"..."},"transient":true}` |
| Each model step | `start-step` ... `finish-step` |
| Text | `text-start {id}`, `text-delta {id, delta}`, `text-end {id}` |
| Reasoning | `reasoning-start`, `reasoning-delta`, `reasoning-end`, same shape as text |
| Tool call | `tool-input-start {toolCallId, toolName}`, `tool-input-delta {toolCallId, inputTextDelta}`, `tool-input-available {toolCallId, toolName, input}`, then `tool-output-available {toolCallId, output}` or `tool-output-error {toolCallId, errorText}` |
| Provider retry | `reset-step`, so the failed attempt's partial output is discarded |
| Normal end | `finish`, then `[DONE]` |
| User stop | `{"type":"abort","reason":"user cancelled"}`, then `[DONE]` |
| Failure | `{"type":"error","errorText":"..."}`, then `[DONE]` |

The frame names come from the AI SDK stream-protocol page for the current docs. `tool-output-error` and the `transient` flag on data parts are from memory and were not on the page I read. Before writing the encoder, check every frame against the installed `ai` package (its chunk schema and type definitions in `node_modules`). Ignore older blog posts that show the `0:` and `9:` prefixed data-stream format: that is the obsolete v4 protocol.

### Encoding rules

- Build every frame with `json.Marshal` on a struct. Never concatenate JSON by hand.
- Embed a tool `input` as raw JSON only if `json.Valid`; otherwise send it as a string.
- bai tool results are text, so `output` is a string. A Fantasy error response maps to `tool-output-error`.
- Plugin tools have dynamic names. Emit them like any other tool; the client renders any part whose type starts with `tool-`.

### History shape

`GET .../messages` must produce what the stream would have produced, or reload looks different. Merge consecutive assistant and tool messages into one assistant UIMessage with parts `text`, `reasoning`, `tool-<name>` (with `toolCallId`, `state` of `output-available` or `output-error`, `input`, `output` or `errorText`) and `step-start` between steps. Use the DB id of the first message in a group, as a string, for the message id.

## Backend tasks

Target about six Go files in a new `internal/web` package plus `cmd/bai/web.go` and `web/embed.go`. Fewer is better.

1. **Gateway seam.** Web needs Fantasy's native stream events (text and reasoning ids, tool input, tool results, step boundaries, retries). Keep `StreamChat(ctx, message)` as it is for the TUI and add one variant that takes a decorator, for example `func(call *fantasy.AgentStreamCall)`, applied after the base callbacks are built. The web side must chain the existing callbacks, not replace them, because they save messages to SQLite and sync the shadow git repo. The TUI path must behave exactly as before and `internal/agent/gateway_test.go` must still pass. `gateway.go` has the user's uncommitted edits: read it fresh, make a minimal diff, do not reformat. If Fantasy v0.45.2 has no tool-input streaming callback, emit `tool-input-start` and `tool-input-available` together at `OnToolCall` and skip the deltas.
2. **Run registry** (`run.go`). One active run at a time; starting a second returns a busy error that the handler turns into 409. A run holds its conversation id, a cancel func, a capped slice of encoded frames and its subscribers. Appending a frame and registering a subscriber happen under one mutex so a late subscriber replays from index 0 with no gap and no duplicate. Delivery to a subscriber is a non-blocking send on a buffered channel; if it is full, drop that subscriber. The POST handler is just the first subscriber: when its client disconnects it unsubscribes and the run keeps going. When the run finishes, remove it from the registry; history in SQLite is then complete.
3. **Handlers** (`server.go`). The routes in the API section, JSON helpers, `http.MaxBytesReader` (1 MB) on every POST body, reject empty messages, validate the model against config. Stop cancels the run context; the gateway already saves partial output on error. Emit `abort` for a user stop.
4. **History converter** (`history.go`). A pure function from `[]fantasy.Message` to UIMessages, per the History shape rules.
5. **Entry point** (`cmd/bai/web.go`, dispatch in `main.go` next to the existing `git` subcommand). `bai web [--addr 127.0.0.1:0]`. Reuse the setup from `start()` (logger, DB, migrate, broker, gateway) by extracting a shared helper only if it stays small and leaves the TUI behaviour untouched. Skip the theme, Bubble Tea and sound. Start one goroutine that drains the broker, otherwise the agent deadlocks. Print the URL with its token to stdout only, never to `bai.log`. On SIGINT cancel the active run, then `Shutdown` the server. If the shadow git repo is not initialized, the run must still complete; log a hint, build no prompt UI.
6. **Embedding.** `web/embed.go` with `//go:embed all:dist` and `fs.Sub`. `go build ./...` must succeed on a fresh clone with no Node installed and no bundle built, so commit only `web/dist/.gitkeep`. The root `.gitignore` ignores every `dist` directory, so un-ignore that one file deliberately. When `index.html` is missing from the embedded files, serve a plain 503 that says to run `make build-web`. Cache headers: hashed `/assets/*` immutable, `index.html` no-cache. Do not add compression; this is loopback.
7. **Makefile.** Add a `build-web` target (install with a frozen lockfile, build, then recreate `web/dist/.gitkeep` because Vite empties the folder) and make `build` depend on it. Do not edit `.goreleaser.yml`; put the `before.hooks` lines it needs in your final report.
8. **Checks** (Go `testing` plus `utils/tests`, table driven, no testify, no `time.Sleep`):
   - frame encoder: text, reasoning and tool sequence matches expected JSON exactly
   - run registry: two subscribers get the same frames, a late subscriber replays, a stalled subscriber is dropped while the publisher never blocks, a second start is refused
   - broker drain: more than 64 events with no TUI do not block
   - guard: missing or wrong token, bad Host and bad Origin are rejected
   - history converter: assistant, tool result, second step and user message in one conversation

   `go test -race ./...` must pass apart from failures that existed before you started.

## Frontend tasks

Build in `web/`. Do not import from, copy files out of, or modify `frontend/`; copy its stack instead.

- **Scaffold.** Same stack as `frontend/package.json`: React 19, Vite, Tailwind 4 with `@tailwindcss/vite`, shadcn with the `@/` alias, `lucide-react`, `clsx`, `tailwind-merge`, `class-variance-authority`, oxlint and oxfmt, pnpm. No router library: there are two routes, `/` and `/c/:id`, so a small hook on `location.pathname` and `popstate` is enough. No state library and no data-fetching library; `fetch` and `useEffect` cover the lists.
- **AI SDK.** Install `ai` and `@ai-sdk/react` at their current stable versions and read the installed `.d.ts` files before coding. The hook is transport based: `sendMessage`, `DefaultChatTransport`, `prepareSendMessagesRequest`, `resume`, `stop`, `addToolApprovalResponse` (do not use the last one).
- **AI Elements.** Run `npx ai-elements@latest` with no arguments and it installs every component. Never do that. List what is available first, then add one at a time: `conversation`, `message`, `prompt-input`, `reasoning`, `tool`, and `code-block` only if the size budget holds. After each add, read the files it wrote and the dependencies it pulled. Reject anything that pulls a graph or canvas library, an animation runtime or similar weight, and justify any component beyond that set in your report. Components are source you now own: do not restyle them heavily.
- **Transport.** Send only the last message, because the server owns history in SQLite:

```ts
new DefaultChatTransport({
  api: '/api/chat',
  prepareSendMessagesRequest: ({ id, messages }) => ({
    body: { id, message: messages.at(-1) },
  }),
})
```

- **New chats.** On the transient `data-conversation` part, call `history.replaceState` to `/c/<id>` and refresh the sidebar. Do not navigate through React state and do not change the `Chat` component's key, because remounting would kill the in-flight stream.
- **Rendering.** `text` through `MessageResponse` (streamdown), `reasoning` through `Reasoning`, any part whose type starts with `tool-` through `Tool`, `ToolInput` and `ToolOutput`, `step-start` ignored. User messages are plain text.
- **Stop.** The button calls `POST /api/chat/{id}/stop` and then the hook's `stop()`. Aborting the fetch alone only disconnects the browser; the run lives on the server.
- **Reload mid-run.** On load fetch `/api/state`. If `running` is set and you are on `/`, open `/c/<id>` with `resume` enabled. History for that conversation stops before the run's output, and the resume stream replays it.
- **Layout.** Sidebar with New chat and the conversation list; header with the folder name and a model select; conversation area; prompt input. Under about 768 px the sidebar collapses. Enter sends, Shift+Enter inserts a newline. Light and dark through `prefers-color-scheme`. Icon buttons have accessible names, focus is visible, streaming status is announced politely.
- **Errors.** Show the stream's `error` text and the hook's error. A 409 reads "a run is already active". A 401 shows one line telling the user to open the URL printed by `bai web`.
- **Dev loop.** Vite proxies `/api` to the Go address. With `bai web --dev` the server also accepts the Vite origin `http://localhost:5173`. Confirm the proxy does not buffer SSE.
- **Budgets.** Measure and report the size of `web/dist`. Target a total under about 3 MB and an initial JavaScript bundle under about 300 KB gzipped, with code highlighting (Shiki) lazy loaded and limited to a few languages. Anything heavier needs a stated reason.
- **Checks.** `pnpm lint`, `pnpm format:check` and `tsc --noEmit` pass. No `any`, no `@ts-ignore`.

## Security requirements

This server lets a model run bash as the user. Anyone holding the token has the same power the TUI user has. Treat this section as non-negotiable.

- Bind `127.0.0.1` by default. A non-loopback `--addr` is allowed only when passed explicitly, and prints a loud warning. There is no TLS in v1.
- Token: 32 bytes from `crypto/rand`, hex encoded, new on every launch (an optional `BAI_WEB_TOKEN` env var is allowed for the dev loop). `GET /?t=<token>` sets an `HttpOnly`, `SameSite=Strict`, `Path=/` cookie and redirects to `/` so the token leaves the address bar. `Authorization: Bearer` is also accepted for curl. Compare with `subtle.ConstantTimeCompare`. Never write the token to the log file.
- Every `/api/*` request needs the token. The `Host` header must be the listen address or `localhost` on the same port, which blocks DNS rebinding. A non-GET request that carries `Origin` must match `Host`. POST bodies must be `application/json`.
- Send no CORS headers. The only exception is the Vite origin `http://localhost:5173` when `--dev` is set.
- Add no generic file, exec or filesystem endpoints. Everything the agent can do goes through the model path the TUI already uses.
- Return hand-built response structs only. Never return provider keys, base URLs, config contents or environment values.
- Cap request bodies (1 MB), the frame buffer (about 4 MB per run) and subscribers per run (16, then 503).
- Treat model text and tool output as untrusted. Render through streamdown, never `dangerouslySetInnerHTML`. Send `Content-Security-Policy: default-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'`, `X-Content-Type-Options: nosniff`, `Referrer-Policy: no-referrer` and `X-Frame-Options: DENY`. Relax a directive only if the app breaks, and say which one in the report.
- Log error details with `slog`; return short messages to the browser.

## Milestones and acceptance checks

Stop after each milestone and report its evidence before starting the next.

| Milestone | Work | Acceptance |
| --- | --- | --- |
| M0 Baseline | Create task.md first, then a new branch. Run `go test -race ./...` and note existing failures. Read Fantasy's `AgentStreamCall` and list its callbacks. Install `ai` and check every frame in the contract against its schema. | task.md exists with every milestone. A short note: callbacks found, frames confirmed or corrected, baseline failures. |
| M1 Streaming slice | Gateway seam, run registry, frame encoder, auth guard, `POST /api/chat`, `bai web` printing its URL. | `curl -N` with the bearer token and a prompt that triggers a tool shows ordered frames ending in `[DONE]`. A second concurrent POST returns 409. New Go tests pass with `-race`. Existing `internal/agent` and `internal/tui` tests still pass. |
| M2 UI slice | `web/` app, AI Elements, chat page, sessions, model picker, history converter, embedding, Makefile target. | `make build`, then `./bin/bai web`, open the printed URL. A prompt that reads a file and runs bash shows reasoning, tool cards and markdown. Reloading shows the same conversation. The model switch applies to the next run. `go build ./...` works on a fresh clone with no Node installed. |
| M3 Resilience | Stop, resume, multiple viewers, slow-subscriber drop. | Reloading mid-run resumes the stream without duplicating text. Two tabs show the same run. Stop saves the partial message and the UI shows it as stopped. Closing the browser does not stop the run. SIGINT shuts down cleanly. |
| M4 Hardening | Walk the security list with curl. Measure bundle size. Lint and format. | No token gives 401. A wrong `Host` gives 403. A cross-origin POST gives 403. Bundle sizes are reported against the budgets. `make test`, `pnpm lint`, `pnpm format:check` and `tsc --noEmit` pass. |

## Do not

**Git and files**

- Do not touch the user's uncommitted work, and do not use `git stash`, `git checkout .`, `git reset --hard` or `git add -A`.
- Do not run `make format`, `make lint-fix` or `gofmt -w .`. Format only files you created or edited.
- Do not edit `.goreleaser.yml`, `frontend/`, the database migrations or schema, or `internal/tui/**` (a minimal compile fix forced by the gateway seam is the only exception).
- Do not commit, push or open pull requests unless asked.

**Backend**

- Add no Go dependency. No Gin, chi, echo, fiber, Huma or gorilla, no WebSocket, no Redis, no ORM, no auth or JWT library. If you believe one is unavoidable, stop and ask.
- Do not use broker events as the web data source, and do not read from the broker's channel anywhere except the drain goroutine.
- Do not tie a run to the request context, block the agent on a subscriber, or keep any buffer unbounded.
- Do not call `SetActiveConversation` while a run is active.
- Do not serialize config structs, and do not add generic file, exec or filesystem endpoints.
- Do not add tool approvals, a theme port, conversation delete or rename, or a project switcher.
- Do not invent the wire format from memory. Verify against the installed `ai` package.

**Frontend**

- No Next.js, Node server, `streamText`, Vercel AI Gateway or `@ai-sdk/openai` and similar providers. bai has its own providers.
- No Redux, Zustand, TanStack Query, SWR or router library. No CSS-in-JS, i18n, analytics, service worker, Electron, Tauri, Storybook, Playwright or Cypress.
- Do not run `npx ai-elements@latest` without arguments, hand-build components that AI Elements already has, or resend the full message history from the client.
- No attachments, voice, image generation, message branching, canvas or workflow views, worktrees or terminal in v1.
- No `any` and no `@ts-ignore`.

**Process**

- Do not report a check as passing without running it. Do not skip hooks or use `--no-verify`.
- Do not fix unrelated failing tests or lint warnings; report them.
- Do not leave dev servers running.
- Add no documentation beyond task.md and a short `bai web` section in `readme`.

## Final report format

When all milestones are done, reply with a report, not a code dump:

1. One paragraph on what was built and how to run it.
2. Per milestone: the commands you ran and what they showed. Confirm every item in task.md is ticked or explained in its notes.
3. Test failures that existed before your changes.
4. Deviations from this prompt, each with its reason.
5. The Fantasy callbacks you found, and which frames you verified against the `ai` package and which you could not.
6. Bundle sizes against the budgets.
7. The exact `before.hooks` lines `.goreleaser.yml` needs to build `web/` before `go build`.
8. Open risks: the token grants shell access, one run per process, and what happens when the shadow git repo is not initialized.
9. At most three suggested next steps, taken from the "Later" rows of the scope table.
