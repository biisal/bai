import type {
  CodeHighlighterPlugin,
  HighlightOptions,
  HighlightResult,
} from "streamdown";

const languages = [
  "go",
  "typescript",
  "tsx",
  "javascript",
  "jsx",
  "json",
  "bash",
  "python",
  "markdown",
  "yaml",
  "html",
  "css",
] as const;

const aliases: Record<string, string> = {
  ts: "typescript",
  js: "javascript",
  md: "markdown",
  yml: "yaml",
  sh: "bash",
  shell: "bash",
  py: "python",
  golang: "go",
};

const supported = new Set<string>(languages);
const cache = new Map<string, HighlightResult>();
const listeners = new Map<string, Set<(result: HighlightResult) => void>>();

let highlighterPromise: Promise<import("shiki/core").HighlighterCore> | null =
  null;

function loadHighlighter() {
  if (!highlighterPromise) {
    highlighterPromise = (async () => {
      const [{ createHighlighterCore }, { createJavaScriptRegexEngine }] =
        await Promise.all([
          import("shiki/core"),
          import("shiki/engine/javascript"),
        ]);
      const [light, dark, ...grammars] = await Promise.all([
        import("shiki/themes/github-light.mjs"),
        import("shiki/themes/github-dark.mjs"),
        import("shiki/langs/go.mjs"),
        import("shiki/langs/typescript.mjs"),
        import("shiki/langs/tsx.mjs"),
        import("shiki/langs/javascript.mjs"),
        import("shiki/langs/jsx.mjs"),
        import("shiki/langs/json.mjs"),
        import("shiki/langs/bash.mjs"),
        import("shiki/langs/python.mjs"),
        import("shiki/langs/markdown.mjs"),
        import("shiki/langs/yaml.mjs"),
        import("shiki/langs/html.mjs"),
        import("shiki/langs/css.mjs"),
      ]);
      return createHighlighterCore({
        engine: createJavaScriptRegexEngine(),
        langs: grammars.map((grammar) => grammar.default),
        themes: [light.default, dark.default],
      });
    })();
  }
  return highlighterPromise;
}

function normalize(language: string): string {
  const lower = language.trim().toLowerCase();
  return aliases[lower] ?? lower;
}

function cacheKey(options: HighlightOptions): string {
  const code = options.code;
  const start = code.slice(0, 100);
  const end = code.length > 100 ? code.slice(-100) : "";
  return `${normalize(options.language)}:${code.length}:${start}:${end}`;
}

function rawResult(code: string): HighlightResult {
  return {
    tokens: code.split("\n").map((line) => [{ content: line }]),
  };
}

export const code: CodeHighlighterPlugin = {
  name: "shiki",
  type: "code-highlighter",
  supportsLanguage(language) {
    return supported.has(normalize(language));
  },
  getSupportedLanguages() {
    return [...languages];
  },
  getThemes() {
    return ["github-light", "github-dark"];
  },
  highlight(options, callback) {
    const language = normalize(options.language);
    if (!supported.has(language)) {
      return rawResult(options.code);
    }

    const key = cacheKey(options);
    const cached = cache.get(key);
    if (cached) {
      return cached;
    }

    if (callback) {
      const pending = listeners.get(key) ?? new Set();
      pending.add(callback);
      listeners.set(key, pending);
    }

    void loadHighlighter()
      .then((highlighter) => {
        const result: HighlightResult = highlighter.codeToTokens(options.code, {
          lang: language,
          themes: { dark: "github-dark", light: "github-light" },
        });
        if (!options.isIncomplete) {
          cache.set(key, result);
        }
        const pending = listeners.get(key);
        if (pending) {
          for (const listener of pending) {
            listener(result);
          }
          listeners.delete(key);
        }
      })
      .catch((error: unknown) => {
        console.error("Failed to highlight code", error);
        listeners.delete(key);
      });

    return null;
  },
};
