import {
  PromptInput,
  PromptInputBody,
  PromptInputFooter,
  PromptInputSubmit,
  PromptInputTextarea,
  PromptInputTools,
  usePromptInputController,
  type PromptInputMessage,
} from "@/components/ai-elements/prompt-input";
import {
  Command,
  CommandEmpty,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import { api, type FileResult } from "@/lib/api";
import type { ChatStatus } from "ai";
import { useEffect, useState } from "react";
import type { KeyboardEvent as ReactKeyboardEvent, ReactNode } from "react";
import { FileIcon } from "lucide-react";

export type ChatCommand = "new" | "models" | "sessions";

const slashCommands: {
  id: ChatCommand;
  name: string;
  description: string;
}[] = [
    { id: "new", name: "/new", description: "Start a new chat" },
    { id: "models", name: "/models", description: "Switch model" },
    { id: "sessions", name: "/sessions", description: "Search sessions" },
  ];

export function PromptComposer({
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
              className="p-4 rounded-sm"
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
