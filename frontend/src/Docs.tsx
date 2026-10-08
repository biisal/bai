import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { ArrowLeft } from "lucide-react";
import { Streamdown } from "streamdown";
import { code as codePlugin } from "@streamdown/code";
import "@/styles/streamdown.css";
import { installCmd, repoUrl, siteName } from "@/config";
import { ThemeJsonEditor } from "@/components/Themes";
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";

type Doc = { id: string; title: string; body: string };

const docs: { group: string; items: Doc[] }[] = [
  {
    group: "Getting Started",
    items: [
      {
        id: "introduction",
        title: "Introduction",
        body: `${siteName} is a terminal-based AI coding agent. Single binary, TUI interface. Status: early development — the core chat loop and coding agent tools work.

## Features

- Streaming chat with visible model reasoning, retries on transient failures
- Provider and model switching at runtime, persisted in SQLite
- Directory-scoped sessions: conversations belong to the folder you started in
- Built-in agent tools: \`read\`, \`write_file\`, \`edit_file\`, \`bash\`
- Custom tools: executable scripts registered in a JSON manifest
- Skills — \`SKILL.md\` instruction files from global and project folders
- \`AGENTS.md\` injected into the system prompt
- Shadow git: an undo history of every change, invisible to your project's git
- Slash commands, file finder, themes, notification sound / Herdr states

## Requirements

- An API key for at least one supported provider — that's all you need to start chatting
- \`git\` on PATH — recommended, not required; it only powers shadow git (undo history), bAI works fine without it
- Go 1.26+ — only if you build from source; the curl install doesn't need it`,
      },
      {
        id: "installation",
        title: "Installation",
        body: `One static binary, ~16 MB. No runtime, no model downloads.

## With curl

\`\`\`bash
${installCmd}
\`\`\`

## From source

\`\`\`bash
git clone ${repoUrl} && cd bai
make build      # builds bin/bai
make run        # build and run
\`\`\`

## Make targets

\`\`\`text
make build            build the binary into bin/bai
make install          build and move to ~/.local/bin
make run              build and run
make dev              build and run in dev mode (debug logging)
make watch            hot reload on go file changes (needs watchexec)
make test             run the test suite
make lint             golangci-lint run
make format           gofmt -w .
make db-generate      sqlc generate for the query layer
make frontend-build   build the landing page
make frontend-dev     run the landing page dev server
make clean            remove build artifacts
\`\`\``,
      },
    ],
  },
  {
    group: "Guides",
    items: [
      {
        id: "configuration",
        title: "Configuration",
        body: `Config lives at \`~/.config/bai/config.json\`. Copy the example and fill in your API keys:

\`\`\`bash
mkdir -p ~/.config/bai
cp config.example.json ~/.config/bai/config.json
\`\`\`

## Full example

\`\`\`json
{
  "database_path": "~/.config/bai/bai.db",
  "log_file_path": "~/.config/bai/bai.log",
  "sound_path": "",
  "skills_paths": [],
  "git_dir_name": ".bai_git",
  "plugins_path": "~/.config/bai/tools/tools.json",
  "providers": [
    {
      "name": "Anthropic",
      "format": "anthropic",
      "api_key": "sk-ant-...",
      "base_url": "https://api.anthropic.com",
      "models": [
        { "id": "claude-sonnet-4-20250514", "context": 200000, "max_output": 64000 }
      ]
    }
  ]
}
\`\`\`

## Keys

- \`format\` — \`anthropic\` or \`openai-compatible\`
- \`variant\` — optional, e.g. \`opencode\`
- \`sound_path\` — audio file played when a turn finishes (see Notifications)
- \`skills_paths\` — extra skill folders, on top of the default ones (see Tools & Skills)
- \`git_dir_name\` — shadow git directory, default \`.bai_git\`
- \`plugins_path\` — custom tools manifest, default \`~/.config/bai/tools/tools.json\`
- Point elsewhere with \`-config path/to.json\`

## Validation

- At least one provider is required; provider names must be unique
- Every provider needs a non-empty \`base_url\` and \`format\`
- Missing \`database_path\`, \`log_file_path\`, \`git_dir_name\` and \`plugins_path\` fall back to defaults`,
      },
      {
        id: "usage",
        title: "Usage",
        body: `\`\`\`bash
bai                        use the default config
bai -config path/to.json   use a specific config
bai -dev                   debug logging to the log file
bai git <args...>          run git against the shadow repo (bai git log)
\`\`\`

The TUI opens with provider and model selection. Type a message and press enter — the response streams, and tool activity (file reads, writes, bash) is rendered as it happens. Esc interrupts a running turn; nothing half-finished is lost.

## Key bindings

\`\`\`text
enter        send the message or run the command
esc          interrupt the agent / close the list
ctrl+c       quit
/            slash commands
@            find files
tab, ctrl+n  move down in a list
shift+tab,
ctrl+p       move up in a list
\`\`\``,
      },
      {
        id: "commands",
        title: "Commands",
        body: `Type \`/\` to list commands, keep typing to filter, enter to run.

\`\`\`text
/models      list provider/model pairs, switch the active model
/sessions    conversations for the current directory, resume one
/new         start a new conversation
/themes      list themes, switch the active theme
/git_init    enable shadow git in this directory
/exit        quit
@query       file finder: filter by name, enter inserts the path
\`\`\`

Sessions, models and themes are persisted: the active model and theme survive a restart, and each conversation is stored with the directory it was started in, so \`/sessions\` only ever shows the current folder's history.`,
      },
      {
        id: "tools-and-skills",
        title: "Tools & Skills",
        body: `## Built-in tools

\`\`\`text
read        path, offset, limit
write_file  path, purpose, content
edit_file   path, purpose, edits[] (old_text, new_text)
bash        command, purpose, timeout (seconds, optional)
\`\`\`

The \`purpose\` string is what the agent writes down before changing anything — it becomes the shadow git commit message.

## Custom tools

A custom tool is an executable script plus one entry in the JSON manifest at \`plugins_path\`:

\`\`\`json
[
  {
    "file_name": "weather.sh",
    "description": "Get the current weather for a city. ",
    "input_schema": { "city": "city name, e.g. Patna" }
  }
]
\`\`\`

- The script lives next to the manifest and is run as \`'<dir>/weather.sh' city=Patna\` — arguments are \`key=value\` pairs, parse by key, no stdin
- Tool name = \`file_name\` without its extension; the script must be \`chmod +x\`
- If the manifest fails to parse, all custom tools are dropped — keep it valid JSON

## Skills

Skills are folders containing a \`SKILL.md\` with YAML frontmatter (\`name\` must match the folder name, lowercase alphanumeric with hyphens). bAI loads from:

\`\`\`text
~/.config/bai/skills     plus any extra skills_paths
~/.agents/skills         ~/.claude/skills
.bai/skills              ./skills
.agents/skills           .claude/skills
\`\`\`

bAI also ships one built-in skill, \`internal:skills:tools_maker\`, which teaches the agent how to write new custom tools. Add \`AGENTS.md\` (or \`agents.md\` / \`agent.md\`) in the project root for instructions that are always injected into the system prompt.`,
      },
      {
        id: "shadow-git",
        title: "Shadow Git",
        body: `Shadow git keeps an undo history of every change — invisible to your project's git. It runs the real \`git\` binary against a separate directory (\`.bai_git\` by default) while your work tree stays untouched:

\`\`\`bash
git --git-dir .bai_git --work-tree . <args>
\`\`\`

## How it works

- On first run in a directory bAI asks whether to enable it: yes / no / don't ask again (the choice is remembered per directory)
- After every agent turn, changed files are committed automatically. The commit message is the agent's stated purposes, or an \`auto-commit\` line with a status summary
- \`.bai_git\` is added to your \`.gitignore\` so it never reaches your remote
- Re-enable anytime with \`/git_init\`

## Inspecting history

\`bai git\` passes arguments straight to the shadow repo, so your project's own git is never involved:

\`\`\`bash
bai git log --oneline -10
bai git diff HEAD~1
bai git show
\`\`\``,
      },
      {
        id: "notifications",
        title: "Notifications",
        body: `bAI reports turn state (working / done / failed / blocked) so you can look away from the terminal.

## Sound

Set \`sound_path\` to an audio file. It plays when a turn completes successfully (errors are silent). A bad path is logged and skipped, never fatal.

## Herdr

When running inside a Herdr pane (\`HERDR_ENV=1\`), bAI claims the pane on startup and reports state transitions to the Herdr server instead of playing a sound — Herdr raises its own notification when the turn goes idle. Outside Herdr every call is a no-op, so nothing changes if you don't use it.`,
      },
    ],
  },
  {
    group: "Reference",
    items: [
      {
        id: "providers",
        title: "Providers",
        body: `Any of these can be configured via config.json:

\`\`\`text
OpenAI      openai-compatible  https://api.openai.com/v1
Anthropic   anthropic          https://api.anthropic.com
DeepSeek    openai-compatible  https://api.deepseek.com/v1
Qwen        openai-compatible  https://dashscope.aliyuncs.com/compatible-mode/v1
Groq        openai-compatible  https://api.groq.com/openai/v1
OpenCode    openai-compatible  https://opencode.ai/zen/v1   (variant: opencode)
\`\`\`

Any other OpenAI-compatible endpoint works too — point \`base_url\` at it and list your models with \`id\`, \`context\` and \`max_output\`.

## Switching models

\`/models\` lists every configured provider/model pair. The selection is written to SQLite and restored on the next start.`,
      },
      {
        id: "themes",
        title: "Themes",
        body: `Themes are JSON files in \`~/.config/bai/themes\`. Pick one with \`/themes\`; the choice is persisted and reloaded on start.

The keys are shadcn-inspired token names. Values are ANSI color numbers or hex colors. Missing keys fall back to the built-in default theme, and a theme that disappears from disk falls back the same way.`,
      },
      {
        id: "project-layout",
        title: "Project Layout",
        body: `\`\`\`text
cmd/bai/              CLI entry: flags, config, shadow git passthrough
internal/agent/       gateway, streaming loop, provider building, DB store
internal/agent/core/  built-in tools, system prompt, skill injection
internal/skills/      SKILL.md loader and the built-in tools_maker skill
internal/tui/         bubbletea model, views, slash commands
internal/git/         shadow git (--git-dir .bai_git) and .gitignore rules
internal/pubsub/      event broker feeding the TUI
internal/notifier/    notification sound and Herdr state reports
internal/db/          sqlite connection, goose migrations, sqlc queries
internal/config/      config loading, validation, themes
internal/files/       current directory and the @ file finder
\`\`\``,
      },
    ],
  },
];

