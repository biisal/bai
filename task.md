# bai web — task list

## M0 Baseline
- [x] Create task.md
- [x] Create a new branch (stage only files I create/edit)
- [x] Run `go test -race ./...` and record pre-existing failures
- [x] Read `charm.land/fantasy` `AgentStreamCall` and list callbacks
- [x] Install `ai` in a scratch spot and verify every contract frame against its schema
- [x] Acceptance: task.md exists; note callbacks, frame corrections, baseline failures

## M1 Streaming slice
- [x] Gateway seam: decorator variant of `StreamChat`, chain base callbacks minimally
- [x] Run registry (`run.go`): one active run, capped frames, replay-safe subscribers, drop slow, 409 busy
- [x] Frame encoder: struct-marshaled AI SDK v1 frames
- [x] Auth guard: token cookie/bearer, Host/Origin checks, constant-time compare
- [x] `POST /api/chat` handler with 1 MB cap, empty-message reject
- [x] `bai web` entry point prints URL+token; broker drain goroutine; SIGINT cancel+shutdown
- [x] Tests: frame encoder, run registry, broker drain >64, auth guard
- [x] Acceptance: `curl -N` ordered frames ending `[DONE]`; concurrent POST 409; new tests pass `-race`; existing agent/tui tests pass

## M2 UI slice
- [x] Scaffold `web/` (React 19, Vite, Tailwind 4, shadcn alias, pnpm, oxlint/oxfmt)
- [x] Install `ai` + `@ai-sdk/react`; read installed `.d.ts`
- [x] Add AI Elements: conversation, message, prompt-input, reasoning, tool (code-block skipped: `@streamdown/code` statically imports all of Shiki)
- [x] Transport: `DefaultChatTransport` sending only the last message
- [x] New chats: handle transient `data-conversation`, `history.replaceState`, refresh sidebar
- [x] Rendering: text/streamdown, reasoning, tool-*, ignore step-start
- [x] Sessions sidebar, header folder name, model select, responsive collapse, a11y
- [x] Backend: `/api/state`, `/api/models`, `PUT /api/model`, `/api/conversations`, messages, history converter
- [x] Embedding: `web/embed.go` + `web/dist/.gitkeep`, 503 hint, cache headers, `.gitignore` un-ignore
- [x] Makefile `build-web`, `build` depends on it
- [x] Acceptance: `make build`; `./bin/bai web` shows reasoning/tool cards/markdown; reload same conversation; model switch next run; fresh-clone `go build ./...`

## M3 Resilience
- [x] Stop endpoint + UI button (`POST stop`, then hook `stop()`); UI shows "Response stopped."
- [x] Resume endpoint `/api/chat/{id}/stream`: 204 idle, replay then live
- [x] Reload mid-run uses `/api/state` + resume, history keeps the run's user message and stops before run output
- [x] Multiple viewers same run; slow subscriber dropped, publisher never blocks
- [x] Closing browser does not stop run; SIGINT clean shutdown
- [x] Shadow git not initialized: run still completes, log hint
- [x] Acceptance: reload resumes without duplicate text; two tabs same run; stop saves partial + UI stopped; SIGINT clean

## M4 Hardening
- [x] Walk security list with curl: 401 no token, 403 bad Host, 403 cross-origin POST
- [x] CSP/security headers, no CORS except Vite origin under `--dev`
- [x] Measure `web/dist` and initial JS gzip against budgets
- [x] `make test`, `pnpm lint`, `pnpm format:check`, `tsc --noEmit`
- [x] Acceptance: security checks pass; sizes reported; all linters/tests pass

## UI polish
- [x] Add this section and capture baseline checks/bundle
- [x] Tokens: current neutral set, softer border/input, sidebar tokens, prune unused vars
- [x] Sidebar: shadcn `sidebar` shell, wordmark, New chat, Recent list, skeleton, footer, Cmd/Ctrl+B, mobile drawer
- [x] Header: slim borderless row with trigger, folder, grouped shadcn select
- [x] Messages: right-aligned user bubble, plain assistant, muted reasoning, flat tool row
- [x] Prompt input: floating rounded muted surface, focus ring, auto-grow, Stop affordance
- [x] Empty state: greeting, folder name, suggestion chips that fill the prompt
- [x] Small pieces: tooltip, scroll-area, sonner, badge
- [x] Cleanup: remove replaced shell, grep forbidden shadows/borders/colors, net LOC note recorded
- [ ] Acceptance: checks pass; forbidden grep only justified generated borders; build and describe light/dark checks (static/build pass; visual pending)

