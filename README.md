# rofi-pass-go

Go replacement for rofi-pass.

## Features

- Fast Go implementation
- Kong CLI with YAML config support (`~/.config/rofi-pass-go/config.yaml`)
- Native TOTP (2FA) generation
- Wayland and X11 support (wtype, xdotool, ydotool)
- Clipboard auto-clear timer

## Installation

```bash
go install github.com/quasi-deus/rofi-pass-go@latest
```

## Rofi Usage

```bash
rofi -show pass -modi "pass:rofi-pass-go"
```

## Configuration

Generate config:
```bash
rofi-pass-go init-config
```

Example config (`~/.config/rofi-pass-go/config.yaml`):

```yaml
store_dir: "~/.password-store"
default_action: "copy-pass"
enable_submenu: false
delay_ms: 200
clear_after_sec: 45
autotype_backend: "auto"
```

## CLI Usage

```bash
rofi-pass-go list
rofi-pass-go show email/personal
rofi-pass-go copy email/personal --field pass
rofi-pass-go copy email/personal --field otp
rofi-pass-go type email/personal
```
