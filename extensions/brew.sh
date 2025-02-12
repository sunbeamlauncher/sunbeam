#!/bin/bash

set -euo pipefail

# check if jq is installed
if ! [ -x "$(command -v jq)" ]; then
    echo "jq is not installed. Please install it." >&2
    exit 1
fi

# check if brew is installed
if ! [ -x "$(command -v brew)" ]; then
    echo "brew is not installed. Please install it." >&2
    exit 1
fi

if [ $# -eq 0 ]; then
    jq -n '{
        title: "Brew",
        actions: [
            { title: "List Installed Packages", type: "run", command: "list" }
        ],
        commands: [
            {
                name: "list",
                description: "List installed packages",
                mode: "filter",
            },
            {
                name: "info",
                mode: "detail",
                description: "Show info about a package",
                params: [
                    { name: "package", type: "string" }
                ]
            }
        ]
    }'
    exit 0
fi

COMMAND=$1
PARAMS=$(cat)

if [ "$COMMAND" = "list" ]; then
    brew list | jq -R '{
        title: .,
        actions: [
            {
                title: "Show Info",
                type: "run",
                command: "info",
                params: { package: . }
            }
        ]
    }' | jq -s '{ items: . }'
elif [ "$COMMAND" = "info" ]; then
    PACKAGE=$(jq -r '.package' <<< "$PARAMS")
    brew info "$PACKAGE" | jq -sR '{ text: . }'
else
    echo "Unknown command: $COMMAND"
    exit 1
fi
