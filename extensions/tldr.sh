#!/bin/bash
# shellcheck disable=SC2016

set -euo pipefail

# This script is an example of how to use tldr with sunbeam
set -eu

# if no arguments are passed, return the extension's manifest
if [ $# -eq 0 ]; then
    jq -n '
{
    title: "TLDR Pages",
    actions: [
        { title: "Search Pages", type: "run", command: "list" }
    ],
    # each command can be called through the cli
    commands: [
        { name: "list", mode: "filter", description: "Search Pages" },
        { name: "view", mode: "detail", description: "View page", params: [{ name: "page", type: "string", description: "Page to show" }] },
        { name: "update", mode: "silent", description: "Update cache" }
    ]
}'
exit 0
fi

# check if tldr is installed
if ! [ -x "$(command -v tldr)" ]; then
    echo "tldr is not installed. Please install it." >&2
    exit 1
fi

COMMAND=$1
PARAMS=$(cat)

if [ "$COMMAND" = "list" ]; then
    tldr --list | jq -R '{
        title: .,
        actions: [
            {title: "View Page", type: "run", command: "view", params: {page: .}},
            { title: "Update Cache", type: "run", command: "update" }
        ]
    }' | jq -s '{ items: . }'
elif [ "$COMMAND" = "view" ]; then
    PAGE=$(jq -r '.page' <<< "$PARAMS")
    tldr --raw "$PAGE" | jq -sR '{
            markdown: ., actions: [
                {title: "Copy Page", type: "copy", text: .}
            ]
        }'
elif [ "$COMMAND" = "update" ]; then
    tldr --update
else
    echo "Invalid command"
    exit 1
fi
