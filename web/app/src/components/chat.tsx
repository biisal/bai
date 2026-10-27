import {
  Conversation,
  ConversationContent,
  ConversationScrollButton,
} from "@/components/ai-elements/conversation";
import {
  Message,
  MessageContent,
  MessageResponse,
} from "@/components/ai-elements/message";
import {
  PromptInput,
  PromptInputBody,
  PromptInputFooter,
  PromptInputProvider,
  PromptInputSubmit,
  PromptInputTextarea,
  PromptInputTools,
  usePromptInputController,
  type PromptInputMessage,
} from "@/components/ai-elements/prompt-input";
import {
  Reasoning,
  ReasoningContent,
  ReasoningTrigger,
} from "@/components/ai-elements/reasoning";
import {
  Command,
  CommandEmpty,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import {
  Tool,
  ToolContent,
  ToolHeader,
  ToolInput,
  ToolOutput,
  type ToolPart,
} from "@/components/ai-elements/tool";
import { api, errorText, type AppUIMessage, type FileResult } from "@/lib/api";
import type { UsageData } from "@/lib/api";
import { DefaultChatTransport, type ChatStatus } from "ai";
import { useChat } from "@ai-sdk/react";
import { useEffect, useMemo, useRef, useState } from "react";
import type { KeyboardEvent as ReactKeyboardEvent, ReactNode } from "react";
import { cn } from "@/lib/utils";
import { CheckIcon, CopyIcon, FileIcon, RefreshCwIcon } from "lucide-react";

export type ChatCommand = "new" | "models" | "sessions";

type ChatViewProps = {
  conversationId: string | null;
  directory: string;
  initialResume: boolean;
  notify: boolean;
  onCommand: (command: ChatCommand) => void;
  onConversation: (id: string) => void;
  onRunEnd: () => void;
};

export function ChatView({
  conversationId,
  directory,
  initialResume,
  notify,
  onCommand,
  onConversation,
  onRunEnd,
}: ChatViewProps) {
  const [messages, setMessages] = useState<AppUIMessage[] | null>(
    conversationId ? null : []
  );
  const [loadError, setLoadError] = useState<string | null>(null);

  useEffect(() => {
    if (!conversationId) {
      setMessages([]);
      return;
    }
    let cancelled = false;
    setMessages(null);
    setLoadError(null);
    api
      .messages(conversationId)
      .then((result) => {
        if (!cancelled) {
          setMessages(result);
        }
      })
      .catch((error: unknown) => {
        if (!cancelled) {
          setLoadError(errorText(error));
        }
      });
    return () => {
      cancelled = true;
    };
  }, [conversationId]);

  if (loadError) {
    return (
      <div className="grid flex-1 place-items-center p-6">
        <p className="max-w-md text-center text-sm text-muted-foreground">
          {loadError}
        </p>
      </div>
    );
  }

  if (messages === null) {
    return (
      <div className="grid flex-1 place-items-center p-6">
        <p className="text-sm text-muted-foreground">Loading conversation...</p>
      </div>
    );
  }

  return (
    <ChatSession
      conversationId={conversationId}
      directory={directory}
      initialMessages={messages}
      initialResume={initialResume}
      notify={notify}
      onCommand={onCommand}
      onConversation={onConversation}
      onRunEnd={onRunEnd}
    />
  );
}

type ChatSessionProps = {
  conversationId: string | null;
  directory: string;
  initialMessages: AppUIMessage[];
  initialResume: boolean;
  notify: boolean;
  onCommand: (command: ChatCommand) => void;
  onConversation: (id: string) => void;
  onRunEnd: () => void;
};

function ChatSession({
  conversationId,
  directory,
  initialMessages,
  initialResume,
  notify,
  onCommand,
  onConversation,
  onRunEnd,
}: ChatSessionProps) {
  const idRef = useRef(conversationId ?? "");
  const [stopped, setStopped] = useState(false);
  const [usage, setUsage] = useState<UsageData | null>(null);

  const transport = useMemo(
    () =>
      new DefaultChatTransport<AppUIMessage>({
        api: "/api/chat",
        prepareSendMessagesRequest: ({ messages }) => ({
          body: {
            id: idRef.current || undefined,
            message: messages.at(-1),
          },
        }),
      }),
    []
  );

  const { messages, sendMessage, status, error, clearError, stop } =
    useChat<AppUIMessage>({
      id: conversationId ?? "new",
      messages: initialMessages,
      transport,
      resume: initialResume,
      onData: (part) => {
        if (part.type === "data-usage") {
          setUsage(part.data);
          return;
        }
        if (part.type !== "data-conversation") {
          return;
        }
        const id = String(part.data.id);
        idRef.current = id;
        history.replaceState({}, "", `/c/${id}`);
        onConversation(id);
      },
      onFinish: () => {
        onRunEnd();
        if (
          notify &&
          document.hidden &&
          "Notification" in window &&
          Notification.permission === "granted"
        ) {
          new Notification("bai", { body: "Response ready." });
        }
      },
    });

  const streaming = status === "submitted" || status === "streaming";

  const handleSubmit = (message: PromptInputMessage) => {
    const text = message.text.trim();
    if (!text || streaming) {
      return;
    }
    clearError();
    setStopped(false);
    setUsage(null);
    void sendMessage({ text });
  };

  const regenerate = (text: string) => {
    if (streaming || !text.trim()) {
      return;
    }
    clearError();
    setStopped(false);
    setUsage(null);
    void sendMessage({ text });
  };

  const lastUserText =
    [...messages]
      .reverse()
      .find((message) => message.role === "user")
      ?.parts.filter((part) => part.type === "text")
      .map((part) => part.text)
      .join("\n") ?? "";
  let lastAssistantIndex = -1;
  for (let index = messages.length - 1; index >= 0; index--) {
    if (messages[index]?.role === "assistant") {
      lastAssistantIndex = index;
      break;
    }
  }

  const handleStop = () => {
    if (!streaming) {
      return;
    }
    const id = idRef.current;
    void (async () => {
      try {
        if (id) {
          await api.stop(id);
        }
      } finally {
        stop();
        setStopped(true);
      }
    })();
  };

  const errorNotice = error ? (
    <div className="mx-auto w-full max-w-3xl px-3 pb-2" role="alert">
      <div className="flex items-start justify-between gap-3 rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
        <span>{errorText(error)}</span>
        <button
          className="shrink-0 rounded px-1 text-xs underline underline-offset-2 focus-visible:ring-2 focus-visible:ring-ring"
          onClick={clearError}
          type="button"
        >
          Dismiss
        </button>
      </div>
    </div>
  ) : null;

  return (
    <PromptInputProvider>
      <div className="flex min-h-0 flex-1 flex-col">
        {messages.length === 0 ? (
          <div className="min-h-0 flex-1 overflow-y-auto">
            <div className="mx-auto flex min-h-full w-full max-w-3xl flex-col justify-center gap-8 px-4 py-10">
              <EmptyState directory={directory} />
              {errorNotice}
              <PromptComposer
                onCommand={onCommand}
                onStop={handleStop}
                onSubmit={handleSubmit}
                status={status}
                stopped={stopped}
              />
            </div>
          </div>
        ) : (
          <>
            <Conversation>
              <ConversationContent className="mx-auto w-full max-w-3xl gap-6 px-4 py-6">
                {messages.map((message, index) => (
                  <ChatMessage
                    canRegenerate={
                      !streaming &&
                      index === lastAssistantIndex &&
                      lastUserText.length > 0
                    }
                    key={message.id || String(index)}
                    message={message}
                    onRegenerate={() => regenerate(lastUserText)}
                    usage={index === lastAssistantIndex ? usage : null}
                  />
                ))}
                {stopped && !streaming && (
                  <p
                    className="text-center text-xs text-muted-foreground"
                    role="status"
                  >
                    Response stopped.
                  </p>
                )}
              </ConversationContent>
              <ConversationScrollButton />
            </Conversation>

            {errorNotice}

            <PromptComposer
              onCommand={onCommand}
              onStop={handleStop}
              onSubmit={handleSubmit}
              status={status}
              stopped={stopped}
            />
          </>
        )}

        <p aria-live="polite" className="sr-only" role="status">
          {status === "submitted"
            ? "Waiting for the assistant."
            : status === "streaming"
              ? "Assistant is responding."
              : stopped
                ? "Response stopped."
                : ""}
        </p>
      </div>
    </PromptInputProvider>
  );
}

const slashCommands: {
  id: ChatCommand;
  name: string;
  description: string;
}[] = [
  { id: "new", name: "/new", description: "Start a new chat" },
  { id: "models", name: "/models", description: "Switch model" },
  { id: "sessions", name: "/sessions", description: "Search sessions" },
];

function PromptComposer({
  status,
  stopped,
  onSubmit,
  onStop,
  onCommand,
}: {
  status: ChatStatus;
  stopped: boolean;
  onSubmit: (message: PromptInputMessage) => void;
  onStop: () => void;
  onCommand: (command: ChatCommand) => void;
}) {
  const { textInput } = usePromptInputController();
  const value = textInput.value;
  const [files, setFiles] = useState<FileResult[]>([]);
  const [fileLoading, setFileLoading] = useState(false);
  const [activeIndex, setActiveIndex] = useState(0);
  const [dismissed, setDismissed] = useState("");
  const slashMatch = /^\/([^\s]*)$/.exec(value);
  const slashOpen = slashMatch !== null && dismissed !== value;
  const mentionMatch = /@([^\s]*)$/.exec(value);
  const mentionOpen =
    !slashOpen && mentionMatch !== null && dismissed !== value;
  const mentionQuery = mentionOpen ? (mentionMatch?.[1] ?? "") : null;
  const slashResults = slashOpen
    ? slashCommands.filter((command) =>
        command.name.slice(1).startsWith((slashMatch?.[1] ?? "").toLowerCase())
      )
    : [];
  const busy = status === "submitted" || status === "streaming";

  useEffect(() => {
    setActiveIndex(0);
  }, [mentionQuery, slashMatch?.[1]]);

  useEffect(() => {
    if (mentionQuery === null) {
      setFiles([]);
      setFileLoading(false);
      return;
    }
    const controller = new AbortController();
    const timer = setTimeout(() => {
      setFileLoading(true);
      api
        .files(mentionQuery)
        .then((result) => {
          if (!controller.signal.aborted) {
            setFiles(result);
          }
        })
        .catch(() => {
          if (!controller.signal.aborted) {
            setFiles([]);
          }
        })
        .finally(() => {
          if (!controller.signal.aborted) {
            setFileLoading(false);
          }
        });
    }, 150);
    return () => {
      clearTimeout(timer);
      controller.abort();
    };
  }, [mentionQuery]);

  const pickMention = (path: string) => {
    textInput.setInput(value.replace(/@[^\s]*$/, `${JSON.stringify(path)} `));
    setFiles([]);
    setDismissed("");
    requestAnimationFrame(() => {
      const input = document.getElementById("prompt-input");
      if (input instanceof HTMLTextAreaElement) {
        input.focus();
        input.setSelectionRange(input.value.length, input.value.length);
      }
    });
  };

  const pickSlash = (command: ChatCommand) => {
    textInput.clear();
    setDismissed("");
    onCommand(command);
  };

  const handleKeyDown = (event: ReactKeyboardEvent<HTMLTextAreaElement>) => {
    if (slashOpen && slashResults.length > 0) {
      if (event.key === "ArrowDown" || event.key === "ArrowUp") {
        event.preventDefault();
        setActiveIndex((index) =>
          event.key === "ArrowDown"
            ? (index + 1) % slashResults.length
            : (index - 1 + slashResults.length) % slashResults.length
        );
        return;
      }
      if (event.key === "Enter" || event.key === "Tab") {
        event.preventDefault();
        const command = slashResults[activeIndex];
        if (command) {
          pickSlash(command.id);
        }
        return;
      }
      if (event.key === "Escape") {
        event.preventDefault();
        setDismissed(value);
        return;
      }
    }

    if (mentionOpen) {
      if (event.key === "ArrowDown" || event.key === "ArrowUp") {
        if (files.length > 0) {
          event.preventDefault();
          setActiveIndex((index) =>
            event.key === "ArrowDown"
              ? (index + 1) % files.length
              : (index - 1 + files.length) % files.length
          );
        }
        return;
      }
      if ((event.key === "Enter" || event.key === "Tab") && files.length > 0) {
        event.preventDefault();
        const file = files[activeIndex];
        if (file) {
          pickMention(file.path);
        }
        return;
      }
      if (event.key === "Escape") {
        event.preventDefault();
        setDismissed(value);
        return;
      }
    }

    if (event.key === "Escape" && value === "" && busy) {
      event.preventDefault();
      onStop();
    }
  };

  return (
    <div className="px-3 pb-3 md:px-4 md:pb-4">
      <div className="relative mx-auto w-full max-w-3xl">
        {slashOpen && (
          <Menu
            onValueChange={(next) => {
              const index = slashResults.findIndex(
                (command) => command.id === next
              );
              if (index >= 0) {
                setActiveIndex(index);
              }
            }}
            value={slashResults[activeIndex]?.id ?? slashResults[0]?.id ?? ""}
          >
            {slashResults.length === 0 ? (
              <CommandEmpty>No matching command</CommandEmpty>
            ) : (
              slashResults.map((command) => (
                <CommandItem
                  key={command.id}
                  onSelect={() => pickSlash(command.id)}
                  value={command.id}
                >
                  <span className="font-mono text-xs">{command.name}</span>
                  <span className="text-xs text-muted-foreground">
                    {command.description}
                  </span>
                </CommandItem>
              ))
            )}
          </Menu>
        )}

        {mentionOpen && (
          <Menu
            onValueChange={(next) => {
              const index = files.findIndex((file) => file.path === next);
              if (index >= 0) {
                setActiveIndex(index);
              }
            }}
            value={files[activeIndex]?.path ?? files[0]?.path ?? ""}
          >
            {files.length === 0 ? (
              <CommandEmpty>
                {fileLoading ? "Looking for files..." : "No files found"}
              </CommandEmpty>
            ) : (
              files.map((file) => (
                <CommandItem
                  key={file.path}
                  onSelect={() => pickMention(file.path)}
                  value={file.path}
                >
                  <FileIcon />
                  <span className="min-w-0 truncate font-mono text-xs">
                    {file.path}
                  </span>
                </CommandItem>
              ))
            )}
          </Menu>
        )}

        <PromptInput
          className="[&_[data-slot=input-group]]:rounded-3xl [&_[data-slot=input-group]]:border-0 [&_[data-slot=input-group]]:bg-muted/50 [&_[data-slot=input-group]]:shadow-xs dark:[&_[data-slot=input-group]]:bg-muted/50"
          onSubmit={onSubmit}
        >
          <PromptInputBody>
            <PromptInputTextarea
              id="prompt-input"
              onKeyDown={handleKeyDown}
              placeholder="Ask bai anything..."
            />
          </PromptInputBody>
          <PromptInputFooter>
            <PromptInputTools>
              <span className="px-1 text-xs text-muted-foreground">
                {status === "submitted"
                  ? "Thinking..."
                  : status === "streaming"
                    ? "Responding..."
                    : stopped
                      ? "Stopped"
                      : "@ files · / commands"}
              </span>
            </PromptInputTools>
            <PromptInputSubmit onStop={onStop} status={status} />
          </PromptInputFooter>
        </PromptInput>
      </div>
    </div>
  );
}

function Menu({
  value,
  onValueChange,
  children,
}: {
  value: string;
  onValueChange: (value: string) => void;
  children: ReactNode;
}) {
  return (
    <div className="absolute inset-x-0 bottom-full z-20 mb-2 overflow-hidden rounded-xl bg-popover shadow-md ring-1 ring-foreground/10">
      <Command onValueChange={onValueChange} shouldFilter={false} value={value}>
        <CommandList className="max-h-64">{children}</CommandList>
      </Command>
    </div>
  );
}

function greeting(): string {
  const hour = new Date().getHours();
  if (hour >= 6 && hour < 12) {
    return "Good morning";
  }
  if (hour >= 12 && hour < 17) {
    return "Good afternoon";
  }
  if (hour >= 17 && hour < 21) {
    return "Good evening";
  }
  return "Hello, night owl";
}

function EmptyState({ directory }: { directory: string }) {
  const folder = directory.split(/[\\/]/).filter(Boolean).at(-1) ?? directory;

  return (
    <div className="text-center">
      <h1 className="text-3xl font-semibold tracking-tight">{greeting()}</h1>
      <p className="mt-2 text-sm text-muted-foreground">
        What should we explore in{" "}
        <span className="font-medium text-foreground">{folder}</span>?
      </p>
    </div>
  );
}

function ChatMessage({
  message,
  usage,
  canRegenerate,
  onRegenerate,
}: {
  message: AppUIMessage;
  usage: UsageData | null;
  canRegenerate: boolean;
  onRegenerate: () => void;
}) {
  const text = message.parts
    .filter((part) => part.type === "text")
    .map((part) => part.text)
    .join("\n");

  return (
    <Message from={message.role}>
      <MessageContent>
        {message.parts.map((part, index) => {
          const key = `${part.type}-${index}`;

          if (part.type === "text") {
            if (message.role === "user") {
              return <UserText key={key} text={part.text} />;
            }
            return (
              <MessageResponse
                isAnimating={part.state === "streaming"}
                key={key}
              >
                {part.text}
              </MessageResponse>
            );
          }

          if (part.type === "reasoning") {
            return (
              <Reasoning
                defaultOpen={part.state === "streaming"}
                isStreaming={part.state === "streaming"}
                key={key}
              >
                <ReasoningTrigger />
                <ReasoningContent>{part.text}</ReasoningContent>
              </Reasoning>
            );
          }

          if (isToolPart(part)) {
            return <ToolCard key={part.toolCallId} part={part} />;
          }

          return null;
        })}
      </MessageContent>

      {text && (
        <div
          className={cn(
            "flex items-center gap-1 text-muted-foreground md:opacity-0 md:transition-opacity md:group-hover:opacity-100",
            message.role === "user" ? "justify-end" : "justify-start"
          )}
        >
          <CopyMessage text={text} />
          {canRegenerate && (
            <button
              aria-label="Regenerate response"
              className="flex items-center gap-1 rounded-md px-1.5 py-1 text-[11px] hover:bg-muted hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
              onClick={onRegenerate}
              type="button"
            >
              <RefreshCwIcon className="size-3" />
              Regenerate
            </button>
          )}
        </div>
      )}

      {usage && (
        <p className="text-[11px] text-muted-foreground">
          {usage.inputTokens.toLocaleString()} in ·{" "}
          {usage.outputTokens.toLocaleString()} out ·{" "}
          {usage.totalTokens.toLocaleString()} total ·{" "}
          {(usage.elapsedMs / 1000).toFixed(1)}s
        </p>
      )}
    </Message>
  );
}

function CopyMessage({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);
  return (
    <button
      aria-label="Copy message"
      className="flex items-center gap-1 rounded-md px-1.5 py-1 text-[11px] hover:bg-muted hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
      onClick={() => {
        void navigator.clipboard.writeText(text).then(() => {
          setCopied(true);
          setTimeout(() => setCopied(false), 1500);
        });
      }}
      type="button"
    >
      {copied ? (
        <CheckIcon className="size-3" />
      ) : (
        <CopyIcon className="size-3" />
      )}
      {copied ? "Copied" : "Copy"}
    </button>
  );
}

function UserText({ text }: { text: string }) {
  const nodes: ReactNode[] = [];
  const pattern = /"([^"\n]+)"/g;
  let last = 0;
  let match: RegExpExecArray | null;
  while ((match = pattern.exec(text)) !== null) {
    const value = match[1] ?? "";
    const isPath = /[\\/]/.test(value) || /\.[a-z0-9]{1,8}$/i.test(value);
    if (!isPath) {
      continue;
    }
    if (match.index > last) {
      nodes.push(text.slice(last, match.index));
    }
    nodes.push(
      <span
        className="inline-flex max-w-full items-center gap-1 rounded bg-background/70 px-1.5 py-0.5 align-baseline font-mono text-[11px]"
        key={`${match.index}-${value}`}
      >
        <FileIcon className="size-3 shrink-0" />
        <span className="truncate">{value}</span>
      </span>
    );
    nodes.push(" ");
    last = match.index + match[0].length;
  }
  if (last > 0) {
    nodes.push(text.slice(last));
  }

  return (
    <p className="whitespace-pre-wrap break-words text-sm">
      {last > 0 ? nodes : text}
    </p>
  );
}

function isToolPart(part: AppUIMessage["parts"][number]): part is ToolPart {
  return part.type === "dynamic-tool" || part.type.startsWith("tool-");
}

function ToolCard({ part }: { part: ToolPart }) {
  return (
    <Tool
      defaultOpen={
        part.state === "input-streaming" || part.state === "input-available"
      }
    >
      {part.type === "dynamic-tool" ? (
        <ToolHeader
          state={part.state}
          toolName={part.toolName}
          type={part.type}
        />
      ) : (
        <ToolHeader state={part.state} type={part.type} />
      )}
      <ToolContent>
        <ToolInput input={part.input} />
        {part.state === "output-available" && (
          <ToolOutput errorText={undefined} output={part.output} />
        )}
        {part.state === "output-error" && (
          <ToolOutput errorText={part.errorText} output={undefined} />
        )}
      </ToolContent>
    </Tool>
  );
}
