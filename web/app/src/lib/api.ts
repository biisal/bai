import type { UIMessage } from "ai";
import { z } from "zod";

const errorBody = z.object({ error: z.string() });

export type ConversationData = {
  id: number;
  title: string;
};

export type UsageData = {
  inputTokens: number;
  outputTokens: number;
  totalTokens: number;
  reasoningTokens: number;
  cacheCreationTokens: number;
  cacheReadTokens: number;
  elapsedMs: number;
};

export type AppUIMessage = UIMessage<
  never,
  { conversation: ConversationData; usage: UsageData }
>;

export type AppState = {
  directory: string;
  version: string;
  provider: string;
  model: string;
  running: { conversationId: number } | null;
};

export type Models = {
  providers: { id: string; name: string; models: string[] }[];
  provider: string;
  model: string;
};

export type Conversation = {
  id: number;
  title: string;
  createdAt: string;
};

export class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

async function readError(response: Response): Promise<string> {
  const text = await response.text();
  if (!text) {
    return response.statusText || `Request failed (${response.status})`;
  }
  try {
    const body = errorBody.safeParse(JSON.parse(text));
    if (body.success) {
      return body.data.error;
    }
  } catch {
    // not JSON; fall through to the raw text
  }
  return text;
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, init);
  if (!response.ok) {
    throw new ApiError(response.status, await readError(response));
  }
  if (response.status === 204) {
    return undefined as T;
  }
  return (await response.json()) as T;
}

export function errorText(error: unknown): string {
  const status =
    typeof error === "object" &&
    error !== null &&
    "statusCode" in error &&
    typeof error.statusCode === "number"
      ? error.statusCode
      : error instanceof ApiError
        ? error.status
        : undefined;

  if (status === 409) {
    return "A run is already active.";
  }
  if (error instanceof Error && error.message) {
    return error.message;
  }
  return "Something went wrong.";
}

export type FileResult = {
  path: string;
  isDir: boolean;
};

export type Settings = {
  models: {
    id: string;
    name: string;
    configured: boolean;
    models: { id: string; context: number; maxOutput: number }[];
  }[];
  active: { provider: string; model: string };
  git: { initialized: boolean; dirty: boolean; dir: string; hint: string };
  about: {
    version: string;
    directory: string;
    configPath: string;
    logPath: string;
    latestUpdate: string;
  };
};

export type SkillItem = {
  name: string;
  description: string;
  source: string;
};

export type PluginItem = SkillItem;

export type DiffResult = {
  diff: string;
  truncated: boolean;
};

export const api = {
  state: () => request<AppState>("/api/state"),
  models: () => request<Models>("/api/models"),
  settings: () => request<Settings>("/api/settings"),
  skills: () => request<SkillItem[]>("/api/skills"),
  plugins: () => request<PluginItem[]>("/api/plugins"),
  files: (query: string) =>
    request<FileResult[]>(`/api/files?q=${encodeURIComponent(query)}`),
  conversations: () => request<Conversation[]>("/api/conversations"),
  messages: (id: string) =>
    request<AppUIMessage[]>(
      `/api/conversations/${encodeURIComponent(id)}/messages`
    ),
  diff: (id: string) =>
    request<DiffResult>(`/api/conversations/${encodeURIComponent(id)}/diff`),
  setModel: (provider: string, model: string) =>
    request<void>("/api/model", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ provider, model }),
    }),
  renameConversation: (id: string, title: string) =>
    request<{ id: number; title: string }>(
      `/api/conversations/${encodeURIComponent(id)}`,
      {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ title }),
      }
    ),
  deleteConversation: (id: string) =>
    request<void>(`/api/conversations/${encodeURIComponent(id)}`, {
      method: "DELETE",
    }),
  stop: (id: string) =>
    request<void>(`/api/chat/${encodeURIComponent(id)}/stop`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: "{}",
    }),
};
