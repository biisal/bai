export type ThemeMode = "light" | "dark" | "system";

export type Appearance = {
  mode: ThemeMode;
  notify: boolean;
};

const storageKey = "bai-appearance";

export const defaultAppearance: Appearance = {
  mode: "system",
  notify: false,
};

export function readAppearance(): Appearance {
  try {
    const raw = localStorage.getItem(storageKey);
    if (!raw) {
      return defaultAppearance;
    }
    const value = JSON.parse(raw) as Partial<Appearance>;
    return {
      mode:
        value.mode === "light" || value.mode === "dark" ? value.mode : "system",
      notify: value.notify === true,
    };
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
