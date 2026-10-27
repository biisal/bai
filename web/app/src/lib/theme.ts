import { z } from "zod";

const appearanceSchema = z.object({
  mode: z.enum(["light", "dark", "system"]).catch("system"),
  notify: z.boolean().catch(false),
});

export type Appearance = z.infer<typeof appearanceSchema>;
export type ThemeMode = Appearance["mode"];

const storageKey = "bai-appearance";

export const defaultAppearance: Appearance = {
  mode: "system",
  notify: false,
};

export function readAppearance(): Appearance {
  try {
    return appearanceSchema.parse(
      JSON.parse(localStorage.getItem(storageKey) ?? "{}")
    );
  } catch {
    return defaultAppearance;
  }
}

export function writeAppearance(appearance: Appearance): void {
  localStorage.setItem(storageKey, JSON.stringify(appearance));
}

export function resolvedTheme(mode: ThemeMode): "light" | "dark" {
  return mode === "dark" ||
    (mode === "system" && matchMedia("(prefers-color-scheme: dark)").matches)
    ? "dark"
    : "light";
}

export function applyAppearance(appearance: Appearance): void {
  document.documentElement.classList.toggle(
    "dark",
    resolvedTheme(appearance.mode) === "dark"
  );
}
