package dao

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const defaultPowerTimeout = 30 * time.Second

// WwctlPowerManager implements PowerManager by shelling out to the wwctl CLI.
type WwctlPowerManager struct {
	wwctlPath string
}

// NewWwctlPowerManager creates a new power manager. It returns an error if
// the wwctl binary cannot be found on PATH.
func NewWwctlPowerManager() (*WwctlPowerManager, error) {
	path, err := exec.LookPath("wwctl")
	if err != nil {
		return nil, fmt.Errorf("wwctl not found in PATH: %w", err)
	}
	return &WwctlPowerManager{wwctlPath: path}, nil
}

// Status returns the power status string for the given node.
func (p *WwctlPowerManager) Status(nodeID string) (string, error) {
	out, err := p.run("status", nodeID)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// On powers on the given node.
func (p *WwctlPowerManager) On(nodeID string) error {
	_, err := p.run("on", nodeID)
	return err
}

// Off powers off the given node.
func (p *WwctlPowerManager) Off(nodeID string) error {
	_, err := p.run("off", nodeID)
	return err
}

// Cycle power-cycles the given node.
func (p *WwctlPowerManager) Cycle(nodeID string) error {
	_, err := p.run("cycle", nodeID)
	return err
}

// Reset sends a hardware reset to the given node.
func (p *WwctlPowerManager) Reset(nodeID string) error {
	_, err := p.run("reset", nodeID)
	return err
}

// run executes a wwctl power subcommand and returns its stdout.
func (p *WwctlPowerManager) run(action, nodeID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultPowerTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, p.wwctlPath, "power", action, nodeID)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("wwctl power %s %s: timed out after %s", action, nodeID, defaultPowerTimeout)
	}
	if err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "", fmt.Errorf("wwctl power %s %s: %s", action, nodeID, errMsg)
	}
	return stdout.String(), nil
}
