#!/bin/sh

if [ $# -eq 0 ]; then
    sunbeam query -n '{
        title: "Oneliners",
        root: [
            { command: "list" },
            { command: "create" }
        ],
        commands: [
            {
                name: "list",
                title: "Manage Oneliners",
                mode: "list",
            },
            {
                name: "run",
                title: "Run Oneliner",
                mode: "tty",
                params: [
                    { name: "index", title: "Index", type: "number", required: true }
                ]
            },
            {
                name: "edit",
                hidden: true,
                title: "Edit Oneliner",
                mode: "silent",
                params: [
                    { name: "index", title: "Index", type: "number", required: true },
                    { title: "Title", name: "title", type: "text", required: true },
                    { title: "Command", name: "command", type: "textarea", required: true },
                    { title: "Exit", name: "exit", type: "checkbox", label: "Exit after running command", required: true }
                ]
            },
            {
                name: "delete",
                title: "Delete Oneliner",
                mode: "silent",
                params: [
                    { name: "index", title: "Index", type: "number", required: true }
                ]
            },
            {
                name: "create",
                title: "Create Oneliner",
                mode: "silent",
                params: [
                    { name: "title", title: "Title", type: "text", required: true },
                    { name: "command", title: "Command", type: "textarea", required: true },
                    { name: "exit", title: "Exit", type: "checkbox", label: "Exit after running command", required: false }
                ]
            }
        ]
    }'
    exit 0
fi

if [ -n "$SUNBEAM_CONFIG" ]; then
    CONFIG_PATH="$SUNBEAM_CONFIG"
elif [ -n "$XDG_CONFIG_HOME" ]; then
    CONFIG_PATH="$XDG_CONFIG_HOME/sunbeam/config.json"
else
    CONFIG_PATH="$HOME/.config/sunbeam/config.json"
fi


COMMAND="$(echo "$1" | sunbeam query -r ".command")"
if [ "$COMMAND" = "list" ]; then
    sunbeam query '.oneliners | to_entries | {
        items: map({
            title: .value.title,
            subtitle: .value.command,
            actions: [
                { title: "Run Oneliner", type: "run", command: "run", params: { index: .key } },
                { title: "Copy Command", key: "c", type: "copy", text: .value.command, exit: true },
                { title: "Edit Oneliner", key: "e", type: "run", "command": "edit", params: {
                    index: .key,
                    title: { default: .value.title },
                    command: { default: .value.command },
                    exit: { default: (.value.exit // false) }
                }, reload: true},
                { title: "Delete Oneliner", key: "d", type: "run", command: "delete", params: { index: .key }, reload: true },
                { title: "Create Oneliner", key: "n", type: "run", command: "create", reload: true }
            ]
        })
    }' "$CONFIG_PATH"
elif [ "$COMMAND" = "run" ]; then
    INDEX=$(echo "$1" | sunbeam query -r ".params.index")
    sunbeam query -r ".oneliners[$INDEX].command" "$CONFIG_PATH" | sh
elif [ "$COMMAND" = "delete" ]; then
    INDEX=$(echo "$1" | sunbeam query -r ".params.index")
    # shellcheck disable=SC2016
    sunbeam query --in-place --argjson idx="$INDEX" 'del(
        .oneliners[$idx]
    )' "$CONFIG_PATH"
elif [ "$COMMAND" = "edit" ]; then
    PARAMS=$(echo "$1" | sunbeam query -r ".params")

    # shellcheck disable=SC2016
    sunbeam query --in-place --argjson params="$PARAMS" '.oneliners[$params.index] = {
        title: $params.title,
        command: $params.command,
        exit: $params.exit
    }' "$CONFIG_PATH"

elif [ "$COMMAND" = "create" ]; then
    PARAMS=$(echo "$1" | sunbeam query -r ".params")

    # shellcheck disable=SC2016
    sunbeam query --in-place --argjson params="$PARAMS" '.oneliners += [
        { title: $params.title, command: $params.command, exit: $params.exit }
    ]' "$CONFIG_PATH"
fi