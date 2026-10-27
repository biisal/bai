import { Plus } from "lucide-react";

const faqs = [
  {
    q: "What is bAI?",
    a: "A terminal-based AI coding agent. It ships as a single binary with a TUI, and it can read, edit and write files and run shell commands for you.",
  },
  {
    q: "Do I need an extra subscription?",
    a: "No subscription from us. You need an API key for at least one supported provider, which goes in ~/.config/bai/config.json.",
  },
  {
    q: "Which providers and models can I use?",
    a: "OpenAI, Anthropic and Groq out of the box, plus any OpenAI-compatible endpoint you add in config. The example config also includes DeepSeek, Qwen and OpenCode.",
  },
  {
    q: "Where is my data stored?",
    a: "Sessions are saved in a local SQLite file, ~/.config/bai/bai.db by default. Your prompts and the code the agent reads go to the provider you configured.",
  },
  {
    q: "Will it mess up my git history?",
    a: "No. bAI keeps its own history in a private .bai_git directory and commits there after each turn, so your repository history is untouched.",
  },
  {
    q: "Can I add my own tools and skills?",
    a: "Yes. Custom tools load from a manifest, skills load from SKILL.md files in the directories you list, and project instructions are read from AGENTS.md.",
  },
  {
    q: "Is it ready for production use?",
    a: "Not yet. bAI is in early development (v0.0.x). The core chat loop and coding tools work, but expect rough edges.",
  },
];

export default function FAQ() {
  return (
    <section className="mx-auto w-full max-w-3xl px-6 py-24">
      <h2 className="mb-10 text-center font-['Playfair_Display'] text-4xl text-white md:text-5xl">
        FAQ
      </h2>
      <div className="divide-y divide-border border-y border-border">
        {faqs.map((f) => (
          <details key={f.q} className="group py-5">
            <summary className="flex cursor-pointer list-none items-center justify-between gap-4 text-white [&::-webkit-details-marker]:hidden">
              {f.q}
              <Plus className="h-4 w-4 shrink-0 text-primary transition-transform group-open:rotate-45" />
            </summary>
            <p className="mt-3 text-sm leading-relaxed text-muted-foreground">
              {f.a}
            </p>
          </details>
        ))}
      </div>
    </section>
  );
}
