#!/usr/bin/env python3
"""
Floating Timer Pill for Persian Calendar
A borderless, always-on-top, draggable floating pill widget.
No titlebar, no close/minimize buttons.
Automatically closes when 'Stop' is clicked or when timer ends.
"""

import sys
import os
import json
import urllib.request
import urllib.error
import gi

gi.require_version('Gtk', '3.0')
gi.require_version('Gdk', '3.0')
from gi.repository import Gtk, Gdk, GLib, Pango

SERVER_BASE = os.environ.get("PERSIAN_CAL_URL", "http://127.0.0.1:41890")

PERSIAN_DIGITS = {'0':'۰', '1':'۱', '2':'۲', '3':'۳', '4':'۴', '5':'۵', '6':'۶', '7':'۷', '8':'۸', '9':'۹'}

def to_persian_digits(s):
    return ''.join(PERSIAN_DIGITS.get(ch, ch) for ch in str(s))

def format_seconds(sec):
    h = sec // 3600
    m = (sec % 3600) // 60
    s = sec % 60
    return f"{h:02d}:{m:02d}:{s:02d}"

CSS_STYLES = b"""
window.pill-window {
    background: transparent;
}

#capsule {
    background-color: rgba(13, 18, 28, 0.94);
    border: 1px solid rgba(255, 255, 255, 0.16);
    border-radius: 24px;
    padding: 6px 14px;
    box-shadow: 0 8px 28px rgba(0, 0, 0, 0.65);
}

#capsule:hover {
    border-color: rgba(59, 130, 246, 0.5);
}

#project-label {
    font-family: 'Lahzeh', -apple-system, sans-serif;
    font-size: 11px;
    font-weight: 700;
    color: #94a3b8;
    margin-right: 6px;
}

#clock-label {
    font-family: 'Lahzeh', -apple-system, sans-serif;
    font-size: 15px;
    font-weight: 800;
    color: #93c5fd;
    letter-spacing: 0.5px;
    margin-left: 8px;
    margin-right: 8px;
}

.status-dot {
    background-color: #10b981;
    border-radius: 4px;
    min-width: 8px;
    min-height: 8px;
}

.status-dot.paused {
    background-color: #f59e0b;
}

.status-dot.idle {
    background-color: #64748b;
}

.pill-btn {
    font-family: 'Lahzeh', -apple-system, sans-serif;
    font-size: 11px;
    font-weight: 700;
    padding: 4px 10px;
    border-radius: 14px;
    border: none;
    outline: none;
    transition: all 150ms ease;
}

.btn-pause {
    background-color: rgba(245, 158, 11, 0.2);
    color: #fbbf24;
    border: 1px solid rgba(245, 158, 11, 0.4);
}
.btn-pause:hover {
    background-color: #f59e0b;
    color: #000;
}

.btn-resume {
    background-color: rgba(16, 185, 129, 0.2);
    color: #34d399;
    border: 1px solid rgba(16, 185, 129, 0.4);
}
.btn-resume:hover {
    background-color: #10b981;
    color: #fff;
}

.btn-stop {
    background-color: rgba(244, 63, 94, 0.2);
    color: #fb7185;
    border: 1px solid rgba(244, 63, 94, 0.4);
}
.btn-stop:hover {
    background-color: #f43f5e;
    color: #fff;
}
"""

def safe_quit():
    try:
        while Gtk.main_level() > 0:
            Gtk.main_quit()
    except Exception:
        pass
    sys.exit(0)

