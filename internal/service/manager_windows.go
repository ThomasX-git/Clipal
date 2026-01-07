//go:build windows

package service

import (
	"fmt"
	"strings"
)

const windowsTaskName = "Clipal"

type windowsManager struct{}

func DefaultManager() Manager {
	return windowsManager{}
}

func (windowsManager) Plan(action Action, opts Options) (*Plan, error) {
	if opts.BinaryPath == "" {
		return nil, fmt.Errorf("binary path is required")
	}
	if opts.ConfigDir == "" {
		return nil, fmt.Errorf("config dir is required")
	}

	runLine := buildWindowsTaskRunLine(opts.BinaryPath, opts.ConfigDir)

	plan := &Plan{}
	switch action {
	case ActionInstall:
		createArgs := []string{"/Create", "/TN", windowsTaskName, "/TR", runLine, "/SC", "ONLOGON"}
		if opts.Force {
			createArgs = append(createArgs, "/F")
		}
		plan.Commands = append(plan.Commands,
			Command{Path: "schtasks.exe", Args: createArgs},
			Command{Path: "schtasks.exe", Args: []string{"/Run", "/TN", windowsTaskName}},
		)
	case ActionUninstall:
		plan.Commands = append(plan.Commands,
			Command{Path: "schtasks.exe", Args: []string{"/End", "/TN", windowsTaskName}, IgnoreError: true},
			Command{Path: "schtasks.exe", Args: []string{"/Delete", "/TN", windowsTaskName, "/F"}},
		)
	case ActionStart:
		plan.Commands = append(plan.Commands, Command{Path: "schtasks.exe", Args: []string{"/Run", "/TN", windowsTaskName}})
	case ActionStop:
		plan.Commands = append(plan.Commands, Command{Path: "schtasks.exe", Args: []string{"/End", "/TN", windowsTaskName}, IgnoreError: true})
	case ActionRestart:
		plan.Commands = append(plan.Commands,
			Command{Path: "schtasks.exe", Args: []string{"/End", "/TN", windowsTaskName}, IgnoreError: true},
			Command{Path: "schtasks.exe", Args: []string{"/Run", "/TN", windowsTaskName}},
		)
	case ActionStatus:
		plan.Commands = append(plan.Commands, Command{Path: "schtasks.exe", Args: []string{"/Query", "/TN", windowsTaskName, "/FO", "LIST", "/V"}})
	default:
		return nil, fmt.Errorf("unsupported action %q", action)
	}

	return plan, nil
}

func buildWindowsTaskRunLine(binaryPath, configDir string) string {
	bin := quoteWindowsCmd(binaryPath)
	cfg := quoteWindowsCmd(configDir)
	return fmt.Sprintf("%s --config-dir %s", bin, cfg)
}

// quoteWindowsCmd quotes a value for a Windows command line fragment.
// This is used for schtasks /TR, which expects a single command line string.
func quoteWindowsCmd(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return `""`
	}
	escaped := strings.ReplaceAll(s, `"`, `\"`)
	return `"` + escaped + `"`
}
