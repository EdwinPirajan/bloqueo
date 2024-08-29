package script

import (
	"os/exec"
)

type ScriptRunner struct {
	scriptPath string
}

func NewScriptRunner(scriptPath string) *ScriptRunner {
	return &ScriptRunner{scriptPath: scriptPath}
}

func (sr *ScriptRunner) Run() error {
	cmd := exec.Command("cmd", "/C", sr.scriptPath)
	return cmd.Run()
}
