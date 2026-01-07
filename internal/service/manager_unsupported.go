//go:build !darwin && !linux && !windows

package service

import "fmt"

type unsupportedManager struct{}

func DefaultManager() Manager {
	return unsupportedManager{}
}

func (unsupportedManager) Plan(action Action, opts Options) (*Plan, error) {
	return nil, fmt.Errorf("service manager not supported on this OS")
}