## Features
- [x] Task section, baseline bundle/checks recorded
- [x] `GET /api/files?q=` scoped/ignored/capped/traversal-rejected + tests
- [x] `@` picker with debounce, keyboard, TUI-compatible insertion, message chips (menu sits above composer, not under caret)
- [x] Slash menu mapped to `/new`, `/models`, `/sessions` (`/themes` omitted by request)
- [x] Sessions search, Today/Yesterday/Earlier, rename/delete with dialogs; 409-while-running (delete confirm lives in Manage chat)
- [x] Settings dialog: Models, Appearance (web palette + light/dark/system), Skills/plugins, Git, About; secrets absent
- [x] Theme parity skipped by request: web keeps its single built-in palette
- [x] Changes button + lazy panel; read-only fixed-arg diff of last shadow commit, 1 MB cap
- [x] Run usage/time, copy/regenerate, background-tab notification opt-in
- [x] Cmd/Ctrl+K palette, `?` shortcut sheet, Esc-stop rule
- [x] New routes guarded; curls for 401/403; `go test -race`, web checks
- [ ] Bundle report done; browser checks pending (no GUI/browser automation installed)

## Notes and blockers
- Branch: `bai-web-ui`.
- Baseline `go test -race ./...`: all packages pass, no pre-existing failures.
- Fantasy v0.45.2 callbacks: OnAgentStart/Finish, OnStepStart/Finish, OnFinish, OnError, OnChunk, OnWarnings, OnTextStart/Delta/End(id,text), OnReasoningStart/Delta/End(id,reasoning), OnToolInputStart(id,toolName)/Delta(id,delta)/End(id), OnToolCall(ToolCallContent), OnToolResult(ToolResultContent), OnSource, OnStreamFinish, OnRetry(ProviderError,delay). Tool input streaming exists.
- Verified against `ai@7.0.153` `UIMessageChunk`: all contract frames exist (`start`, `start-step`, `finish-step`, `reset-step`, text/reasoning start/delta/end, tool-input-start/delta/available, tool-output-available/error, `finish`, `abort.reason`, `error.errorText`, `data-*` with `transient`). Header `x-vercel-ai-ui-message-stream: v1` and `data: [DONE]` confirmed. No corrections needed.
- `DefaultChatTransport` reconnect URL is `<api>/<chatId>/stream` (204 = no active stream), matching the contract.
- M1 evidence: local fake OpenAI stream produced ordered text/tool/step frames ending `[DONE]`; a concurrent POST returned 409; `go test -race ./...` passes.
- Deviation: history converter takes `[]repo.Message` instead of `[]fantasy.Message` so it can use the DB id of the first message in a group, as required by the History shape rules.
- M2 evidence: `make build` passes; `./bin/bai web` serves the built SPA; fake OpenAI stream rendered text/tool frames; model switch sent `other-model` on the next run; history excludes the active run's output but now keeps its user message (`trimRunUserMessage` returns `messages[:i+1]`); resume replay matched the live stream byte-for-byte after the fix; idle resume returns 204.
- Bundle evidence: `web/dist` 2.5 MB; entry 676.84 kB minified / 209.50 kB gzip; lazy Streamdown chunk 488.57 kB / 150.41 kB gzip; Shiki language chunks lazy. No browser automation is available, so visual behavior was verified by API/static checks only.
- M3 evidence: disconnecting the SSE client early left the run completed in the DB; two clients joined the same run and received byte-identical 40-line streams; stop returned 204, emitted `abort` `user cancelled`, and saved partial assistant text; SIGINT stopped the server in ~100 ms; an uninitialized shadow git logged the hint and the run still completed. UI stop now shows "Response stopped." in the footer/status and after the message list.
- M4 evidence: missing/wrong token -> 401; bad Host -> 403; cross-origin POST -> 403; same-origin POST without JSON -> 415; non-dev responses have no CORS header even for the Vite origin; `--dev` preflight/GET return the Vite origin header while evil origins stay 403; CSP/nosniff/no-referrer/DENY present on 401 and 200; `/api/models` contains no keys or base URLs. Vite proxy SSE delivered the first text delta within 0.8 s, so it does not buffer; `VITE_BAI_TOKEN` injects a dev-only bearer token into proxied requests. `make test` and all web checks pass.
- Budgets: `web/dist` 2.5 MB; entry `index-CUY5POIS.js` 677,072 bytes / 209,583 bytes gzip; lazy Streamdown `mermaid-*.js` 488.57 kB / 150.41 kB gzip. Under the 3 MB total and 300 KB initial-JS gzip targets.
- UI polish baseline: `pnpm lint`, `pnpm format:check`, `tsc --noEmit` pass; `web/dist` 2.5 MB; entry 209,583 bytes gzip.
- UI polish result: all three checks pass; `make build-web` and `make build` pass. Entry `index-81BJAwji.js` 681,832 bytes / 211,627 bytes gzip, +2,044 bytes (+1.0%) vs baseline. `web/dist` 2.7 MB. UI-only chunks load after mount: sidebar 10.56 kB, model select 17.71 kB, tooltip 9.93 kB, sonner 0.86 kB gzip.
- UI polish border justifications: `ui/sheet.tsx` borders/shadows are the mobile drawer floating edge; `ui/sidebar.tsx` desktop `border-r` is negated by the app's `border-none`; `ui/select.tsx`/`ui/dropdown-menu.tsx`/`ui/hover-card.tsx` shadows are allowed floating layers; `ui/sidebar-base.tsx` inset-variant shadow is unused (offcanvas variant); prompt input's generated border is overridden with `border-0`; model select uses `border-0`. No `border-*`/`shadow-*`/hard-coded colours remain in hand-written app components.
- UI polish deviations: header tooltip, model select, toaster, and sidebar are lazy-loaded to keep entry growth at ~2 KB gzip instead of ~66 KB; fallbacks keep layout stable. `scroll-area` drives the Recent list; `sonner` shows model-switch errors; 401 remains the full-page muted notice; tool parts keep the `Badge` status. Custom app code is 1,556 LOC vs ~1,424 baseline (+132 for the empty state, suggestions, scroll/skeleton, and lazy boundary); generated shadcn adds 982 LOC.
- UI polish smoke: built `bin/bai web` served the new SPA on a temp config/port; token redirect + cookie returned index 200, entry JS 200 (681,832 bytes), CSS 200 (92,119 bytes), `/api/state` 200. No GUI/browser automation is installed, so light/dark and 375px visual checks are still pending.
- `.goreleaser.yml` before-hooks needed: keep `- go mod tidy` and add `- make build-web` (release runner needs pnpm and Node).
- Features evidence: `GET /api/files` returned `prompts/features.md`, rejected `../secret` with 400; settings returned only `about/active/git/models` and grep found no key/base URL; skills/plugins listed the temp fixtures. `curl` showed 401 without token and 403 on wrong Host or cross-origin GET; PATCH/DELETE with evil Origin were 403. Rename returned `{id,title}`, delete returned 204 and removed messages. `go test -race ./...` passes; `pnpm format:check`, `pnpm lint`, `tsc --noEmit`, and `make build` pass.
- Features diff fix: the gateway auto-commits after a run, so `git diff HEAD` was empty. The endpoint now uses fixed `git show --format= --unified=3 HEAD --` args to show the last run's commit; a unit test executes a real temp shadow repo and asserts the `-old/+new` hunks.
- Features bundle: entry `index-BARW9V3Z.js` 693.64 kB / 215.07 kB gzip, +3.44 kB gzip vs 211,627 B baseline (well under +40 kB). Lazy chunks: settings 2.70 kB, changes 1.59 kB, command palette 18.83 kB gzip.
- Theme deviation by request: no TUI theme parity, theme selector, or `/themes` action. Settings Appearance now only has light/dark/system plus notification opt-in; the web keeps one built-in neutral palette.
- Browser acceptance remains pending: no GUI or browser automation is installed. Visual checks for the `@` menu placement, chips, theme modes, Changes panel, notifications, and 375px layout were not run.
- Web server port: `bai web` now defaults to `127.0.0.1:7777`, walks forward up to 100 ports when busy, then falls back to an OS-assigned port. It opens the token URL in the browser unless `--no-open` is passed. Verified with two instances: 7777 then 7778; `TestListenIncrementsBusyPort` covers the increment.
