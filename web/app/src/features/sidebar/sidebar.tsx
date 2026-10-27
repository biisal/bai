import {
  Sidebar,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuAction,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSkeleton,
  SidebarRail,
  useSidebar,
} from "@/components/ui/sidebar";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { TooltipProvider } from "@/components/ui/tooltip";
import { api, errorText, type Conversation } from "@/lib/api";
import { cn } from "@/lib/utils";
import {
  MoreHorizontalIcon,
  PlusIcon,
  SearchIcon,
  SettingsIcon,
} from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";
import { toast } from "sonner";

type SidebarProps = {
  conversations: Conversation[];
  activeId: string | null;
  loading: boolean;
  directory: string;
  version: string;
  searchNonce: number;
  onNew: () => void;
  onSelect: (id: number) => void;
  onChanged: (conversation: Conversation) => void;
  onDeleted: (id: number) => void;
  onSettings: () => void;
};

type Group = { label: string; items: Conversation[] };

function groupConversations(conversations: Conversation[]): Group[] {
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  const yesterday = new Date(today.getTime() - 86_400_000);
  const groups: Group[] = [
    { label: "Today", items: [] },
    { label: "Yesterday", items: [] },
    { label: "Earlier", items: [] },
  ];
  for (const conversation of conversations) {
    const created = new Date(conversation.createdAt);
    if (created >= today) {
      groups[0]?.items.push(conversation);
    } else if (created >= yesterday) {
      groups[1]?.items.push(conversation);
    } else {
      groups[2]?.items.push(conversation);
    }
  }
  return groups.filter((group) => group.items.length > 0);
}

export function AppSidebar({
  conversations,
  activeId,
  loading,
  version,
  searchNonce,
  onNew,
  onSelect,
  onChanged,
  onDeleted,
  onSettings,
}: SidebarProps) {
  const { isMobile, setOpenMobile } = useSidebar();
  const [search, setSearch] = useState("");
  const [managed, setManaged] = useState<Conversation | null>(null);
  const searchRef = useRef<HTMLInputElement>(null);
  const closeMobile = () => {
    if (isMobile) {
      setOpenMobile(false);
    }
  };
  useEffect(() => {
    if (searchNonce > 0) {
      searchRef.current?.focus();
      if (isMobile) {
        setOpenMobile(true);
      }
    }
  }, [searchNonce, isMobile, setOpenMobile]);

  const filtered = useMemo(() => {
    const query = search.trim().toLowerCase();
    return query
      ? conversations.filter((item) => item.title.toLowerCase().includes(query))
      : conversations;
  }, [conversations, search]);
  const groups = groupConversations(filtered);

  return (
    <TooltipProvider delay={300}>
      <Sidebar className="border-none" collapsible="offcanvas">
        <SidebarHeader className="gap-3 px-3 pt-4">
          <div className="flex items-center gap-2 px-1">
            <span className="grid size-6 place-items-center rounded-md bg-primary text-xs font-semibold text-primary-foreground">
              b
            </span>
            <span className="text-sm font-semibold tracking-tight">bai</span>
          </div>
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton
                className="bg-background/70 hover:bg-sidebar-accent"
                onClick={() => {
                  onNew();
                  closeMobile();
                }}
                tooltip="New chat"
              >
                <PlusIcon />
                <span>New chat</span>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
          <div className="relative px-1">
            <SearchIcon className="pointer-events-none absolute top-1/2 left-3 size-3.5 -translate-y-1/2 text-muted-foreground" />
            <Input
              aria-label="Search sessions"
              className="h-8 border-0 bg-muted/50 pl-8 text-xs shadow-none focus-visible:bg-background"
              onChange={(event) => setSearch(event.target.value)}
              placeholder="Search sessions"
              ref={searchRef}
              value={search}
            />
          </div>
        </SidebarHeader>

        <ScrollArea className="min-h-0 flex-1">
          {loading ? (
            <SidebarGroup>
              <SidebarGroupLabel>Recent</SidebarGroupLabel>
              <SidebarGroupContent>
                <SidebarMenu>
                  {Array.from({ length: 4 }).map((_, index) => (
                    <SidebarMenuItem key={index}>
                      <SidebarMenuSkeleton />
                    </SidebarMenuItem>
                  ))}
                </SidebarMenu>
              </SidebarGroupContent>
            </SidebarGroup>
          ) : groups.length === 0 ? (
            <p className="px-4 py-6 text-center text-xs text-muted-foreground">
              {search ? "No matching chats" : "No chats yet"}
            </p>
          ) : (
            groups.map((group) => (
              <SidebarGroup key={group.label}>
                <SidebarGroupLabel>{group.label}</SidebarGroupLabel>
                <SidebarGroupContent>
                  <SidebarMenu>
                    {group.items.map((conversation) => (
                      <SidebarMenuItem
                        className="group/menu-item"
                        key={conversation.id}
                      >
                        <SidebarMenuButton
                          className="pr-8 text-md"
                          isActive={activeId === String(conversation.id)}
                          onClick={() => {
                            onSelect(conversation.id);
                            closeMobile();
                          }}
                          tooltip={conversation.title}
                        >
                          <span>{conversation.title}</span>
                        </SidebarMenuButton>
                        <SidebarMenuAction
                          aria-label="Manage conversation"
                          onClick={(event) => {
                            event.stopPropagation();
                            setManaged(conversation);
                          }}
                          showOnHover
                        >
                          <MoreHorizontalIcon />
                        </SidebarMenuAction>
                      </SidebarMenuItem>
                    ))}
                  </SidebarMenu>
                </SidebarGroupContent>
              </SidebarGroup>
            ))
          )}
        </ScrollArea>

        <SidebarFooter className="gap-1 px-3 pb-3">
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton
                className={cn("text-muted-foreground")}
                onClick={onSettings}
                tooltip="Settings"
              >
                <SettingsIcon />
                <span>Settings</span>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
          <div className="px-1 text-[11px] text-muted-foreground/80">
            v{version}
          </div>
        </SidebarFooter>
        <SidebarRail />
      </Sidebar>

      <ManageConversationDialog
        conversation={managed}
        onChanged={onChanged}
        onClose={() => setManaged(null)}
        onDeleted={onDeleted}
      />
    </TooltipProvider>
  );
}

