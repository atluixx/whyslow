package service

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	UnitName   = "whyslow.service"
	UnitPath   = "/etc/systemd/system/whyslow.service"
	BinaryPath = "/usr/local/bin/whyslow"
	unitMarker = "# Managed by whyslow. Remove with `whyslow uninstall`."
)

type Status struct {
	Installed bool
	State     string
	PID       int
	Uptime    string
}

func Install() error {
	if err := requireRoot(); err != nil {
		return err
	}
	if err := requireSystemd(); err != nil {
		return err
	}
	if err := installBinary(); err != nil {
		return err
	}
	if err := installUnit(); err != nil {
		return err
	}
	if _, err := systemctl("daemon-reload"); err != nil {
		return err
	}
	if _, err := systemctl("enable", UnitName); err != nil {
		return err
	}
	return nil
}

func Uninstall() error {
	if err := requireRoot(); err != nil {
		return err
	}
	data, err := os.ReadFile(UnitPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read service unit: %w", err)
	}
	if !strings.Contains(string(data), unitMarker) {
		return fmt.Errorf("refusing to remove unmanaged service unit %s", UnitPath)
	}
	if err := requireSystemd(); err != nil {
		return err
	}
	_, _ = systemctl("stop", UnitName)
	_, _ = systemctl("disable", UnitName)
	if err := os.Remove(UnitPath); err != nil {
		return fmt.Errorf("remove service unit: %w", err)
	}
	if _, err := systemctl("daemon-reload"); err != nil {
		return err
	}
	if err := os.Remove(BinaryPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove installed binary: %w", err)
	}
	return nil
}

func Start() error   { return control("start") }
func Stop() error    { return control("stop") }
func Restart() error { return control("restart") }

func control(action string) error {
	if !Installed() {
		return fmt.Errorf("whyslow service is not installed; run `whyslow install` as root first")
	}
	if err := requireSystemd(); err != nil {
		return err
	}
	_, err := systemctl(action, UnitName)
	return err
}

func GetStatus() (Status, error) {
	if !Installed() {
		return Status{}, nil
	}
	if err := requireSystemd(); err != nil {
		return Status{}, err
	}
	output, err := systemctl("show", UnitName, "--property=ActiveState", "--property=MainPID", "--property=ActiveEnterTimestampMonotonic")
	if err != nil {
		return Status{}, err
	}
	values := parseProperties(output)
	pid, _ := strconv.Atoi(values["MainPID"])
	status := Status{Installed: true, State: values["ActiveState"], PID: pid}
	if status.State == "active" {
		status.Uptime = uptime(values["ActiveEnterTimestampMonotonic"])
	}
	return status, nil
}

func Installed() bool {
	data, err := os.ReadFile(UnitPath)
	return err == nil && strings.Contains(string(data), unitMarker)
}

func requireRoot() error {
	if os.Geteuid() != 0 {
		return errors.New("this command requires root; rerun it with sudo or as root")
	}
	return nil
}

func requireSystemd() error {
	if _, err := exec.LookPath("systemctl"); err != nil {
		return errors.New("systemd is not available: systemctl was not found")
	}
	if _, err := os.Stat("/run/systemd/system"); err != nil {
		return errors.New("systemd is not running on this system")
	}
	return nil
}

func installBinary() error {
	source, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate current executable: %w", err)
	}
	if target, err := os.Stat(BinaryPath); err == nil {
		if current, statErr := os.Stat(source); statErr == nil && os.SameFile(target, current) {
			return nil
		}
		return fmt.Errorf("refusing to overwrite existing %s", BinaryPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect install target: %w", err)
	}
	data, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("read current executable: %w", err)
	}
	temporary := BinaryPath + ".tmp"
	if err := os.WriteFile(temporary, data, 0755); err != nil {
		return fmt.Errorf("write install target: %w", err)
	}
	if err := os.Rename(temporary, BinaryPath); err != nil {
		_ = os.Remove(temporary)
		return fmt.Errorf("install binary: %w", err)
	}
	return nil
}

func installUnit() error {
	if data, err := os.ReadFile(UnitPath); err == nil && !strings.Contains(string(data), unitMarker) {
		return fmt.Errorf("refusing to overwrite unmanaged service unit %s", UnitPath)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect service unit: %w", err)
	}
	return os.WriteFile(UnitPath, []byte(UnitContents()), 0644)
}

func UnitContents() string {
	return unitMarker + `
[Unit]
Description=whyslow Linux performance diagnostic monitor
After=multi-user.target

[Service]
Type=simple
ExecStart=/usr/local/bin/whyslow service --interval=5s
Restart=on-failure
RestartSec=5s
DynamicUser=yes
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ProtectHome=yes
ProtectKernelTunables=yes
ProtectKernelModules=yes
ProtectControlGroups=yes

[Install]
WantedBy=multi-user.target
`
}

func systemctl(args ...string) (string, error) {
	command := exec.Command("systemctl", args...)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	if err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return "", fmt.Errorf("systemctl %s: %s", strings.Join(args, " "), message)
	}
	return string(output), nil
}

func parseProperties(output string) map[string]string {
	values := make(map[string]string)
	for _, line := range strings.Split(output, "\n") {
		key, value, ok := strings.Cut(line, "=")
		if ok {
			values[key] = value
		}
	}
	return values
}

func uptime(activeSince string) string {
	microseconds, err := strconv.ParseUint(activeSince, 10, 64)
	if err != nil || microseconds == 0 {
		return "unknown"
	}
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return "unknown"
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return "unknown"
	}
	seconds, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return "unknown"
	}
	duration := int64(seconds) - int64(microseconds/1_000_000)
	if duration < 0 {
		return "unknown"
	}
	return formatDuration(duration)
}

func formatDuration(seconds int64) string {
	hours, minutes, seconds := seconds/3600, (seconds%3600)/60, seconds%60
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