class FloatingPill(Gtk.Window):
    def __init__(self):
        super().__init__(type=Gtk.WindowType.TOPLEVEL)
        
        self.set_title("تایمر شناور")
        self.set_decorated(False)
        self.set_keep_above(True)
        self.set_skip_taskbar_hint(True)
        self.set_skip_pager_hint(True)
        self.set_resizable(False)
        self.set_app_paintable(True)
        self.get_style_context().add_class("pill-window")

        # RGBA for transparency
        screen = Gdk.Screen.get_default()
        if screen:
            visual = screen.get_rgba_visual()
            if visual:
                self.set_visual(visual)

        self.is_running = False
        self.is_paused = False
        self.elapsed = 0
        self.project_name = "پروژه عمومی"
        self.failed_polls = 0

        # Build UI
        self.event_box = Gtk.EventBox()
        self.event_box.set_visible_window(False)
        self.event_box.connect("button-press-event", self.on_button_press)
        self.add(self.event_box)

        self.capsule = Gtk.Box(orientation=Gtk.Orientation.HORIZONTAL, spacing=8)
        self.capsule.set_name("capsule")
        self.capsule.set_direction(Gtk.TextDirection.RTL)
        self.event_box.add(self.capsule)

        # Status Dot
        self.dot = Gtk.Box()
        self.dot.get_style_context().add_class("status-dot")
        self.dot.set_valign(Gtk.Align.CENTER)
        self.capsule.pack_start(self.dot, False, False, 0)

        # Project Label
        self.lbl_project = Gtk.Label(label="پروژه عمومی")
        self.lbl_project.set_name("project-label")
        self.lbl_project.set_ellipsize(Pango.EllipsizeMode.END)
        self.lbl_project.set_max_width_chars(12)
        self.capsule.pack_start(self.lbl_project, False, False, 0)

        # Separator / Clock
        self.lbl_clock = Gtk.Label(label="۰۰:۰۰:۰۰")
        self.lbl_clock.set_name("clock-label")
        self.lbl_clock.set_direction(Gtk.TextDirection.LTR)
        self.capsule.pack_start(self.lbl_clock, False, False, 4)

        # Action Buttons
        self.btn_pause_resume = Gtk.Button(label="توقف")
        self.btn_pause_resume.get_style_context().add_class("pill-btn")
        self.btn_pause_resume.get_style_context().add_class("btn-pause")
        self.btn_pause_resume.connect("clicked", self.on_toggle_pause)
        self.capsule.pack_start(self.btn_pause_resume, False, False, 0)

        self.btn_stop = Gtk.Button(label="پایان")
        self.btn_stop.get_style_context().add_class("pill-btn")
        self.btn_stop.get_style_context().add_class("btn-stop")
        self.btn_stop.connect("clicked", self.on_stop_clicked)
        self.capsule.pack_start(self.btn_stop, False, False, 0)

        # Position at top-right by default
        self.set_default_size(270, 44)
        if screen:
            disp = screen.get_display()
            if disp and disp.get_n_monitors() > 0:
                monitor = disp.get_primary_monitor() or disp.get_monitor(0)
                if monitor:
                    geom = monitor.get_geometry()
                    self.move(geom.x + geom.width - 340, geom.y + 60)

        # Apply CSS
        css_provider = Gtk.CssProvider()
        css_provider.load_from_data(CSS_STYLES)
        Gtk.StyleContext.add_provider_for_screen(
            screen,
            css_provider,
            Gtk.STYLE_PROVIDER_PRIORITY_APPLICATION
        )

        # Start Polling
        self.poll_timer()
        GLib.timeout_add(1000, self.poll_timer)

    def on_button_press(self, widget, event):
        if event.button == 1:
            # Left click drag
            self.begin_move_drag(event.button, int(event.x_root), int(event.y_root), event.time)
        return True

    def poll_timer(self):
        try:
            req = urllib.request.Request(f"{SERVER_BASE}/api/timer")
            with urllib.request.urlopen(req, timeout=1.5) as resp:
                data = json.loads(resp.read().decode())
                timer = data.get("timer", {})
                self.elapsed = data.get("elapsed_seconds", 0)
                self.is_running = timer.get("is_running", False)
                self.is_paused = timer.get("is_paused", False)
                self.project_name = timer.get("project", "پروژه عمومی")

                self.failed_polls = 0
                self.update_view()
        except Exception as e:
            self.failed_polls += 1
            if self.failed_polls >= 5:
                print("Backend server unreachable. Closing floating pill.")
                safe_quit()
                return False
        return True

    def update_view(self):
        ctx_dot = self.dot.get_style_context()
        ctx_btn = self.btn_pause_resume.get_style_context()

        if not self.is_running:
            self.lbl_project.set_text("آماده کار")
            self.lbl_clock.set_text(to_persian_digits("۰۰:۰۰:۰۰"))
            if not ctx_dot.has_class("idle"):
                ctx_dot.add_class("idle")
            if ctx_dot.has_class("paused"):
                ctx_dot.remove_class("paused")

            self.btn_pause_resume.set_label("شروع ▶")
            if ctx_btn.has_class("btn-pause"):
                ctx_btn.remove_class("btn-pause")
            if not ctx_btn.has_class("btn-resume"):
                ctx_btn.add_class("btn-resume")

            self.btn_stop.set_label("بستن ✕")
            return

        self.lbl_project.set_text(self.project_name or "پروژه عمومی")
        time_str = format_seconds(self.elapsed)
        self.lbl_clock.set_text(to_persian_digits(time_str))
        self.btn_stop.set_label("پایان")

        if ctx_dot.has_class("idle"):
            ctx_dot.remove_class("idle")

        if self.is_paused:
            if not ctx_dot.has_class("paused"):
                ctx_dot.add_class("paused")
            self.btn_pause_resume.set_label("ادامه ▶")
            if ctx_btn.has_class("btn-pause"):
                ctx_btn.remove_class("btn-pause")
            if not ctx_btn.has_class("btn-resume"):
                ctx_btn.add_class("btn-resume")
        else:
            if ctx_dot.has_class("paused"):
                ctx_dot.remove_class("paused")
            self.btn_pause_resume.set_label("توقف ⏸")
            if ctx_btn.has_class("btn-resume"):
                ctx_btn.remove_class("btn-resume")
            if not ctx_btn.has_class("btn-pause"):
                ctx_btn.add_class("btn-pause")

    def on_toggle_pause(self, widget):
        if not self.is_running:
            try:
                payload = json.dumps({"project": "پروژه عمومی"}).encode('utf-8')
                req = urllib.request.Request(
                    f"{SERVER_BASE}/api/timer/start",
                    data=payload,
                    headers={"Content-Type": "application/json"},
                    method="POST"
                )
                with urllib.request.urlopen(req, timeout=2.0) as resp:
                    pass
                self.poll_timer()
            except Exception as e:
                print("Failed to start timer:", e)
            return

        endpoint = "/api/timer/resume" if self.is_paused else "/api/timer/pause"
        try:
            req = urllib.request.Request(f"{SERVER_BASE}{endpoint}", method="POST")
            with urllib.request.urlopen(req, timeout=2.0) as resp:
                pass
            self.poll_timer()
        except Exception as e:
            print("Failed to toggle pause:", e)

    def on_stop_clicked(self, widget):
        if self.is_running:
            print("Stop clicked! Sending stop request and removing floating pill.")
            try:
                req = urllib.request.Request(f"{SERVER_BASE}/api/timer/stop", method="POST")
                with urllib.request.urlopen(req, timeout=2.0) as resp:
                    pass
            except Exception as e:
                print("Failed to stop timer:", e)
        self.destroy()
        safe_quit()

if __name__ == "__main__":
    win = FloatingPill()
    win.show_all()
    Gtk.main()
