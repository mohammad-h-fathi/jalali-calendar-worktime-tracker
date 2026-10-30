/**
 * App Main Controller & Realtime SSE Synchronizer
 */

// Helper: Convert ASCII digits to Persian digits
window.toPersianDigits = function (str) {
  if (str === null || str === undefined) return '';
  const persianDigits = ['۰', '۱', '۲', '۳', '۴', '۵', '۶', '۷', '۸', '۹'];
  return String(str).replace(/[0-9]/g, (w) => persianDigits[+w]);
};

// Helper: Format seconds to HH:MM:SS
window.formatSeconds = function (sec) {
  const s = Math.max(0, Math.floor(sec || 0));
  const hrs = Math.floor(s / 3600);
  const mins = Math.floor((s % 3600) / 60);
  const secs = s % 60;
  return `${String(hrs).padStart(2, '0')}:${String(mins).padStart(2, '0')}:${String(secs).padStart(2, '0')}`;
};

class AppController {
  constructor() {
    this.todayInfo = null;
    this.sse = null;
  }

  async start() {
    this.bindGlobalTabs();
    this.bindFloatingBtn();

    await this.fetchTodayInfo();

    // Initialize modules
    await window.calendarModule.init(this.todayInfo);
    await window.worktimeModule.init();
    await window.todoModule.init();

    this.connectSSE();
  }

  bindGlobalTabs() {
    document.querySelectorAll('.seg-btn').forEach((tab) => {
      tab.addEventListener('click', () => {
        this.switchTab(tab.dataset.tab);
      });
    });
  }

  switchTab(tabName) {
    document.querySelectorAll('.seg-btn').forEach((t) => t.classList.remove('active'));
    document.querySelectorAll('.tab-pane').forEach((p) => p.classList.remove('active'));

    const targetTab = document.querySelector(`.seg-btn[data-tab="${tabName}"]`);
    const targetPane = document.getElementById(`pane-${tabName}`);

    if (targetTab) targetTab.classList.add('active');
    if (targetPane) targetPane.classList.add('active');
  }

  bindFloatingBtn() {
    const pipBtn = document.getElementById('openPipBtn');
    if (!pipBtn) return;

    pipBtn.addEventListener('click', async () => {
      // 1. Try launching native borderless floating pill via backend
      try {
        const res = await fetch('/api/timer/floating/open', { method: 'POST' });
        if (res.ok) return;
      } catch (err) {}

      // 2. Fallback to browser Document Picture-in-Picture
      if (window.floatingTimer && window.floatingTimer.isSupported()) {
        window.floatingTimer.openFloatingWindow();
      }
    });
  }

  async fetchTodayInfo() {
    try {
      const res = await fetch('/api/today');
      this.todayInfo = await res.json();

      document.getElementById('currentFullDateLabel').textContent = `${this.todayInfo.weekday_name}، ${this.todayInfo.formatted}`;
    } catch (err) {
      console.error('Failed to load today info:', err);
      this.todayInfo = { year: 1405, month: 8, day: 4, iso: '1405-08-04', weekday_name: 'دوشنبه', formatted: '۴ آبان ۱۴۰۵' };
    }
  }

  connectSSE() {
    if (!window.EventSource) return;

    this.sse = new EventSource('/api/events/stream');

    this.sse.addEventListener('timer_update', (e) => {
      try {
        const payload = JSON.parse(e.data);
        if (window.worktimeModule) {
          window.worktimeModule.applyTimerState(payload.data.timer, payload.data.elapsed_seconds);
        }
      } catch (err) {}
    });

    this.sse.addEventListener('task_update', () => {
      if (window.todoModule) window.todoModule.loadTasks();
    });

    this.sse.addEventListener('task_deleted', () => {
      if (window.todoModule) window.todoModule.loadTasks();
    });

    this.sse.addEventListener('session_added', () => {
      if (window.worktimeModule) window.worktimeModule.loadDailySummary();
      if (window.calendarModule) {
        window.calendarModule.loadMonth(window.calendarModule.currentYear, window.calendarModule.currentMonth);
      }
    });

    this.sse.onerror = () => {};
  }
}

window.app = new AppController();
document.addEventListener('DOMContentLoaded', () => {
  window.app.start();
});
