package clipboard

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// Copy copies the target string to the system clipboard (Wayland or X11).
func Copy(text string) error {
	// Detect Wayland vs X11
	if isWayland() {
		if err := copyWlClipboard(text); err == nil {
			return nil
		}
	}

	// Try X11 xclip
	if err := copyXclip(text); err == nil {
		return nil
	}

	// Try X11 xsel
	if err := copyXsel(text); err == nil {
		return nil
	}

	return fmt.Errorf("no supported clipboard tool found (installed: wl-clipboard, xclip, or xsel required)")
}

// CopyAndClear copies text to clipboard and schedules auto-clear after clearAfter seconds.
func CopyAndClear(text string, clearAfterSec int) error {
	if err := Copy(text); err != nil {
		return err
	}

	if clearAfterSec <= 0 {
		return nil
	}

	// Launch async auto-clear timer
	go func() {
		time.Sleep(time.Duration(clearAfterSec) * time.Second)
		_ = Clear()
	}()

	return nil
}

// Clear clears the current clipboard content.
func Clear() error {
	return Copy("")
}

func isWayland() bool {
	return os.Getenv("WAYLAND_DISPLAY") != "" || os.Getenv("XDG_SESSION_TYPE") == "wayland"
}

func copyWlClipboard(text string) error {
	cmd := exec.Command("wl-copy")
	cmd.Stdin = bytes.NewBufferString(text)
	return cmd.Run()
}

func copyXclip(text string) error {
	cmd := exec.Command("xclip", "-selection", "clipboard")
	cmd.Stdin = bytes.NewBufferString(text)
	return cmd.Run()
}

func copyXsel(text string) error {
	cmd := exec.Command("xsel", "--clipboard", "--input")
	cmd.Stdin = bytes.NewBufferString(text)
	return cmd.Run()
}
