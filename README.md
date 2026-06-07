# deej (Homebrew Version)

This is a fork of the [original deej project](https://github.com/omriharel/deej) by Omri Harel, specifically modified for better compatibility with **Bazzite** and other immutable Linux distributions.

It provides native Homebrew (Linuxbrew) installation and supports standard Linux configuration paths (`~/.config/deej/`).

## Quick Start (Linux / Bazzite)

### 1. Install via Homebrew

Ensure you have [Homebrew](https://brew.sh) installed, then run:

```bash
brew install --build-from-source mxkolb30/deej/deej
```

### 2. Configure

Copy the default configuration to your home folder:

```bash
mkdir -p ~/.config/deej
cp $(brew --prefix)/etc/deej/config.yaml ~/.config/deej/config.yaml
```

Edit `~/.config/deej/config.yaml` to map your sliders and set your port (e.g., `/dev/ttyUSB0`).

### 3. Run & Autostart

Start it manually:
```bash
deej
```

To enable autostart (GNOME/KDE):
```bash
mkdir -p ~/.config/autostart
cat <<EOF > ~/.config/autostart/deej.desktop
[Desktop Entry]
Type=Application
Name=deej
Exec=$(brew --prefix)/bin/deej
Icon=$(brew --prefix)/opt/deej/share/deej/logo.svg
Comment=Hardware volume mixer
Terminal=false
Categories=Audio;Utility;
EOF
```

## Credits

All credit for the original logic, hardware design, and Arduino code goes to [Omri Harel](https://github.com/omriharel/deej). Please visit the [original repository](https://github.com/omriharel/deej) for hardware schematics and the Arduino sketch.

## License
[MIT License](./LICENSE)
