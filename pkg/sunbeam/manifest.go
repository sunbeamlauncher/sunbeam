package sunbeam

type Manifest struct {
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Root        []Action  `json:"root,omitempty"`
	Commands    []Command `json:"commands"`
}

type Command struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Params      []Param     `json:"params,omitempty"`
	Mode        CommandMode `json:"mode,omitempty"`
}

type Platfom string

const (
	PlatformLinux Platfom = "linux"
	PlatformMac   Platfom = "macos"
)

type Requirement struct {
	Name string `json:"name"`
	Link string `json:"link,omitempty"`
}

type CommandMode string

const (
	CommandModeSearch CommandMode = "search"
	CommandModeFilter CommandMode = "filter"
	CommandModeDetail CommandMode = "detail"
	CommandModeAction CommandMode = "action"
	CommandModeSilent CommandMode = "silent"
)

type InputType string

const (
	InputString  InputType = "string"
	InputBoolean InputType = "boolean"
	InputNumber  InputType = "number"
)

type Param struct {
	Type        InputType `json:"type"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Optional    bool      `json:"optional,omitempty"`
	Default     any       `json:"default,omitempty"`
}
