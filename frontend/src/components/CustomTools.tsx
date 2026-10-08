import { useEffect, useRef, useState } from "react";
import { Link } from "react-router-dom";
import TerminalHeader from "./ui/terminal-header";

type Line = {
  id: number;
  kind: "user" | "reasoning" | "read" | "write" | "bash" | "assistant";
  text: string;
};

const PROMPT = "make a tool that shows the weather for a city";
const DIR = "~/.config/bai/tools";

type Step = { kind: Exclude<Line["kind"], "user">; text: string };
const STEPS: Step[] = [
  {
    kind: "reasoning",
    text: "A custom tool is an executable script plus one manifest entry. I will read the tools_maker skill, then check what already exists.",
  },
  { kind: "read", text: "internal:skills:tools_maker" },
  { kind: "bash", text: `ls -la ${DIR}` },
  { kind: "write", text: `${DIR}/weather.sh` },
  { kind: "bash", text: `chmod +x ${DIR}/weather.sh` },
  { kind: "write", text: `${DIR}/tools.json` },
  { kind: "bash", text: `${DIR}/weather.sh city=Patna` },
  {
    kind: "assistant",
    text: "Done. The tool is ready. I will use this to find the weather for a city.",
  },
];

const finalLines: Line[] = [
  { id: 1, kind: "user", text: PROMPT },
  ...STEPS.map((s, i) => ({ id: 2 + i, ...s })),
];

const steps = [
  {
    title: "Describe it",
    body: "Tell bAI in plain words what the tool should do.",
  },
  {
    title: "bAI builds it",
    body: "It writes an executable script, adds one entry to your tools manifest, and runs the script to check that it works.",
  },
];

const sleep = (ms: number) =>
  new Promise<void>((resolve) => setTimeout(resolve, ms));

