import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import type { ReactElement } from "react";

export function SidebarTooltip({ children }: { children: ReactElement }) {
  return (
    <TooltipProvider delay={300}>
      <Tooltip>
        <TooltipTrigger render={children} />
        <TooltipContent side="bottom">
          Toggle sidebar
          <kbd className="ml-1 rounded bg-background/20 px-1 font-sans text-[10px]">
            ⌘B
          </kbd>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
