import { code } from "@/lib/code-plugin";
import { cn } from "@/lib/utils";
import { Streamdown } from "streamdown";

export type MarkdownProps = {
  children: string;
  className?: string;
  isAnimating?: boolean;
};

export function Markdown({ children, className, isAnimating }: MarkdownProps) {
  return (
    <Streamdown
      className={cn(
        "size-full [&>*:first-child]:mt-0 [&>*:last-child]:mb-0",
        className
      )}
      controls={{ code: { copy: true } }}
      isAnimating={isAnimating}
      plugins={{ code }}
    >
      {children}
    </Streamdown>
  );
}
