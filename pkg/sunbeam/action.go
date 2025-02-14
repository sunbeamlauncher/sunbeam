package sunbeam

import (
	"encoding/json"
	"fmt"
)

type Action struct {
	Title string     `json:"title,omitempty"`
	Type  ActionType `json:"type,omitempty"`

	Open *OpenAction `json:"-"`
	Copy *CopyAction `json:"-"`
	Run  *RunAction  `json:"-"`
}

func (a *Action) UnmarshalJSON(bts []byte) error {
	var action struct {
		Title string `json:"title,omitempty"`
		Type  string `json:"type,omitempty"`
	}

	if err := json.Unmarshal(bts, &action); err != nil {
		return err
	}

	a.Title = action.Title
	a.Type = ActionType(action.Type)

	switch a.Type {
	case ActionTypeRun:
		a.Run = &RunAction{
			Params: map[string]any{},
		}
		return json.Unmarshal(bts, a.Run)
	case ActionTypeOpen:
		a.Open = &OpenAction{}
		return json.Unmarshal(bts, a.Open)
	case ActionTypeCopy:
		a.Copy = &CopyAction{}
		return json.Unmarshal(bts, a.Copy)
	}

	return nil
}

func (a Action) MarshalJSON() ([]byte, error) {
	switch a.Type {
	case ActionTypeRun:
		output := map[string]interface{}{
			"title":   a.Title,
			"type":    a.Type,
			"command": a.Run.Command,
		}

		if a.Run.Extension != "" {
			output["extension"] = a.Run.Extension
		}

		if len(a.Run.Params) > 0 {
			output["params"] = a.Run.Params
		}

		return json.Marshal(output)
	case ActionTypeOpen:
		return json.Marshal(map[string]interface{}{
			"title":  a.Title,
			"type":   a.Type,
			"target": a.Open.Target,
		})
	case ActionTypeCopy:
		return json.Marshal(map[string]interface{}{
			"title": a.Title,
			"type":  a.Type,
			"text":  a.Copy.Text,
		})
	}

	return nil, fmt.Errorf("unknown action type: %s", a.Type)
}

type ReloadAction struct{}

type RunAction struct {
	Extension string         `json:"extension,omitempty"`
	Command   string         `json:"command,omitempty"`
	Params    map[string]any `json:"params,omitempty"`
}

type CopyAction struct {
	Text string `json:"text,omitempty"`
}

type ExecAction struct {
	Interactive bool   `json:"interactive,omitempty"`
	Command     string `json:"command,omitempty"`
	Dir         string `json:"dir,omitempty"`
}

type OpenAction struct {
	Target string `json:"target,omitempty"`
}

type ActionType string

const (
	ActionTypeRun  ActionType = "run"
	ActionTypeOpen ActionType = "open"
	ActionTypeCopy ActionType = "copy"
)
