import {
  Message,
  MessageContent,
  MessageResponse,
} from "@/components/ai-elements/message";
import {
  Reasoning,
  ReasoningContent,
  ReasoningTrigger,
} from "@/components/ai-elements/reasoning";
import {
  Tool,
  ToolContent,
  ToolHeader,
  ToolInput,
  ToolOutput,
  type ToolPart,
} from "@/components/ai-elements/tool";
import type { AppUIMessage, UsageData } from "@/lib/api";
import { cn } from "@/lib/utils";
import { useState } from "react";
import type { ReactNode } from "react";
import { CheckIcon, CopyIcon, FileIcon, RefreshCwIcon } from "lucide-react";

export function ChatMessage({
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
