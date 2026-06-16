package alert

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"

	"github.com/rksdevs/sleepguard/internal/sensor"
)

// Notifier delivers an alert for a motion event.
type Notifier interface {
	Notify(ctx context.Context, event sensor.Event) error
}

// LogNotifier writes a prominent log line when motion is detected.
type LogNotifier struct {
	log *slog.Logger
}

// NewLogNotifier returns a notifier that logs alerts.
func NewLogNotifier(log *slog.Logger) *LogNotifier {
	return &LogNotifier{log: log}
}

// Notify logs the alert event.
func (n *LogNotifier) Notify(_ context.Context, event sensor.Event) error {
	n.log.Warn("ALERT motion detected", "event", event.String())
	return nil
}

// ExecNotifier runs a shell command when motion is detected.
type ExecNotifier struct {
	command string
	log     *slog.Logger
}

// NewExecNotifier returns a notifier that executes command on alert.
func NewExecNotifier(command string, log *slog.Logger) *ExecNotifier {
	return &ExecNotifier{command: command, log: log}
}

// Notify executes the configured command.
func (n *ExecNotifier) Notify(ctx context.Context, event sensor.Event) error {
	cmd := exec.CommandContext(ctx, "sh", "-c", n.command)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("alert command failed: %w", err)
	}
	n.log.Info("alert command executed", "command", n.command, "event", event.String())
	return nil
}