export default function CustomTools() {
  const sectionRef = useRef<HTMLElement>(null);
  const chatRef = useRef<HTMLDivElement>(null);
  const [visible, setVisible] = useState(false);
  const [reduced, setReduced] = useState(false);
  const [input, setInput] = useState("");
  const [lines, setLines] = useState<Line[]>([]);
  const [working, setWorking] = useState(false);

  // Follow the newest line once the chat overflows, like a web chat.
  useEffect(() => {
    const el = chatRef.current;
    if (el) el.scrollTop = el.scrollHeight;
  }, [lines, working]);

  // Only animate while the section is on screen.
  useEffect(() => {
    setReduced(window.matchMedia("(prefers-reduced-motion: reduce)").matches);
    const el = sectionRef.current;
    if (!el) return;
    const io = new IntersectionObserver(
      ([entry]) => setVisible(entry.isIntersecting),
      { threshold: 0.3 }
    );
    io.observe(el);
    return () => io.disconnect();
  }, []);

  useEffect(() => {
    if (reduced) {
      setLines(finalLines);
      setInput("");
      setWorking(false);
      return;
    }
    if (!visible) return;

    let cancelled = false;
    let nextId = 1;
    const wait = async (ms: number) => {
      await sleep(ms);
      if (cancelled) throw new Error("cancelled");
    };
    const push = (kind: Line["kind"], text: string) => {
      const id = nextId++;
      setLines((l) => [...l, { id, kind, text }]);
      return id;
    };
    const stream = async (kind: Line["kind"], text: string) => {
      const id = push(kind, "");
      for (let i = 3; i < text.length + 3; i += 3) {
        setLines((l) =>
          l.map((x) => (x.id === id ? { ...x, text: text.slice(0, i) } : x))
        );
        await wait(20);
      }
    };

    const run = async () => {
      while (!cancelled) {
        setLines([]);
        setInput("");
        setWorking(false);
        nextId = 1;
        await wait(1200);

        for (let i = 1; i <= PROMPT.length; i++) {
          setInput(PROMPT.slice(0, i));
          await wait(40 + Math.random() * 40);
        }
        await wait(500);

        setInput("");
        push("user", PROMPT);
        setWorking(true);
        await wait(900);

        for (const step of STEPS) {
          if (step.kind === "assistant") setWorking(false);
          if (step.kind === "reasoning" || step.kind === "assistant") {
            await stream(step.kind, step.text);
            await wait(500);
          } else {
            push(step.kind, step.text);
            await wait(900);
          }
        }

        await wait(5500);
      }
    };

    run().catch(() => {});
    return () => {
      cancelled = true;
    };
  }, [visible, reduced]);

  return (
    <section
      ref={sectionRef}
      id="tools"
      className="mx-auto w-full max-w-7xl px-6 py-24"
    >
      <div className="grid items-center gap-12 lg:grid-cols-2 lg:gap-16">
        {/* Explanation */}
        <div>
          <h2 className="mb-5 font-['Playfair_Display'] text-4xl leading-[1.15] text-white md:text-5xl">
            Need a tool?
            <br />
            Just ask for one.
          </h2>
          <p className="mb-10 max-w-lg text-muted-foreground">
            bAI ships with four built-in tools. For everything else, tell it
            what you want and it builds the tool for you.
          </p>

          <ol className="mb-10 space-y-6">
            {steps.map((s, i) => (
              <li key={s.title} className="flex gap-4">
                <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-sm border border-border font-['JetBrains_Mono'] text-xs text-primary">
                  {i + 1}
                </span>
                <div>
                  <h3 className="mb-1 text-white">{s.title}</h3>
                  <p className="text-sm leading-relaxed text-muted-foreground">
                    {s.body}
                  </p>
                </div>
              </li>
            ))}
          </ol>

          <Link
            to="/docs"
            className="text-sm text-primary transition-opacity hover:opacity-80"
          >
            Read how custom tools work &rarr;
          </Link>
        </div>

        {/* Animated terminal */}
        <div className="overflow-hidden rounded-sm border border-border bg-[#1c1713] shadow-2xl">
          <TerminalHeader title="bai" />

          <p className="sr-only">
            Animated preview: the user asks bAI to make a weather tool. bAI
            reads its tools_maker skill, writes a script, registers it in
            tools.json, tests it and tells the user to restart bAI.
          </p>

          <div
            aria-hidden
            className="flex h-[30rem] flex-col font-['JetBrains_Mono'] text-sm"
          >
            <div className="flex min-h-0 flex-1 flex-col overflow-hidden pt-5">
              <p className="shrink-0 px-5 text-[#8a5a5a]">
                ~/projects/my-app (main)
              </p>
              <div
                ref={chatRef}
                className="pointer-events-none mt-4 min-h-0 flex-1 space-y-4 overflow-hidden pb-3"
              >
                {lines.map((line) => (
                  <TranscriptLine key={line.id} line={line} />
                ))}
                {working && (
                  <p className="px-5 text-[#8a5a5a]">
                    <span className="mr-2 inline-block animate-pulse">⠶</span>
                    working...
                  </p>
                )}
              </div>
            </div>

            <div className="shrink-0 border-t border-border/80 px-5 py-3">
              <div className="flex items-center border-l border-white/70 pl-3">
                {input ? (
                  <>
                    <span className="truncate text-white">{input}</span>
                    <span className="ml-0.5 inline-block h-4 w-2 shrink-0 animate-pulse bg-white motion-reduce:animate-none" />
                  </>
                ) : (
                  <span className="text-muted-foreground/70">
                    <span className="animate-pulse bg-white text-[#1c1713] motion-reduce:animate-none">
                      S
                    </span>
                    end a message...
                  </span>
                )}
              </div>
            </div>
            <div className="shrink-0 border-t border-border/80 px-5 py-3 text-[#8a5a5a]">
              Anthropic - claude-fable-5-1
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}

function TranscriptLine({ line }: { line: Line }) {
  switch (line.kind) {
    case "user":
      return <p className="bg-[#26221e] px-5 py-5 text-white">{line.text}</p>;
    case "reasoning":
      return <p className="px-5 leading-relaxed text-[#8a5a5a]">{line.text}</p>;
    case "read":
      return <p className="break-all px-5 text-success">{line.text}</p>;
    case "write":
      return <p className="break-all px-5 text-warning">{line.text}</p>;
    case "bash":
      return (
        <p className="break-all px-5 font-bold text-[#e8742a]">$ {line.text}</p>
      );
    case "assistant":
      return (
        <p className="px-5 leading-relaxed text-foreground">{line.text}</p>
      );
  }
}