function ManageConversationDialog({
  conversation,
  onChanged,
  onDeleted,
  onClose,
}: {
  conversation: Conversation | null;
  onChanged: (conversation: Conversation) => void;
  onDeleted: (id: number) => void;
  onClose: () => void;
}) {
  const [title, setTitle] = useState("");
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    setTitle(conversation?.title ?? "");
  }, [conversation]);

  const save = async () => {
    if (!conversation) {
      return;
    }
    const next = title.trim();
    if (!next || next === conversation.title) {
      onClose();
      return;
    }
    setSaving(true);
    try {
      const result = await api.renameConversation(
        String(conversation.id),
        next
      );
      onChanged({ ...conversation, title: result.title });
      toast.success("Chat renamed");
      onClose();
    } catch (error: unknown) {
      toast.error(errorText(error));
    } finally {
      setSaving(false);
    }
  };

  const remove = async () => {
    if (!conversation) {
      return;
    }
    setSaving(true);
    try {
      await api.deleteConversation(String(conversation.id));
      onDeleted(conversation.id);
      toast.success("Chat deleted");
      onClose();
    } catch (error: unknown) {
      toast.error(errorText(error));
    } finally {
      setSaving(false);
    }
  };

  return (
    <Dialog
      onOpenChange={(open) => !open && onClose()}
      open={conversation !== null}
    >
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Manage chat</DialogTitle>
          <DialogDescription>
            Rename it, or permanently delete it and its messages.
          </DialogDescription>
        </DialogHeader>
        <form
          className="space-y-2"
          onSubmit={(event) => {
            event.preventDefault();
            void save();
          }}
        >
          <label
            className="text-xs font-medium text-muted-foreground"
            htmlFor="chat-title"
          >
            Title
          </label>
          <Input
            autoFocus
            id="chat-title"
            maxLength={200}
            onChange={(event) => setTitle(event.target.value)}
            value={title}
          />
          <div className="flex justify-end">
            <Button disabled={saving || !title.trim()} size="sm" type="submit">
              Save
            </Button>
          </div>
        </form>
        <DialogFooter className="bg-destructive/5">
          <Button
            disabled={saving}
            onClick={() => void remove()}
            type="button"
            variant="destructive"
          >
            Delete chat
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
