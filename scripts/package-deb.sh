#!/usr/bin/env bash
set -e

# Usage: ./scripts/package-deb.sh <version> <arch> <binary_path> <output_dir>
VERSION="${1:-1.0.0}"
ARCH="${2:-amd64}"
BINARY="${3:-bin/persian-calendar}"
OUTPUT_DIR="${4:-dist}"

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BUILD_ROOT="/tmp/persian-calendar-deb-$$"

mkdir -p "$BUILD_ROOT/DEBIAN"
mkdir -p "$BUILD_ROOT/usr/bin"
mkdir -p "$BUILD_ROOT/usr/share/applications"
mkdir -p "$BUILD_ROOT/usr/share/icons/hicolor/scalable/apps"
mkdir -p "$BUILD_ROOT/usr/share/fonts/truetype/lahzeh"

# 1. Binary & Pill helper
install -m 755 "$BINARY" "$BUILD_ROOT/usr/bin/persian-calendar"
install -m 755 "$DIR/scripts/floating_timer.py" "$BUILD_ROOT/usr/bin/persian-timer-pill"

# 2. Desktop entry
cat > "$BUILD_ROOT/usr/share/applications/persian-calendar.desktop" <<EOF
[Desktop Entry]
Name=تقویم جلالی
GenericName=تقویم شمسی و تایمر کار
Comment=تقویم خورشیدی به همراه زمان‌سنج کارکرد و مدیریت وظایف در نوار بالا
Exec=/usr/bin/persian-calendar
Icon=persian-calendar
Terminal=false
Type=Application
Categories=Utility;Clock;Calendar;Office;
StartupNotify=true
StartupWMClass=persian-calendar
EOF
chmod 644 "$BUILD_ROOT/usr/share/applications/persian-calendar.desktop"

# 3. Icon
cp "$DIR/web/assets/icons/app-icon.svg" "$BUILD_ROOT/usr/share/icons/hicolor/scalable/apps/persian-calendar.svg"

# 4. Lahzeh fonts
cp "$DIR"/fonts/*.ttf "$BUILD_ROOT/usr/share/fonts/truetype/lahzeh/" 2>/dev/null || true

# 5. DEBIAN/control
cat > "$BUILD_ROOT/DEBIAN/control" <<EOF
Package: persian-calendar
Version: ${VERSION#v}
Section: utils
Priority: optional
Architecture: $ARCH
Depends: libc6
Recommends: python3-gi, gir1.2-gtk-3.0
Maintainer: Persian Calendar Team
Description: Persian (Solar Hijri) Calendar & Worktime Tracker
 Native Iranian calendar widget for Linux with live top-bar indicator,
 task management, worktime punch clock, and borderless floating action timer.
EOF

# 6. Post-install script to refresh icon & font caches
cat > "$BUILD_ROOT/DEBIAN/postinst" <<'EOF'
#!/bin/sh
set -e
if [ -x /usr/bin/fc-cache ]; then
    fc-cache -f /usr/share/fonts/truetype/lahzeh || true
fi
if [ -x /usr/bin/update-icon-caches ]; then
    update-icon-caches /usr/share/icons/hicolor || true
fi
exit 0
EOF
chmod 755 "$BUILD_ROOT/DEBIAN/postinst"

# 7. Build package
mkdir -p "$OUTPUT_DIR"
DEB_NAME="persian-calendar_${VERSION#v}_${ARCH}.deb"
dpkg-deb --build --root-owner-group "$BUILD_ROOT" "$OUTPUT_DIR/$DEB_NAME"

rm -rf "$BUILD_ROOT"
echo "Successfully created: $OUTPUT_DIR/$DEB_NAME"
