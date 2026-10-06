---
name: tools_maker
description: Guide for creating new custom tools and skills for bai. Read it whenever the user asks you to make, add, or fix a tool, plugin, or skill.
---

# Creating bai tools and skills

## 1. Custom tools

A custom tool is an executable script plus one entry in a JSON manifest. bai turns every manifest entry into a tool you can call.

The user's configured locations (resolved by bai for this session, do not guess other paths):

    manifest:  {{tools_manifest}}
    scripts:   {{tools_dir}}/<file_name>    (always the manifest's own directory)

The manifest path comes from the `plugins_path` key in the user's config, so it can be anywhere and the file does not have to be called `tools.json`. Use the paths above everywhere below.

### Before you write anything

1. Run `ls -la {{tools_dir}}` and read `{{tools_manifest}}`. If either is missing, create the directory and start the manifest as `[]`. Existing entries are your style reference and the file you must not break.
2. Check the request is not already covered by a built-in tool (`read`, `write_file`, `edit_file`, `bash`) or an existing custom tool. If it is, say so instead of building a duplicate.

### Manifest format

`tools.json` is a JSON array. Add one object per tool:

    [
      {
        "file_name": "weather.sh",
        "description": "Get the current weather for a city. ",
        "input_schema": {
          "city": "city name, e.g. Patna",
          "units": "optional: metric or imperial"
        }
      }
    ]

Rules, all enforced by the loader:

- `file_name`: script name, relative to the manifest directory. No subpaths.
- `description`: say what the tool does and when to use it. bai appends `args: key: description ...` straight after it with no separator, so end it with ". " (period and space).
- `input_schema`: an object whose keys are argument names and whose values are plain-string descriptions. Every value must be a string. There are no types and no required flag: mark optional arguments with the word "optional" in the description, and handle them being absent in the script.
- Tool name = `file_name` without its extension, with every character outside `A-Z a-z 0-9 _ -` replaced by `_`. `weather.sh` becomes `weather`. Never reuse a built-in name or an existing tool name.
- The manifest must stay valid JSON. If it fails to parse, bai silently drops ALL custom tools. Always edit it with the edit tool (keep the existing entries untouched) and validate afterwards with `python3 -m json.tool {{tools_manifest}} > /dev/null` (or `jq . {{tools_manifest}}`).

### How bai runs the script

- It runs `'<tools dir>/<file_name>' "$@"` through bash, in the directory bai was started from.
- Each argument arrives as a separate positional parameter in the form `key=value`, in no guaranteed order. Parse by key, never by position. Split on the first `=` only, because values may contain `=`.
- There is no stdin.
- stdout and stderr are merged and returned to you. Empty output becomes "(no output)". Output over 2000 lines or 256KB is truncated to the tail, so keep it short and useful.
- A non-zero exit code is reported as an error together with the output. Print a clear message to stderr and exit non-zero on failure. Exit 0 on success.
- There is no default timeout, so make sure the script cannot hang (use `timeout`, `curl --max-time`, and so on).

### Script template (bash)

    #!/usr/bin/env bash
    set -euo pipefail

    city="" units="metric"
    for arg in "$@"; do
      case "$arg" in
        city=*)  city="${arg#city=}" ;;
        units=*) units="${arg#units=}" ;;
      esac
    done

    [ -n "$city" ] || { echo "missing required arg: city" >&2; exit 1; }

    curl -fsS --max-time 10 "https://wttr.in/${city}?format=3"

Any language works (python, node, go binary) as long as the file starts with a shebang or is a compiled binary, and it parses the same `key=value` arguments. Prefer bash for anything short. Do not add dependencies the user has not installed.

### Steps

1. Look at the existing tools (see above).
2. Write the script to `{{tools_dir}}/<file_name>` with `write_file`.
3. `chmod +x {{tools_dir}}/<file_name>`. A non-executable script fails at runtime.
4. Add the entry to `{{tools_manifest}}` with `edit_file` and validate the JSON.
5. Test it with bash exactly as bai will call it, for example `{{tools_dir}}/weather.sh city=Patna`. Also test a missing required argument. Fix until both behave.
6. Tell the user the tool name and its arguments, and that **bai loads tools at startup, so they must restart bai before the new tool appears**.

### Safety

- Never hardcode secrets. Read them from environment variables and tell the user which variable to set.
- Quote every variable expansion. Never `eval` or interpolate argument values into a shell command string.
- Do not make a tool destructive (delete, overwrite, push, send) unless the user explicitly asked for that, and name it so the effect is obvious.
- Do not overwrite or rename an existing script or entry unless the user asked you to change that tool.

## 2. Skills

A skill is a markdown instruction file that bai lists in its prompt and you read when the task matches its description. Use a skill (not a tool) when the job is a procedure or knowledge for you to follow rather than a command to run.

Location: skill directories come from the `skills_paths` key in the user's bai config (`~/.config/bai/config.json`, or the file passed with `-config`). Read that file and use its first entry. If the key is absent, the defaults apply, so use `~/.config/bai/skills`. Other default directories (`~/.agents/skills`, `~/.claude/skills`, and the project-local `.bai/skills`) also work. If you cannot tell which config the user runs with, ask.

    <skills dir>/<skill-name>/SKILL.md

Format:

    ---
    name: skill-name
    description: One line saying what it does and when to read it.
    ---

    Instructions in markdown.

Rules:

- The directory name and `name:` must match exactly, otherwise the skill is ignored.
- Names are 1 to 64 characters: lowercase letters and digits, single hyphens between words, no leading, trailing, or double hyphens. `pdf-notes` is valid. `PDF_Notes` is not.
- `description` must be one line. It is the only thing bai shows up front, so make it specific enough to trigger at the right moment.
- The file must start with `---` on the very first line, and the frontmatter must be closed by a second `---` line.
- Keep it short and imperative. Link to scripts or files rather than pasting large content.
- Skills are discovered at startup, so tell the user to restart bai after you add one.

## 3. Tool or skill?

- Needs to execute something and return output (call an API, run a CLI, query a DB): tool.
- Needs to teach you a workflow, conventions, or domain knowledge: skill.
- Both (a workflow that uses a script): make the tool, then a skill that explains when and how to use it.
