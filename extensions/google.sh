#!/bin/sh

set -eu

if [ $# -eq 0 ]; then
    jq -n '{
        title: "Google Search",
        description: "Search Google",
        root: [
            { title: "Search Google", type: "run", command: "search" }
        ],
        commands: [
            {
                name: "search",
                mode: "search",
                description: "Search Google",
                params: [
                    { name: "query", title: "Query", type: "string" }
                ]
            }
        ]
    }'
    exit 0
fi

if [ "$1" = "search" ]; then
    QUERY=$(cat | jq -r '.query')

    # urlencode the query
    QUERY=$(echo "$QUERY" | jq -rR '@uri')
    curl "https://suggestqueries.google.com/complete/search?client=firefox&q=$QUERY" | jq '.[1] | {
        dynamic: true,
        items: map({
            title: .,
            actions: [
                { title: "Search", type: "open", target: "https://www.google.com/search?q=\(.)" }
            ]
        })
    }'
else
    echo "Unknown command: $1" >&2
    exit 1
fi
