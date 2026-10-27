import { ChatView, type ChatCommand } from "@/components/chat";
import { Button } from "@/components/ui/button";
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from "@/components/ui/resizable";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar-base";
import {
  api,
  errorText,
  type AppState,
  type Conversation,
  type Models,
} from "@/lib/api";
import { navigate, replacePath, useRoute } from "@/lib/use-route";
import {
  applyAppearance,
  readAppearance,
  resolvedTheme,
  writeAppearance,
  type Appearance,
} from "@/lib/theme";
import { useIsMobile } from "@/hooks/use-mobile";
import type { SettingsTab } from "@/components/settings-dialog";
import { FolderIcon, GitBranchIcon } from "lucide-react";
import type { PanelImperativeHandle } from "react-resizable-panels";
import {
  lazy,
  Suspense,
  useCallback,
  useEffect,
  useRef,
  useState,
} from "react";

const SidebarTooltip = lazy(() =>
  import("@/components/sidebar-tooltip").then((module) => ({
    default: module.SidebarTooltip,
  }))
);
const AppSidebar = lazy(() =>
  import("@/components/sidebar").then((module) => ({
    default: module.AppSidebar,
  }))
);
const ModelSelect = lazy(() =>
  import("@/components/model-select").then((module) => ({
    default: module.ModelSelect,
  }))
);
const Toaster = lazy(() =>
  import("@/components/ui/sonner").then((module) => ({
    default: module.Toaster,
  }))
);
const SettingsDialog = lazy(() =>
  import("@/components/settings-dialog").then((module) => ({
    default: module.SettingsDialog,
  }))
);
const ChangesPanel = lazy(() =>
  import("@/components/changes-panel").then((module) => ({
    default: module.ChangesPanel,
  }))
);
const CommandPalette = lazy(() =>
  import("@/components/command-palette").then((module) => ({
    default: module.CommandPalette,
  }))
);
const ShortcutsDialog = lazy(() =>
  import("@/components/shortcuts-dialog").then((module) => ({
    default: module.ShortcutsDialog,
  }))
);

