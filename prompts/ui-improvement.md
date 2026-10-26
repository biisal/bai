# bai web UI: polish

Follow-up to `impl.md`. Restyle `web/` so it looks clean like the shadcn chat/dashboard examples: real sidebar, almost no shadow, almost no border. Frontend only, no behavior changes. Add a "UI polish" section to `task.md` and work from it.

## Look

- Separate regions with background tone (`background`, `muted`, `sidebar`), not lines. Border only where two same-tone surfaces meet.
- No borders on messages, tool rows, header or buttons. No border + ring + shadow stacking.
- No shadows, except `shadow-xs` on the prompt input and floating layers (select, popover).
- Neutral shadcn tokens for light and dark, one accent (send button, focus ring, active sidebar item). One radius scale from `--radius`. No hard-coded colours.
- 14px body, message column `max-w-3xl` centered, `gap-6`.

## Tasks

1. **Tokens.** Update `index.css` to the current shadcn neutral set: softer `--border`/`--input`, add `--sidebar-*`. Drop custom shadows and unused variables.
2. **Sidebar.** `npx shadcn@latest add sidebar`, read what it pulls. Replace the hand-built one: `SidebarProvider`, `Sidebar` + `SidebarInset`, header with wordmark and New chat, "Recent" group of `SidebarMenu` items (`isActive`, truncated titles, `Skeleton` while loading), footer with folder name and version. `SidebarTrigger`, Cmd/Ctrl+B, drawer below 768px.
3. **Header.** Slim row, no border or shadow: trigger, folder name, shadcn model select on the right grouped by provider.
4. **Messages.** User: right-aligned `bg-muted` bubble. Assistant: plain text, no bubble. Reasoning collapsed and muted. Tool parts: one flat `bg-muted/40` row (icon, name, status) that expands to input/output. Restyle AI Elements through `className` and tokens only.
5. **Prompt input.** Floating, `rounded-xl`, `bg-muted/50`, focus ring instead of border, auto-grow, send button becomes Stop while streaming.
6. **Empty state.** Centered greeting, folder name, 3-4 ghost suggestion chips that fill the prompt without sending.
7. **Small pieces.** One at a time via the shadcn CLI: `tooltip` on icon buttons, `scroll-area` for the list, `sonner` for error toasts (401 stays a full-page muted notice), `badge` for running state.
8. **Cleanup.** Delete what the new shell replaces. Grep out forbidden `shadow-*`, `border-*`, `bg-gray-*` leftovers. Net LOC flat or down.

## Constraints

- Keep the `Chat` key and the `history.replaceState` flow untouched.
- CSP unchanged; skip any piece that needs it relaxed.
- Dependencies only what the shadcn CLI adds for the components above. Initial JS grows at most ~30 KB gzipped.
- Keep focus rings, aria-labels, and 4.5:1 contrast in both themes.
- No extra features (theme switcher, settings, search, delete/rename).

## Acceptance

- `pnpm lint`, `pnpm format:check`, `tsc --noEmit` pass (note pre-existing failures first).
- `grep -rnE "shadow-(sm|md|lg|xl|2xl)|border-(gray|slate|zinc)" web/src` is empty outside kept generated shadcn files; justify each remaining `border` in one line in `task.md`.
- `make build-web`, `./bin/bai web`, then check in light and dark: sidebar, new chat, collapse, 375px drawer, empty state, streaming reply with reasoning and a tool call, stop, model switch, error toast. Describe what you saw.

## Report

Per task done/skipped/deviated, components added or rejected, bundle size before/after, leftovers.
