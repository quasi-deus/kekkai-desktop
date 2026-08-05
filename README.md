# rofi-pass-go

Minimal Go replacement for `rofi-pass`.

## Features

- Sub-millisecond execution
- Simple YAML configuration (`~/.config/rofi-pass-go/config.yaml`)
- Automatic Wayland/X11 clipboard management (`wl-copy` / `xclip`) with auto-clear timer

## Installation

```bash
go install github.com/quasi-deus/rofi-pass-go@latest
```

## Rofi Usage

```bash
rofi -show pass -modi "pass:rofi-pass-go"
```

## Configuration

Default configuration (`~/.config/rofi-pass-go/config.yaml`):

```yaml
store_dir: "~/.password-store"
clear_after_sec: 45
```
