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

export function applyAppearance(appearance: Appearance): void {
  const root = document.documentElement;
  const dark =
    appearance.mode === "dark" ||
    (appearance.mode === "system" &&
      matchMedia("(prefers-color-scheme: dark)").matches);
  root.classList.toggle("dark", dark);
}
