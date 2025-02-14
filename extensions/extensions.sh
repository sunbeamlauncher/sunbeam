#!/usr/bin/env bash

set -euo pipefail


if [ $# -eq 0 ]; then
  jq -n '{
    title: "Sunbeam",
    actions: [
      { title: "Search Extensions", type: "run", command: "ls" }
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
    sunbeam| jq '.[] |{
        title: .name,
        accessories: [
            .entrypoint
        ],
        actions: [
            { title: "Open extension", type: "open", target: .entrypoint },
            { title: "Copy Path", type: "copy", text: .entrypoint },
            { title: "Remove Extension", type: "run", command: "rm", params: { path: .entrypoint }, reload: true }
        ]
    }' | jq -s '{ items: . }'
elif [ "$COMMAND" = "rm" ]; then
  EXTENSION_PATH=$(jq -r '.path' <<< "$PARAMS")
  rm "$EXTENSION_PATH"
fi
