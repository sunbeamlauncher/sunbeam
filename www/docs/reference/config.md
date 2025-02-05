---
title: Config
---

The config will be searched in the following locations in order:

- `$SUNBEAM_CONFIG`
- `$PWD/sunbeam.json`, and all parent directories
- `$XDG_CONFIG_HOME/sunbeam/sunbeam.json` if `XDG_CONFIG_HOME` is set
- `$HOME/.config/sunbeam/sunbeam.json`

If no config is found, and the `SUNBEAM_CONFIG` environment variable is not set, a default config will be created.

This fallback mechanism allows you to have a project specific configs, and a global config.

```json
{
    // additional items to show in the root list
    "oneliners": [
        {
            "title": "Open Sunbeam Docs",
            // command to run
            "command": "sunbeam open https://pomdtr.github.io/sunbeam",
            // Whether to exit sunbeam after running the command
            "exit": true
        },
        {
            "title": "Edit fish config",
            // command to run
            "command": "sunbeam edit config.fish",
            // working directory to run the command in
            "cwd": "~/.config/fish"
        }
    ],
    // the list of extensions to load
    "extensions": {
        "github": {
            "origin": "~/Developer/github.com/pomdtr/sunbeam/extensions/github.sh",
            // preferences for the extension, use it to pass config or secrets
            "preferences": {
                "token": "xxxx"
            },
            // additional root items to show
            "root": [
                {
                    "title": "List Sunbeam Issues",
                    "command": "list-issues",
                    "params": {
                        "repo": "pomdtr/sunbeam",
                    }
                }
            ]
        }
    }
}
```
