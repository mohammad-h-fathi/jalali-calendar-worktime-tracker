#!/usr/bin/env bash
set -e

DIR="$(cd "$(dirname "$0")/.." && pwd)"

echo "Building Persian Calendar..."
mkdir -p "$DIR/bin"
go build -o "$DIR/bin/persian-calendar" "$DIR/cmd/persian-calendar"

mkdir -p "$HOME/.local/bin"
rm -f "$HOME/.local/bin/persian-calendar"
install -m 755 "$DIR/bin/persian-calendar" "$HOME/.local/bin/persian-calendar"
install -m 755 "$DIR/scripts/floating_timer.py" "$HOME/.local/bin/persian-timer-pill"

mkdir -p "$HOME/.local/share/fonts/Lahzeh"
cp "$DIR"/fonts/*.ttf "$HOME/.local/share/fonts/Lahzeh/" 2>/dev/null || true
fc-cache -f "$HOME/.local/share/fonts/Lahzeh" 2>/dev/null || true

# Install icons in multiple resolutions
mkdir -p "$HOME/.local/share/icons/hicolor/scalable/apps"
cp "$DIR/web/assets/icons/app-icon.svg" "$HOME/.local/share/icons/hicolor/scalable/apps/persian-calendar.svg"

for size in 16 32 48 64 128 256 512; do
    mkdir -p "$HOME/.local/share/icons/hicolor/${size}x${size}/apps"
    if [ -f "$DIR/web/assets/icons/app-icon-${size}.png" ]; then
        cp "$DIR/web/assets/icons/app-icon-${size}.png" "$HOME/.local/share/icons/hicolor/${size}x${size}/apps/persian-calendar.png"
    elif [ -f "$DIR/web/assets/icons/app-icon.png" ]; then
        cp "$DIR/web/assets/icons/app-icon.png" "$HOME/.local/share/icons/hicolor/${size}x${size}/apps/persian-calendar.png"
    fi
done

# Also install to pixmaps for legacy/universal desktop matching
mkdir -p "$HOME/.local/share/pixmaps"
cp "$DIR/web/assets/icons/app-icon.svg" "$HOME/.local/share/pixmaps/persian-calendar.svg"
if [ -f "$DIR/web/assets/icons/app-icon-256.png" ]; then
    cp "$DIR/web/assets/icons/app-icon-256.png" "$HOME/.local/share/pixmaps/persian-calendar.png"
fi

gtk-update-icon-cache -f -t "$HOME/.local/share/icons/hicolor" 2>/dev/null || true

ICON_PATH="$HOME/.local/share/icons/hicolor/scalable/apps/persian-calendar.svg"

mkdir -p "$HOME/.local/share/applications"
cat > "$HOME/.local/share/applications/persian-calendar.desktop" <<EOF
[Desktop Entry]
Name=تقویم جلالی
GenericName=تقویم شمسی و تایمر کار
Comment=تقویم خورشیدی به همراه زمان‌سنج کارکرد و مدیریت وظایف در نوار بالا
Exec=$HOME/.local/bin/persian-calendar
Icon=$ICON_PATH
Terminal=false
Type=Application
Categories=Utility;Clock;Calendar;Office;
StartupNotify=true
StartupWMClass=persian-calendar
EOF

chmod +x "$HOME/.local/share/applications/persian-calendar.desktop"
update-desktop-database "$HOME/.local/share/applications" 2>/dev/null || true

mkdir -p "$HOME/.config/autostart"
cat > "$HOME/.config/autostart/persian-calendar.desktop" <<EOF
[Desktop Entry]
Name=تقویم جلالی
Comment=اجرای خودکار تقویم فارسی و ثبت زمان کار در نوار بالای اوبونتو
Exec=$HOME/.local/bin/persian-calendar -no-window
Icon=$ICON_PATH
Terminal=false
Type=Application
X-GNOME-Autostart-enabled=true
EOF

chmod +x "$HOME/.config/autostart/persian-calendar.desktop"

echo "Installation complete!"
echo "• Binary installed to: $HOME/.local/bin/persian-calendar"
echo "• Desktop entry installed to: $HOME/.local/share/applications/persian-calendar.desktop"
echo "• Autostart on login enabled at: $HOME/.config/autostart/persian-calendar.desktop"
