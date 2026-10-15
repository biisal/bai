---
name: theme_maker
description: Guide for creating a new bai color theme. Read it whenever the user asks you to make, add, install, tweak, or fix a theme, a color scheme, or the TUI colors.
---

# Creating bai themes

A theme is one JSON file of color values. bai lists every `*.json` in the themes directory and applies the one the user picks; nothing else has to be registered, imported, or compiled.

The user's configured location (resolved by bai for this session, do not guess another path):

    {{themes_dir}}/<file_name>.json

### Before you write anything

1. Run `ls -la {{themes_dir}}` and read one or two existing files. They are your style reference.
2. Reuse an existing theme if one already matches what the user wants — say so instead of adding a near-duplicate.

### File format

The file name is the theme name shown in the picker. Name it after the scheme or the look, lowercase, hyphenated: `colors.gruvbox.json`, `colors.solarized-light.json`. Do not edit `default.json`; it is the fallback theme.

Every key must be present or it silently falls back to an unset color:

    {
      "background": "#282a36",
      "foreground": "#f8f8f2",
      "muted": "#44475a",
      "mutedForeground": "#6272a4",
      "primary": "#bd93f9",
      "primaryForeground": "#282a36",
      "secondary": "#6272a4",
      "secondaryForeground": "#f8f8f2",
      "accent": "#ff79c6",
      "accentForeground": "#282a36",
      "border": "#bd93f9",
      "destructive": "#ff5555",
      "destructiveForeground": "#282a36",
      "success": "#50fa7b",
      "successForeground": "#282a36",
      "warning": "#f1fa8c",
      "warningForeground": "#282a36"
    }

- Values are `#rrggbb` hex or a 0-255 ANSI index as a string (`"236"`). Hex is the safer choice; it renders the same on every terminal.
- `background` stays `""` in the built-in default theme, meaning "leave the terminal background alone". Set it to a hex value only when the user asks for a full-color theme, because it fights with a transparent terminal.
- The rest are plain semantic roles, not structural: pick colors that keep text readable. `*Foreground` is the text drawn on top of its base color, so contrast the pair — a light `primary` gets a dark `primaryForeground` and vice versa. Get this wrong and the UI becomes unreadable.
- `destructive` is the error color, `success` and `warning` are the notice colors. Keep them recognizable (red / green / yellow) even in an opinionated palette.

### Steps

1. Write the file to `{{themes_dir}}/<name>.json` with `write_file`.
2. Validate it: `python3 -m json.tool {{themes_dir}}/<name>.json > /dev/null`. A malformed file is reported as an error and falls back to the default theme.
3. Tell the user to run `/themes` and pick it. The list is read from disk on every open, so no restart is needed, and the choice is saved for the next start.

### Safety

- Do not delete or overwrite an existing theme unless the user asked for that theme specifically.
- The alphabetically first file becomes the default on a fresh install (no saved choice yet). If the user is adding a one-off theme, leave the `colors.*` prefix off or accept that it may become the default.
