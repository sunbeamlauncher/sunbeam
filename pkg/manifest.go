package pkg

import (
	"encoding/json"
	"fmt"
)

type Manifest struct {
	Title       string    `json:"title"`
	Homepage    string    `json:"homepage,omitempty"`
	Description string    `json:"description,omitempty"`
	Commands    []Command `json:"commands"`
}

type Entrypoint []string

func (e *Entrypoint) UnmarshalJSON(b []byte) error {
	var entrypoint string
	if err := json.Unmarshal(b, &entrypoint); err == nil {
		*e = Entrypoint{entrypoint}
		return nil
	}

	var entrypoints []string
	if err := json.Unmarshal(b, &entrypoints); err == nil {
		*e = Entrypoint(entrypoints)
		return nil
	}

	return fmt.Errorf("invalid entrypoint: %s", string(b))
}

type Command struct {
	Name        string         `json:"name"`
	Title       string         `json:"title"`
	Mode        CommandMode    `json:"mode"`
	Description string         `json:"description,omitempty"`
	Params      []CommandParam `json:"params,omitempty"`
}

type CommandMode string

const (
	CommandModeFilter    CommandMode = "filter"
	CommandModeGenerator CommandMode = "generator"
	CommandModeDetail    CommandMode = "detail"
	CommandModeForm      CommandMode = "form"
	CommandModeSilent    CommandMode = "silent"
)

type CommandParam struct {
	Name        string    `json:"name"`
	Type        ParamType `json:"type"`
	Optional    bool      `json:"optional,omitempty"`
	Description string    `json:"description,omitempty"`
}

type ParamType string

const (
	ParamTypeString  ParamType = "string"
	ParamTypeBoolean ParamType = "boolean"
)

type CommandInput struct {
	Query  string        `json:"query,omitempty"`
	Params CommandParams `json:"params,omitempty"`
}
