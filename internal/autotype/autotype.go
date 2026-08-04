package autotype

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"
)

// Backend specifies the typing automation tool.
type Backend string

const (
	BackendAuto    Backend = "auto"
	BackendXdotool Backend = "xdotool"
	BackendWtype   Backend = "wtype"
	BackendYdotool Backend = "ydotool"
)

// TypeText simulates keypresses to type the target text into the active window.
func TypeText(text string, delayMs int, backend Backend) error {
	if text == "" {
		return nil
	}

	if delayMs > 0 {
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
	}

	selectedBackend := backend
	if selectedBackend == BackendAuto {
		selectedBackend = DetectBackend()
	}

	switch selectedBackend {
	case BackendWtype:
		return typeWtype(text)
	case BackendXdotool:
		return typeXdotool(text, delayMs)
	case BackendYdotool:
		return typeYdotool(text)
	default:
		return fmt.Errorf("no autotype backend available (install wtype, xdotool, or ydotool)")
	}
}

// DetectBackend identifies whether Wayland or X11 is running and selects the appropriate backend.
func DetectBackend() Backend {
	if isWayland() {
		if _, err := exec.LookPath("wtype"); err == nil {
			return BackendWtype
		}
		if _, err := exec.LookPath("ydotool"); err == nil {
			return BackendYdotool
		}
	}

	if _, err := exec.LookPath("xdotool"); err == nil {
		return BackendXdotool
	}

	if _, err := exec.LookPath("wtype"); err == nil {
		return BackendWtype
	}

	return BackendAuto
}

func isWayland() bool {
	return os.Getenv("WAYLAND_DISPLAY") != "" || os.Getenv("XDG_SESSION_TYPE") == "wayland"
}

func typeWtype(text string) error {
	cmd := exec.Command("wtype", text)
	return cmd.Run()
}

func typeXdotool(text string, delayMs int) error {
	// xdotool type --clearmodifiers --delay <ms> "text"
	args := []string{"type", "--clearmodifiers"}
	if delayMs > 0 {
		args = append(args, "--delay", strconv.Itoa(delayMs))
	}
	args = append(args, text)

	cmd := exec.Command("xdotool", args...)
	return cmd.Run()
}

func typeYdotool(text string) error {
	cmd := exec.Command("ydotool", "type", text)
	return cmd.Run()
}
