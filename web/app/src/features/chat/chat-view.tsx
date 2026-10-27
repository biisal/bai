import {
  Conversation,
  ConversationContent,
  ConversationScrollButton,
} from "@/components/ai-elements/conversation";
import {
  PromptInputProvider,
  type PromptInputMessage,
} from "@/components/ai-elements/prompt-input";
import { ChatMessage } from "@/features/chat/chat-message";
import { EmptyState } from "@/features/chat/empty-state";
import {
  PromptComposer,
  type ChatCommand,
} from "@/features/chat/prompt-composer";
import { api, errorText, type AppUIMessage } from "@/lib/api";
import type { UsageData } from "@/lib/api";
import { DefaultChatTransport } from "ai";
import { useChat } from "@ai-sdk/react";
import { useEffect, useMemo, useRef, useState } from "react";

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

  return (
    <PromptInputProvider>
      <div className="flex min-h-0 flex-1 flex-col">
        {messages.length === 0 ? (
          <div className="min-h-0 flex-1 overflow-y-auto">
            <div className="mx-auto flex min-h-full w-full max-w-3xl flex-col justify-center gap-8 px-4 py-10">
              <EmptyState directory={directory} />
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
                {error && (
                  <p
                    className="text-sm break-words whitespace-pre-wrap text-destructive"
                    role="alert"
                  >
                    <span className="font-medium">Error: </span>
                    {errorText(error)}
                  </p>
                )}
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
