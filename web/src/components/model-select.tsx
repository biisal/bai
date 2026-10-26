import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { Models } from "@/lib/api";

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
  const items =
    models?.providers.flatMap((provider) =>
      provider.models.map((model) => ({
        label: model,
        value: `${provider.id}:${model}`,
      }))
    ) ?? [];

  return (
    <Select
      disabled={disabled || models === null}
      items={items}
      onValueChange={(next) => {
        if (next) {
          onChange(next);
        }
      }}
      value={value || null}
    >
      <SelectTrigger
        aria-label="Model"
        className="h-8 max-w-64 border-0 bg-transparent hover:bg-muted focus-visible:ring-2"
      >
        <SelectValue placeholder="Loading models..." />
      </SelectTrigger>
      <SelectContent align="end" className="max-h-96">
        {models?.providers.map((provider) => (
          <SelectGroup key={provider.id}>
            <SelectLabel>{provider.name}</SelectLabel>
            {provider.models.map((model) => (
              <SelectItem
                key={`${provider.id}:${model}`}
                value={`${provider.id}:${model}`}
              >
                {model}
              </SelectItem>
            ))}
          </SelectGroup>
        ))}
      </SelectContent>
    </Select>
  );
}
