# kekkai desktop

Go-native connectors for `pass` (and, over time, other secret-store backends),
built as a monorepo of small per-connector binaries sharing common store/config/
clipboard logic. This directory holds the Linux side; see `../android` for the
planned Android frontend.

## Connectors

- **kekkai-rofi** (`cmd/kekkai-rofi`) — Rofi script-mode connector. Sub-millisecond
  entry listing, automatic Wayland/X11 clipboard management (`wl-copy` / `xclip`)
  with an auto-clear timer.

## Installation

```bash
go install github.com/quasi-deus/kekkai/desktop/cmd/kekkai-rofi@latest
```

## Rofi Usage

```bash
rofi -show pass -modi "pass:kekkai-rofi"
```

## Configuration

Default configuration (`~/.config/kekkai/config.yaml`):

```yaml
store_dir: "~/.password-store"
clear_after_sec: 45
```
