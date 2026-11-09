import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
  CommandShortcut,
} from "@/components/ui/command";
import type { Conversation, Models } from "@/lib/api";
import type { SettingsTab } from "@/features/settings/settings-dialog";
import {
  GitBranchIcon,
  MessageSquareIcon,
  PlusIcon,
  SettingsIcon,
  SparklesIcon,
} from "lucide-react";

type CommandPaletteProps = {
  open: boolean;
  conversations: Conversation[];
  models: Models | null;
  onOpenChange: (open: boolean) => void;
  onNew: () => void;
  onSelect: (id: number) => void;
  onModel: (provider: string, model: string) => void;
  onSettings: (tab: SettingsTab) => void;
  onChanges: () => void;
};

export function CommandPalette({
  open,
  conversations,
  models,
  onOpenChange,
  onNew,
  onSelect,
  onModel,
  onSettings,
  onChanges,
}: CommandPaletteProps) {
  const run = (action: () => void) => {
    onOpenChange(false);
    action();
  };
  const modelItems =
    models?.providers.flatMap((provider) =>
      provider.models.map((model) => ({
        provider: provider.id,
        model,
        value: `${provider.id}:${model}`,
      }))
    ) ?? [];

  return (
    <CommandDialog
      description="New chat, sessions, models, settings and changes."
      onOpenChange={onOpenChange}
      open={open}
      title="Command palette"
    >
      <CommandInput placeholder="Type a command or search..." />
      <CommandList>
        <CommandEmpty>No results.</CommandEmpty>
        <CommandGroup heading="Actions">
          <CommandItem onSelect={() => run(onNew)} value="new chat">
            <PlusIcon />
            New chat
          </CommandItem>
          <CommandItem
            onSelect={() => run(() => onSettings("models"))}
            value="switch model"
          >
            <SparklesIcon />
            Switch model
          </CommandItem>
          <CommandItem
            onSelect={() => run(() => onSettings("appearance"))}
            value="open settings"
          >
            <SettingsIcon />
            Open settings
            <CommandShortcut>⌘,</CommandShortcut>
          </CommandItem>
          <CommandItem onSelect={() => run(onChanges)} value="open changes">
            <GitBranchIcon />
            Open changes
          </CommandItem>
        </CommandGroup>
        {conversations.length > 0 && (
          <>
            <CommandSeparator />
            <CommandGroup heading="Sessions">
              {conversations.slice(0, 20).map((conversation) => (
                <CommandItem
                  key={conversation.id}
                  onSelect={() => run(() => onSelect(conversation.id))}
                  value={`${conversation.title} ${conversation.id}`}
                >
                  <MessageSquareIcon />
                  {conversation.title || "Untitled"}
                </CommandItem>
              ))}
            </CommandGroup>
          </>
        )}
        {modelItems.length > 0 && (
          <>
            <CommandSeparator />
            <CommandGroup heading="Models">
              {modelItems.map((item) => (
                <CommandItem
                  key={item.value}
                  onSelect={() => run(() => onModel(item.provider, item.model))}
                  value={item.value}
                >
                  <SparklesIcon />
                  {item.model}
                  <CommandShortcut>{item.provider}</CommandShortcut>
                </CommandItem>
              ))}
            </CommandGroup>
          </>
        )}
      </CommandList>
    </CommandDialog>
  );
}
