import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  api,
  errorText,
  type PluginItem,
  type Settings,
  type SkillItem,
} from "@/lib/api";
import type { Appearance, ThemeMode } from "@/lib/theme";
import { cn } from "@/lib/utils";
import { useEffect, useState } from "react";
import { toast } from "sonner";

export type SettingsTab = "models" | "appearance" | "skills" | "git" | "about";

type SettingsDialogProps = {
  open: boolean;
  initialTab: SettingsTab;
  appearance: Appearance;
  onOpenChange: (open: boolean) => void;
  onAppearanceChange: (appearance: Appearance) => void;
  onModelChange: (provider: string, model: string) => Promise<void>;
};

const tabs: { id: SettingsTab; label: string }[] = [
  { id: "models", label: "Models" },
  { id: "appearance", label: "Appearance" },
  { id: "skills", label: "Skills & plugins" },
  { id: "git", label: "Git" },
  { id: "about", label: "About" },
];

export function SettingsDialog({
  open,
  initialTab,
  appearance,
  onOpenChange,
  onAppearanceChange,
  onModelChange,
}: SettingsDialogProps) {
  const [tab, setTab] = useState<SettingsTab>(initialTab);
  const [settings, setSettings] = useState<Settings | null>(null);
  const [skills, setSkills] = useState<SkillItem[]>([]);
  const [plugins, setPlugins] = useState<PluginItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [switching, setSwitching] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (open) {
      setTab(initialTab);
    }
  }, [open, initialTab]);

  useEffect(() => {
    if (!open) {
      return;
    }
    let cancelled = false;
    setLoading(true);
    setError(null);
    Promise.all([api.settings(), api.skills(), api.plugins()])
      .then(([nextSettings, nextSkills, nextPlugins]) => {
        if (cancelled) {
          return;
        }
        setSettings(nextSettings);
        setSkills(nextSkills);
        setPlugins(nextPlugins);
      })
      .catch((cause: unknown) => {
        if (!cancelled) {
          setError(errorText(cause));
        }
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [open]);

  const changeMode = (mode: ThemeMode) => {
    onAppearanceChange({ ...appearance, mode });
  };

  const toggleNotifications = async () => {
    if (!appearance.notify) {
      if (!("Notification" in window)) {
        toast.error("Notifications are not supported by this browser.");
        return;
      }
      const permission = await Notification.requestPermission();
      if (permission !== "granted") {
        toast.error("Notification permission was not granted.");
        return;
      }
    }
    onAppearanceChange({ ...appearance, notify: !appearance.notify });
  };

  const switchModel = async (provider: string, model: string) => {
    setSwitching(true);
    try {
      await onModelChange(provider, model);
      setSettings((current) =>
        current ? { ...current, active: { provider, model } } : current
      );
    } catch {
      // App reports the failure; keep the dialog state unchanged.
    } finally {
      setSwitching(false);
    }
  };

  return (
    <Dialog onOpenChange={onOpenChange} open={open}>
      <DialogContent className="h-[min(680px,calc(100dvh-2rem))] w-[calc(100%-2rem)] max-w-3xl gap-0 overflow-hidden p-0 sm:max-w-3xl">
        <div className="grid h-full min-h-0 grid-rows-[auto_1fr] md:grid-cols-[190px_1fr] md:grid-rows-1">
          <aside className="flex gap-1 overflow-x-auto border-b bg-muted/30 p-3 md:flex-col md:border-r md:border-b-0">
            <div className="mb-2 hidden px-2 pt-1 md:block">
              <DialogTitle>Settings</DialogTitle>
              <DialogDescription className="mt-1 text-xs">
                Read-only where it matters.
              </DialogDescription>
            </div>
            {tabs.map((item) => (
              <button
                className={cn(
                  "shrink-0 rounded-md px-3 py-2 text-left text-sm transition-colors",
                  tab === item.id
                    ? "bg-background text-foreground shadow-xs"
                    : "text-muted-foreground hover:bg-background/60 hover:text-foreground"
                )}
                key={item.id}
                onClick={() => setTab(item.id)}
                type="button"
              >
                {item.label}
              </button>
            ))}
          </aside>

          <section className="min-h-0 min-w-0 overflow-y-auto p-5">
            <DialogHeader className="mb-4 md:hidden">
              <DialogTitle>Settings</DialogTitle>
            </DialogHeader>
            {error ? (
              <p className="text-sm text-destructive">{error}</p>
            ) : loading || !settings ? (
              <p className="text-sm text-muted-foreground">
                Loading settings...
              </p>
            ) : tab === "models" ? (
              <ModelsTab
                active={settings.active}
                models={settings.models}
                onSwitch={switchModel}
                switching={switching}
              />
            ) : tab === "appearance" ? (
              <AppearanceTab
                appearance={appearance}
                onMode={changeMode}
                onNotifications={() => void toggleNotifications()}
              />
            ) : tab === "skills" ? (
              <SkillsTab plugins={plugins} skills={skills} />
            ) : tab === "git" ? (
              <GitTab git={settings.git} />
            ) : (
              <AboutTab about={settings.about} />
            )}
          </section>
        </div>
      </DialogContent>
    </Dialog>
  );
}

function ModelsTab({
  active,
  models,
  switching,
  onSwitch,
}: {
  active: Settings["active"];
  models: Settings["models"];
  switching: boolean;
  onSwitch: (provider: string, model: string) => Promise<void>;
}) {
  return (
    <div className="space-y-5">
      <Section
        description="Only configured secrets are reported; they never leave the server."
        title="Models"
      />
      {models.map((provider) => (
        <div className="space-y-2" key={provider.id}>
          <div className="flex items-center gap-2">
            <h3 className="text-sm font-medium">{provider.name}</h3>
            <span className="rounded-full bg-muted px-2 py-0.5 text-[11px] text-muted-foreground">
              {provider.configured ? "configured" : "no key"}
            </span>
          </div>
          <div className="overflow-hidden rounded-lg bg-muted/40">
            {provider.models.map((model) => {
              const selected =
                active.provider === provider.id && active.model === model.id;
              return (
                <button
                  className={cn(
                    "flex w-full items-center justify-between gap-4 px-3 py-2.5 text-left text-sm transition-colors hover:bg-muted",
                    selected && "bg-background"
                  )}
                  disabled={switching || selected}
                  key={model.id}
                  onClick={() => void onSwitch(provider.id, model.id)}
                  type="button"
                >
                  <span className="min-w-0 truncate">{model.id}</span>
                  <span className="shrink-0 text-xs text-muted-foreground">
                    {model.context.toLocaleString()} ctx ·{" "}
                    {model.maxOutput.toLocaleString()} out
                  </span>
                </button>
              );
            })}
          </div>
        </div>
      ))}
    </div>
  );
}

function AppearanceTab({
  appearance,
  onMode,
  onNotifications,
}: {
  appearance: Appearance;
  onMode: (mode: ThemeMode) => void;
  onNotifications: () => void;
}) {
  return (
    <div className="space-y-6">
      <Section
        description="The web UI keeps its own clean light/dark palette."
        title="Appearance"
      />
      <div className="space-y-2">
        <p className="text-sm font-medium">Mode</p>
        <div className="flex rounded-lg bg-muted/50 p-1">
          {(["system", "light", "dark"] as const).map((mode) => (
            <button
              className={cn(
                "flex-1 rounded-md px-3 py-1.5 text-sm capitalize transition-colors",
                appearance.mode === mode
                  ? "bg-background text-foreground shadow-xs"
                  : "text-muted-foreground hover:text-foreground"
              )}
              key={mode}
              onClick={() => onMode(mode)}
              type="button"
            >
              {mode}
            </button>
          ))}
        </div>
      </div>

      <label className="flex items-start gap-3 rounded-lg bg-muted/40 p-3">
        <input
          checked={appearance.notify}
          className="mt-0.5 size-4 accent-primary"
          onChange={onNotifications}
          type="checkbox"
        />
        <span className="space-y-1">
          <span className="block text-sm font-medium">
            Notify when a run finishes
          </span>
          <span className="block text-xs text-muted-foreground">
            Only when this tab is in the background. Uses browser notifications.
          </span>
        </span>
      </label>
    </div>
  );
}

function SkillsTab({
  skills,
  plugins,
}: {
  skills: SkillItem[];
  plugins: PluginItem[];
}) {
  return (
    <div className="space-y-6">
      <ItemList
        description="Loaded from the configured skill directories."
        empty="No skills found."
        items={skills}
        title="Skills"
      />
      <ItemList
        description="Read from the configured plugins manifest."
        empty="No plugins found."
        items={plugins}
        title="Plugins"
      />
    </div>
  );
}

function ItemList({
  title,
  description,
  empty,
  items,
}: {
  title: string;
  description: string;
  empty: string;
  items: SkillItem[];
}) {
  return (
    <section className="space-y-2">
      <Section description={description} title={title} />
      {items.length === 0 ? (
        <p className="text-sm text-muted-foreground">{empty}</p>
      ) : (
        <div className="divide-y divide-border/60 overflow-hidden rounded-lg bg-muted/30">
          {items.map((item) => (
            <div
              className="space-y-1 px-3 py-2.5"
              key={`${item.name}-${item.source}`}
            >
              <p className="text-sm font-medium">{item.name}</p>
              {item.description && (
                <p className="text-xs text-muted-foreground">
                  {item.description}
                </p>
              )}
              <p className="truncate font-mono text-[11px] text-muted-foreground/80">
                {item.source}
              </p>
            </div>
          ))}
        </div>
      )}
    </section>
  );
}

function GitTab({ git }: { git: Settings["git"] }) {
  return (
    <div className="space-y-4">
      <Section
        description="Changes are checkpointed in a shadow repository, not your project git."
        title="Shadow repository"
      />
      <div className="space-y-2 rounded-lg bg-muted/40 p-3 text-sm">
        <p>
          Status:{" "}
          <span className="font-medium">
            {git.initialized
              ? git.dirty
                ? "dirty"
                : "clean"
              : "not initialized"}
          </span>
        </p>
        <p className="font-mono text-xs text-muted-foreground">{git.dir}</p>
      </div>
      {!git.initialized && (
        <div className="space-y-1 text-sm text-muted-foreground">
          <p>Runs still work, but no checkpoints are saved.</p>
          <code className="block rounded-md bg-muted px-3 py-2 font-mono text-xs break-all">
            {git.hint}
          </code>
        </div>
      )}
    </div>
  );
}

function AboutTab({ about }: { about: Settings["about"] }) {
  return (
    <div className="space-y-5">
      <Section description="bai web build information." title="About" />
      <dl className="grid gap-3 rounded-lg bg-muted/40 p-3 text-sm">
        <Row label="Version" value={about.version} />
        <Row label="Directory" value={about.directory} mono />
        <Row label="Config" value={about.configPath} mono />
        <Row label="Log" value={about.logPath} mono />
        {about.latestUpdate && (
          <Row label="Latest release" value={about.latestUpdate} />
        )}
      </dl>
    </div>
  );
}

function Row({
  label,
  value,
  mono,
}: {
  label: string;
  value: string;
  mono?: boolean;
}) {
  return (
    <div className="grid gap-1 sm:grid-cols-[110px_1fr]">
      <dt className="text-muted-foreground">{label}</dt>
      <dd className={cn("break-all", mono && "font-mono text-xs")}>{value}</dd>
    </div>
  );
}

function Section({
  title,
  description,
}: {
  title: string;
  description: string;
}) {
  return (
    <div className="space-y-1">
      <h2 className="text-sm font-medium">{title}</h2>
      <p className="text-xs text-muted-foreground">{description}</p>
    </div>
  );
}
