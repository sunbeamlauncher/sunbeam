package types

import "encoding/json"

type View struct {
	Type  ViewType `json:"type"`
	Title string   `json:"title,omitempty"`
}

type ViewType string

const (
	ViewTypeList   ViewType = "list"
	ViewTypeForm   ViewType = "form"
	ViewTypeDetail ViewType = "detail"
)

type List struct {
	Title     string     `json:"title,omitempty"`
	Items     []ListItem `json:"items,omitempty"`
	Dynamic   bool       `json:"dynamic,omitempty"`
	EmptyText string     `json:"emptyText,omitempty"`
	Actions   []Action   `json:"actions,omitempty"`
}

func (l List) MarshalJSON() ([]byte, error) {
	type Alias List
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  "list",
		Alias: (*Alias)(&l),
	})
}

type Detail struct {
	Title    string   `json:"title,omitempty"`
	Actions  []Action `json:"actions,omitempty"`
	Markdown string   `json:"markdown,omitempty"`
}

func (d Detail) MarshalJSON() ([]byte, error) {
	type Alias Detail
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  "detail",
		Alias: (*Alias)(&d),
	})
}

type Form struct {
	Title  string  `json:"title,omitempty"`
	Fields []Field `json:"fields,omitempty"`
}

func (f Form) MarshalJSON() ([]byte, error) {
	type Alias Form
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  "form",
		Alias: (*Alias)(&f),
	})
}

type ListItem struct {
	Id          string   `json:"id,omitempty"`
	Title       string   `json:"title"`
	Subtitle    string   `json:"subtitle,omitempty"`
	Accessories []string `json:"accessories,omitempty"`
	Actions     []Action `json:"actions,omitempty"`
}

type Metadata struct {
	Title  string `json:"title,omitempty"`
	Text   string `json:"text,omitempty"`
	Target string `json:"target,omitempty"`
}

type InputType string

const (
	TextInput     InputType = "text"
	TextAreaInput InputType = "textarea"
	SelectInput   InputType = "select"
	CheckboxInput InputType = "checkbox"
)

type Field struct {
	Title    string `json:"title"`
	Name     string `json:"name,omitempty"`
	Required bool   `json:"required,omitempty"`
	Input    `json:"input"`
}

// TODO: move distinct types to their own structs
type Input struct {
	Type        InputType `json:"type"`
	Placeholder string    `json:"placeholder,omitempty"`
	Default     any       `json:"default,omitempty"`

	// Only for dropdown
	Choices []DropDownItem `json:"choices,omitempty"`

	// Only for checkbox
	Label string `json:"label,omitempty"`
}

type DropDownItem struct {
	Title string `json:"title"`
	Value string `json:"value"`
}

type Action struct {
	Title    string  `json:"title,omitempty"`
	Key      string  `json:"key,omitempty"`
	OnAction Command `json:"onAction,omitempty"`
}