export default function App() {
  const route = useRoute();
  const [state, setState] = useState<AppState | null>(null);
  const [models, setModels] = useState<Models | null>(null);
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [conversationsLoading, setConversationsLoading] = useState(true);
  const [activeId, setActiveId] = useState<string | null>(route.conversationId);
  const [sessionKey, setSessionKey] = useState(0);
  const [switchingModel, setSwitchingModel] = useState(false);
  const [fatal, setFatal] = useState<string | null>(null);
  const [booted, setBooted] = useState(false);
  const [appearance, setAppearance] = useState<Appearance>(readAppearance);
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [settingsTab, setSettingsTab] = useState<SettingsTab>("models");
  const [paletteOpen, setPaletteOpen] = useState(false);
  const [shortcutsOpen, setShortcutsOpen] = useState(false);
  const [changesOpen, setChangesOpen] = useState(false);
  const [searchNonce, setSearchNonce] = useState(0);
  const initialised = useRef(false);
  const isMobile = useIsMobile();
  const changesPanel = useRef<PanelImperativeHandle | null>(null);
  const changesWidth = useRef("40%");

  const load = useCallback(async () => {
    setFatal(null);
    setConversationsLoading(true);
    try {
      const [nextState, nextModels, nextConversations] = await Promise.all([
        api.state(),
        api.models(),
        api.conversations(),
      ]);
      setState(nextState);
      setModels(nextModels);
      setConversations(nextConversations);
    } catch (error: unknown) {
      setFatal(errorText(error));
    } finally {
      setConversationsLoading(false);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  useEffect(() => {
    applyAppearance(appearance);
    writeAppearance(appearance);
    if (appearance.mode !== "system") {
      return;
    }
    const media = matchMedia("(prefers-color-scheme: dark)");
    const sync = () => applyAppearance(appearance);
    media.addEventListener("change", sync);
    return () => media.removeEventListener("change", sync);
  }, [appearance]);

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "k") {
        event.preventDefault();
        setPaletteOpen((open) => !open);
        return;
      }
      if ((event.metaKey || event.ctrlKey) && event.key === ",") {
        event.preventDefault();
        openSettings("models");
        return;
      }
      const target = event.target;
      const editing =
        target instanceof HTMLInputElement ||
        target instanceof HTMLTextAreaElement ||
        target instanceof HTMLSelectElement ||
        (target instanceof HTMLElement && target.isContentEditable);
      if (!editing && event.key === "?") {
        event.preventDefault();
        setShortcutsOpen(true);
      }
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, []);

  useEffect(() => {
    if (!state || initialised.current) {
      return;
    }
    if (state.running && !route.conversationId) {
      const id = String(state.running.conversationId);
      setActiveId(id);
      replacePath(`/c/${id}`);
    }
    initialised.current = true;
    setBooted(true);
  }, [state, route.conversationId]);

  useEffect(() => {
    setActiveId(route.conversationId);
  }, [route.conversationId]);

  useEffect(() => {
    const panel = changesPanel.current;
    if (!panel) {
      return;
    }
    if (changesOpen && !isMobile) {
      panel.resize(changesWidth.current);
    } else {
      panel.collapse();
    }
  }, [changesOpen, isMobile]);

  const refreshConversations = useCallback(async () => {
    try {
      setConversations(await api.conversations());
    } catch {
      return;
    }
  }, []);

  const refreshState = useCallback(async () => {
    try {
      setState(await api.state());
    } catch {
      return;
    }
  }, []);

  const openSettings = (tab: SettingsTab) => {
    setSettingsTab(tab);
    setSettingsOpen(true);
  };

  const handleNew = () => {
    navigate("/");
    setSessionKey((key) => key + 1);
    setActiveId(null);
  };

  const handleChatCommand = (command: ChatCommand) => {
    if (command === "new") {
      handleNew();
    } else if (command === "sessions") {
      setSearchNonce((value) => value + 1);
    } else {
      openSettings("models");
    }
  };

  const handleSelect = (id: number) => {
    navigate(`/c/${id}`);
    setSessionKey((key) => key + 1);
    setActiveId(String(id));
  };

  const handleConversation = (id: string) => {
    setActiveId(id);
    void refreshConversations();
  };

  const handleRunEnd = () => {
    void refreshConversations();
    void refreshState();
  };

  const handleModelChange = async (value: string, selected?: string) => {
    const separator = selected === undefined ? value.indexOf(":") : -1;
    const provider = selected === undefined ? value.slice(0, separator) : value;
    const model =
      selected === undefined ? value.slice(separator + 1) : selected;
    if (!provider || !model) {
      return;
    }
    setSwitchingModel(true);
    try {
      await api.setModel(provider, model);
      setModels((current) =>
        current ? { ...current, provider, model } : current
      );
      setState((current) =>
        current ? { ...current, provider, model } : current
      );
    } catch (error: unknown) {
      void import("sonner").then(({ toast }) => {
        toast.error(errorText(error));
      });
      throw error;
    } finally {
      setSwitchingModel(false);
    }
  };

  if (fatal) {
    return (
      <main className="grid min-h-dvh place-items-center bg-background p-6 text-foreground">
        <div className="max-w-md space-y-4 text-center">
          <p className="text-sm text-muted-foreground">{fatal}</p>
          {!fatal.startsWith("Unauthorized.") && (
            <Button onClick={() => void load()} type="button">
              Retry
            </Button>
          )}
        </div>
      </main>
    );
  }

  if (!state || !booted) {
    return (
      <main className="grid min-h-dvh place-items-center bg-background p-6 text-foreground">
        <p className="text-sm text-muted-foreground">Loading bai...</p>
      </main>
    );
  }

  const runningId = state.running ? String(state.running.conversationId) : null;
  const initialResume =
    route.conversationId !== null && route.conversationId === runningId;
  const modelValue = models ? `${models.provider}:${models.model}` : "";
  const chatView = (
    <ChatView
      conversationId={route.conversationId}
      directory={state.directory}
      initialResume={initialResume}
      key={`${route.conversationId ?? "new"}:${sessionKey}`}
      notify={appearance.notify}
      onCommand={handleChatCommand}
      onConversation={handleConversation}
      onRunEnd={handleRunEnd}
    />
  );

  return (
    <>
      <SidebarProvider className="h-dvh">
        <Suspense
          fallback={
            <div
              aria-hidden
              className="hidden w-64 shrink-0 bg-sidebar md:block"
            />
          }
        >
          <AppSidebar
            activeId={activeId}
            conversations={conversations}
            directory={state.directory}
            loading={conversationsLoading}
            onChanged={(conversation) =>
              setConversations((current) =>
                current.map((item) =>
                  item.id === conversation.id ? conversation : item
                )
              )
            }
            onDeleted={(id) => {
              setConversations((current) =>
                current.filter((item) => item.id !== id)
              );
              if (activeId === String(id)) {
                handleNew();
              }
            }}
            onNew={handleNew}
            onSelect={handleSelect}
            onSettings={() => openSettings("models")}
            searchNonce={searchNonce}
            version={state.version}
          />
        </Suspense>

        <SidebarInset className="h-dvh min-w-0 overflow-hidden">
          <header className="flex h-12 shrink-0 items-center gap-3 px-3">
            <Suspense fallback={<SidebarTrigger />}>
              <SidebarTooltip>
                <SidebarTrigger />
              </SidebarTooltip>
            </Suspense>
            <div className="flex min-w-0 items-center gap-2 text-sm text-muted-foreground">
              <FolderIcon className="size-4 shrink-0" />
              <span className="truncate" title={state.directory}>
                {state.directory}
              </span>
            </div>
            <div className="ml-auto flex shrink-0 items-center gap-2">
              <Button
                aria-pressed={changesOpen}
                disabled={!route.conversationId}
                onClick={() => setChangesOpen((open) => !open)}
                size="sm"
                type="button"
                variant="ghost"
              >
                <GitBranchIcon />
                <span className="hidden sm:inline">Changes</span>
              </Button>
              <Suspense
                fallback={
                  <div
                    aria-hidden
                    className="h-8 w-28 rounded-lg bg-muted/50"
                  />
                }
              >
                <ModelSelect
                  disabled={switchingModel}
                  models={models}
                  onChange={(value) =>
                    void handleModelChange(value).catch(() => undefined)
                  }
                  value={modelValue}
                />
              </Suspense>
            </div>
          </header>

          <ResizablePanelGroup
            className="min-h-0 flex-1"
            orientation="horizontal"
          >
            <ResizablePanel minSize="35%" style={{ overflow: "hidden" }}>
              <div className="flex h-full min-h-0 flex-col">{chatView}</div>
            </ResizablePanel>
            <ResizableHandle
              className={changesOpen && !isMobile ? undefined : "hidden"}
              withHandle
            />
            <ResizablePanel
              className="min-w-0"
              collapsedSize="0%"
              collapsible
              defaultSize="0%"
              maxSize="70%"
              minSize="25%"
              onResize={(size) => {
                if (size.asPercentage > 0) {
                  changesWidth.current = `${size.asPercentage}%`;
                }
              }}
              panelRef={changesPanel}
              style={{ overflow: "hidden" }}
            >
              {!isMobile && changesOpen && (
                <Suspense fallback={<div className="h-full bg-background" />}>
                  <ChangesPanel
                    conversationId={route.conversationId}
                    inline
                    onOpenChange={setChangesOpen}
                    open={changesOpen}
                    theme={resolvedTheme(appearance.mode)}
                  />
                </Suspense>
              )}
            </ResizablePanel>
          </ResizablePanelGroup>
        </SidebarInset>
      </SidebarProvider>
      <Suspense fallback={null}>
        <Toaster position="bottom-right" />
      </Suspense>
      <Suspense fallback={null}>
        <SettingsDialog
          appearance={appearance}
          initialTab={settingsTab}
          onAppearanceChange={setAppearance}
          onModelChange={handleModelChange}
          onOpenChange={setSettingsOpen}
          open={settingsOpen}
        />
      </Suspense>
      {isMobile && (
        <Suspense fallback={null}>
          <ChangesPanel
            conversationId={route.conversationId}
            onOpenChange={setChangesOpen}
            open={changesOpen}
            theme={resolvedTheme(appearance.mode)}
          />
        </Suspense>
      )}
      <Suspense fallback={null}>
        <CommandPalette
          conversations={conversations}
          models={models}
          onChanges={() => setChangesOpen(true)}
          onModel={(provider, model) =>
            void handleModelChange(provider, model).catch(() => undefined)
          }
          onNew={handleNew}
          onOpenChange={setPaletteOpen}
          onSelect={handleSelect}
          onSettings={openSettings}
          open={paletteOpen}
        />
      </Suspense>
      <Suspense fallback={null}>
        <ShortcutsDialog onOpenChange={setShortcutsOpen} open={shortcutsOpen} />
      </Suspense>
    </>
  );
}
