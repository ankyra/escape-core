/*
Copyright 2017, 2018 Ankyra

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package core

import (
	"fmt"
)

type ExecStage struct {

	// The command to run. Its arguments, if any, should be defined using the
	// "args" field.
	Cmd string `json:"cmd"`
	// Arguments to the command.
	Args []string `json:"args"`

	// An inline script, which will be executed using bash. It's an error to
	// specify both the "cmd" and "inline" fields.
	Inline string `json:"inline"`

	// Relative path to a script. Deprecated field. Will be used to populate
	// "cmd" and "args" fields to execute this script in `bash`. If the "cmd"
	// field is already populated then this field will be ignored entirely.
	Script string `json:"script"`
}

func (e *ExecStage) ValidateAndFix() error {
	if e.Cmd == "" && e.Script != "" {
		e.Cmd = "bash"
		e.Args = []string{"-c", "./" + e.Script + " .escape/outputs.json"}
		e.Script = ""
	}
	if e.Cmd != "" && e.Inline != "" {
		return fmt.Errorf("Both the cmd and inline field are given.")
	}
	if e.Cmd == "" && e.Inline == "" {
		return fmt.Errorf("Missing script, cmd or inline field.")
	}
	return nil
}