export default function Docs() {
  const [active, setActive] = useState(docs[0].items[0].id);

  useEffect(() => {
    window.scrollTo(0, 0);
  }, [active]);

  const flat = docs.flatMap((g) => g.items);
  const idx = flat.findIndex((d) => d.id === active);
  const current = flat[idx];
  const prev = flat[idx - 1];
  const next = flat[idx + 1];

  return (
    <div className="min-h-screen bg-background text-foreground antialiased selection:bg-primary/20 selection:text-primary">
      <SidebarProvider>
        <Sidebar>
          <SidebarHeader>
            <div className="px-2 py-1.5 font-['JetBrains_Mono'] text-xs uppercase tracking-[0.2em] text-muted-foreground">
              {siteName} docs
            </div>
          </SidebarHeader>
          <SidebarContent>
            {docs.map((group) => (
              <SidebarGroup key={group.group}>
                <SidebarGroupLabel>{group.group}</SidebarGroupLabel>
                <SidebarGroupContent>
                  <SidebarMenu>
                    {group.items.map((doc) => (
                      <SidebarMenuItem key={doc.id}>
                        <SidebarMenuButton
                          isActive={doc.id === active}
                          onClick={() => setActive(doc.id)}
                        >
                          {doc.title}
                        </SidebarMenuButton>
                      </SidebarMenuItem>
                    ))}
                  </SidebarMenu>
                </SidebarGroupContent>
              </SidebarGroup>
            ))}
          </SidebarContent>
        </Sidebar>

        <SidebarInset>
          <header className="sticky top-0 z-40 flex h-14 items-center justify-between border-b border-border bg-background/80 px-4 backdrop-blur-sm">
            <div className="flex items-center gap-3">
              <SidebarTrigger />
              <Link
                to="/"
                className="flex items-center gap-2 text-sm text-muted-foreground transition-colors hover:text-foreground"
              >
                <ArrowLeft className="w-4 h-4" />
                {siteName}
              </Link>
            </div>
            <span className="font-['JetBrains_Mono'] text-xs text-muted-foreground">
              docs
            </span>
          </header>

          <div className="mx-auto w-full max-w-3xl flex-1 px-6 py-10">
            <h1 className="mb-6 text-3xl font-['Playfair_Display'] text-white sm:text-4xl">
              {current.title}
            </h1>
            <article className="doc-md">
              <Streamdown
                mode="static"
                plugins={{ code: codePlugin }}
                shikiTheme={["gruvbox-dark-hard", "gruvbox-dark-hard"]}
                controls={{ code: { download: false } }}
              >
                {current.body}
              </Streamdown>
              {current.id === "themes" && (
                <div className="mt-6">
                  <ThemeJsonEditor />
                </div>
              )}
            </article>

            {/* Prev / Next */}
            <nav className="mt-12 flex gap-4 border-t border-border pt-6">
              {prev ? (
                <button
                  onClick={() => setActive(prev.id)}
                  className="group flex-1 rounded-sm border border-border px-4 py-3 text-left transition-colors hover:border-primary/50"
                >
                  <span className="mb-1 block font-['JetBrains_Mono'] text-xs text-muted-foreground">
                    ← Previous
                  </span>
                  <span className="text-sm text-foreground transition-colors group-hover:text-primary">
                    {prev.title}
                  </span>
                </button>
              ) : (
                <span className="flex-1" />
              )}
              {next ? (
                <button
                  onClick={() => setActive(next.id)}
                  className="group flex-1 rounded-sm border border-border px-4 py-3 text-right transition-colors hover:border-primary/50"
                >
                  <span className="mb-1 block font-['JetBrains_Mono'] text-xs text-muted-foreground">
                    Next →
                  </span>
                  <span className="text-sm text-foreground transition-colors group-hover:text-primary">
                    {next.title}
                  </span>
                </button>
              ) : (
                <span className="flex-1" />
              )}
            </nav>
          </div>
        </SidebarInset>
      </SidebarProvider>
    </div>
  );
}
