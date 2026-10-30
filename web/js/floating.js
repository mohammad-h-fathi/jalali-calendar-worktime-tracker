/**
 * Floating Action Timer using Document Picture-in-Picture API
 * Creates a native Always-on-Top OS floating widget over all applications.
 */
class FloatingTimerManager {
  constructor() {
    this.pipWindow = null;
    this.timerData = {
      isRunning: false,
      isPaused: false,
      project: 'پروژه عمومی',
      elapsedSeconds: 0,
    };
  }

  isSupported() {
    return 'documentPictureInPicture' in window;
  }

  async openFloatingWindow() {
    if (!this.isSupported()) {
      alert('قابلیت پنجره همیشه رو (Picture-in-Picture) در این مرورگر پشتیبانی نمی‌شود.');
      return;
    }

    if (this.pipWindow) {
      this.pipWindow.focus();
      return;
    }

    try {
      this.pipWindow = await window.documentPictureInPicture.requestWindow({
        width: 270,
        height: 48,
      });

      this.pipWindow.addEventListener('pagehide', () => {
        this.pipWindow = null;
      });

      this.setupPiPContent();
    } catch (err) {
      console.error('Failed to open Picture-in-Picture window:', err);
    }
  }

  closeFloatingWindow() {
    if (this.pipWindow) {
      this.pipWindow.close();
      this.pipWindow = null;
    }
  }

  setupPiPContent() {
    if (!this.pipWindow) return;

    const doc = this.pipWindow.document;
    doc.title = 'تایمر کارکرد';
    doc.documentElement.dir = 'rtl';
    doc.documentElement.lang = 'fa';

    // Inject fonts & styles
    const fontLink = doc.createElement('link');
    fontLink.rel = 'stylesheet';
    fontLink.href = 'css/fonts.css';
    doc.head.appendChild(fontLink);

    const style = doc.createElement('style');
    style.textContent = `
      * { box-sizing: border-box; margin: 0; padding: 0; }
      body {
        font-family: 'Lahzeh', -apple-system, sans-serif;
        background: #090d16;
        color: #f8fafc;
        display: flex;
        align-items: center;
        justify-content: center;
        height: 100vh;
        overflow: hidden;
        user-select: none;
        padding: 8px;
      }
      .pip-capsule {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 10px;
        background: #141b29;
        border: 1px solid rgba(255, 255, 255, 0.15);
        border-radius: 999px;
        padding: 6px 14px;
        width: 100%;
        max-width: 290px;
        box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
      }
      .pip-meta {
        display: flex;
        align-items: center;
        gap: 6px;
      }
      .pip-dot {
        width: 8px;
        height: 8px;
        border-radius: 50%;
        background: #10b981;
      }
      .pip-dot.paused { background: #f59e0b; }
      .pip-project {
        font-size: 0.76rem;
        color: #94a3b8;
        font-weight: 600;
        max-width: 80px;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
      }
      .pip-clock {
        font-size: 1.25rem;
        font-weight: 800;
        color: #93c5fd;
        direction: ltr;
        letter-spacing: 0.5px;
      }
      .pip-actions {
        display: flex;
        align-items: center;
        gap: 5px;
      }
      .pip-btn {
        background: rgba(255, 255, 255, 0.1);
        border: none;
        color: #fff;
        border-radius: 999px;
        padding: 4px 10px;
        font-family: 'Lahzeh', sans-serif;
        font-size: 0.72rem;
        font-weight: 700;
        cursor: pointer;
      }
      .pip-btn:hover { background: rgba(255, 255, 255, 0.2); }
      .pip-btn-pause { background: #f59e0b; color: #000; }
      .pip-btn-resume { background: #10b981; }
      .pip-btn-stop { background: #f43f5e; }
    `;
    doc.head.appendChild(style);

    doc.body.innerHTML = `
      <div class="pip-capsule">
        <div class="pip-meta">
          <span class="pip-dot" id="pipDot"></span>
          <span class="pip-project" id="pipProjectName">${this.timerData.project || 'پروژه عمومی'}</span>
        </div>
        <div class="pip-clock" id="pipClockDigits">۰۰:۰۰:۰۰</div>
        <div class="pip-actions">
          <button class="pip-btn pip-btn-pause" id="pipPauseResumeBtn">توقف</button>
          <button class="pip-btn pip-btn-stop" id="pipStopBtn">پایان</button>
        </div>
      </div>
    `;

    const pauseResumeBtn = doc.getElementById('pipPauseResumeBtn');
    const stopBtn = doc.getElementById('pipStopBtn');

    pauseResumeBtn.addEventListener('click', () => {
      if (this.timerData.isPaused) {
        window.worktimeModule.resumeTimer();
      } else {
        window.worktimeModule.pauseTimer();
      }
    });

    stopBtn.addEventListener('click', async () => {
      await window.worktimeModule.stopTimer();
      this.closeFloatingWindow();
    });

    this.updatePiPView();
  }

  updateState(timer, elapsedSeconds) {
    if (!timer || !timer.is_running) {
      this.closeFloatingWindow();
      return;
    }
    this.timerData = {
      isRunning: timer.is_running,
      isPaused: timer.is_paused,
      project: timer.project || 'پروژه عمومی',
      elapsedSeconds: elapsedSeconds,
    };
    this.updatePiPView();
  }

  updatePiPView() {
    if (!this.pipWindow) return;

    const doc = this.pipWindow.document;
    const projectElem = doc.getElementById('pipProjectName');
    const clockElem = doc.getElementById('pipClockDigits');
    const pauseResumeBtn = doc.getElementById('pipPauseResumeBtn');
    const dotElem = doc.getElementById('pipDot');

    if (projectElem) {
      projectElem.textContent = this.timerData.project;
    }

    if (clockElem) {
      clockElem.textContent = window.toPersianDigits(window.formatSeconds(this.timerData.elapsedSeconds));
    }

    if (dotElem) {
      dotElem.className = this.timerData.isPaused ? 'pip-dot paused' : 'pip-dot';
    }

    if (pauseResumeBtn) {
      if (this.timerData.isPaused) {
        pauseResumeBtn.textContent = 'ادامه ▶';
        pauseResumeBtn.className = 'pip-btn pip-btn-resume';
      } else {
        pauseResumeBtn.textContent = 'توقف ⏸';
        pauseResumeBtn.className = 'pip-btn pip-btn-pause';
      }
    }
  }
}

window.floatingTimer = new FloatingTimerManager();
