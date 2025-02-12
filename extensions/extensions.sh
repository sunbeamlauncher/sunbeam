#!/usr/bin/env bash

set -euo pipefail

EXTENSIONS_DIR="${SUNBEAM_EXTENSIONS_DIR:-$HOME/.config/sunbeam/extensions}"

if [ $# -eq 0 ]; then
  jq -n --arg dir "$EXTENSIONS_DIR" '{
    title: "Sunbeam",
    actions: [
      { title: "Search Extensions", type: "run", command: "ls" },
      { title: "Open Extensions Dir", type: "open", target: $dir }
    ],
    commands: [
      { name: "ls", mode: "filter" },
      { name: "rm", mode: "silent", params: [{ name: "path", type: "string" }] } 
    ]
  }'
  exit 0
fi

COMMAND=$1
PARAMS=$(cat)

if [ "$COMMAND" = "ls" ]; then
    find "$EXTENSIONS_DIR" -type f -or -type l | jq --arg dir "$EXTENSIONS_DIR" -R '{
        title: (. | split("/") | last),
        accessories: [
            .
        ],
        actions: [
            { title: "Open extension", type: "open", target: . },
            { title: "Copy Path", type: "copy", text: . },
            { title: "Remove Extension", type: "run", command: "rm", params: { path: . }, reload: true }
        ]
    }' | jq -s '{ items: . }'
elif [ "$COMMAND" = "rm" ]; then
  EXTENSION_PATH=$(jq -r '.path' <<< "$PARAMS")
  rm -r "$EXTENSION_PATH"
fi
