package utils

import (
	"os"
	"os/exec"

	"github.com/google/shlex"
)

func editorEnv() string {
	if editor, ok := os.LookupEnv("VISUAL"); ok {
		return editor
	}

	if editor, ok := os.LookupEnv("EDITOR"); ok {
		return editor
	}

	return "vi"
}

func EditCmd(filepath string) (*exec.Cmd, error) {
	editorArgs, err := shlex.Split(editorEnv())
	if err != nil {
		return nil, err
	}

	editCmd := exec.Command(editorArgs[0])
	editCmd.Args = append(editCmd.Args, editorArgs[1:]...)
	editCmd.Args = append(editCmd.Args, filepath)

	return editCmd, nil
}

func FindShell() string {
	if shell, ok := os.LookupEnv("SHELL"); ok {
		return shell
	}

	return "/bin/sh"
}

func FindPager() string {
	if pager, ok := os.LookupEnv("PAGER"); ok {
		return pager
	}

	return "less"
}
