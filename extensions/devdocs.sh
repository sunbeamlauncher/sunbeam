#!/bin/bash

set -euo pipefail

if [ $# -eq 0 ]; then
  jq -n '{
    title: "DevDocs",
    description: "Search DevDocs.io",
    actions: [
      { title: "Search Docsets", type: "run", command: "list-docsets" }
    ],
    commands: [
      {
        name: "list-docsets",
        mode: "filter"
      },
      {
        name: "list-entries",
        mode: "filter",
        params: [
          { name: "slug", title: "Slug", type: "string" }
        ]
      }
    ]
  }'
  exit 0
fi

COMMAND=$1
PARAMS=$(cat)

if [ "$COMMAND" = "list-docsets" ]; then
  # shellcheck disable=SC2016
  curl -sSf https://devdocs.io/docs/docs.json | jq 'map({
      title: .name,
      subtitle: (.release // "latest"),
      accessories: [ .slug ],
      actions: [
        {
          title: "Browse entries",
          type: "run",
          command: "list-entries",
          params: {
            slug: .slug
          }
        }
      ]
    }) | {  items: . }'
elif [ "$1" = "list-entries" ]; then
  SLUG=$(jq -r '.slug' <<< "$PARAMS")
  # shellcheck disable=SC2016
  curl -sSf "https://devdocs.io/docs/$SLUG/index.json" | jq --arg slug "$SLUG" '.entries | map({
      title: .name,
      subtitle: .type,
      actions: [
        {title: "Open in Browser", type: "open", target: "https://devdocs.io/\($slug)/\(.path)"},
        {title: "Copy URL", key: "c", type: "copy", text: "https://devdocs.io/\($slug)/\(.path)"}
      ]
    }) | {  items: . }'
fi
