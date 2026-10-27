import { Popover as PopoverPrimitive } from "@base-ui/react/popover";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import type { Models } from "@/lib/api";
import { ChevronsUpDownIcon } from "lucide-react";
import { useState } from "react";

type ModelSelectProps = {
  models: Models | null;
  value: string;
  disabled: boolean;
  onChange: (value: string) => void;
};

export function ModelSelect({
  models,
  value,
  disabled,
  onChange,
}: ModelSelectProps) {
  const [open, setOpen] = useState(false);
  const items =
    models?.providers.flatMap((provider) =>
      provider.models.map((model) => ({
        label: model,
        value: `${provider.id}:${model}`,
        provider: provider.name,
      }))
    ) ?? [];
  const selected = items.find((item) => item.value === value);

  return (
    <PopoverPrimitive.Root onOpenChange={setOpen} open={open}>
      <PopoverPrimitive.Trigger
        aria-label="Model"
        className="flex h-8 max-w-64 items-center gap-1.5 rounded-lg bg-transparent px-2 text-sm transition-colors outline-none hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
        disabled={disabled || models === null}
        type="button"
      >
        <span className="truncate">
          {models === null
            ? "Loading models..."
            : (selected?.label ?? "Select model")}
        </span>
        <ChevronsUpDownIcon className="size-4 shrink-0 text-muted-foreground" />
      </PopoverPrimitive.Trigger>
      <PopoverPrimitive.Portal>
        <PopoverPrimitive.Positioner
          align="end"
          className="isolate z-50"
          sideOffset={4}
        >
          <PopoverPrimitive.Popup className="w-72 origin-(--transform-origin) overflow-hidden rounded-xl bg-popover p-1 text-popover-foreground shadow-md ring-1 ring-foreground/10 outline-none data-[side=bottom]:slide-in-from-top-2 data-[side=top]:slide-in-from-bottom-2 data-open:animate-in data-open:fade-in-0 data-open:zoom-in-95 data-closed:animate-out data-closed:fade-out-0 data-closed:zoom-out-95">
            <Command>
              <CommandInput placeholder="Search models..." />
              <CommandList>
                <CommandEmpty>No models found.</CommandEmpty>
                {models?.providers.map((provider) => (
                  <CommandGroup heading={provider.name} key={provider.id}>
                    {provider.models.map((model) => {
                      const itemValue = `${provider.id}:${model}`;
                      const isSelected = itemValue === value;
                      return (
                        <CommandItem
                          data-checked={isSelected ? "true" : "false"}
                          disabled={isSelected}
                          key={itemValue}
                          keywords={[provider.name, provider.id]}
                          onSelect={() => {
                            onChange(itemValue);
                            setOpen(false);
                          }}
                          value={itemValue}
                        >
                          <span className="truncate">{model}</span>
                        </CommandItem>
                      );
                    })}
                  </CommandGroup>
                ))}
              </CommandList>
            </Command>
          </PopoverPrimitive.Popup>
        </PopoverPrimitive.Positioner>
      </PopoverPrimitive.Portal>
    </PopoverPrimitive.Root>
  );
}
