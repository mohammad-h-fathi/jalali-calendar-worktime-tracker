/**
 * Worktime module: Punch clock, interactive project chips, compact bar sync
 */
class WorktimeModule {
  constructor() {
    this.projects = [];
    this.selectedProject = 'پروژه عمومی';
    this.selectedColor = '#3b82f6';
    this.timer = {
      is_running: false,
      is_paused: false,
      project: 'پروژه عمومی',
      elapsed_seconds: 0,
    };
    this.localTickInterval = null;
  }

  async init() {
    this.bindEvents();
    await this.loadProjects();
    await this.refreshTimerStatus();
    await this.loadDailySummary();
  }

  bindEvents() {
    document.getElementById('timerStartBtn').addEventListener('click', () => this.startTimer());
    document.getElementById('timerPauseBtn').addEventListener('click', () => this.pauseTimer());
    document.getElementById('timerResumeBtn').addEventListener('click', () => this.resumeTimer());
    document.getElementById('timerStopBtn').addEventListener('click', () => this.stopTimer());

    // Add Project Modal
    document.getElementById('addNewProjectBtn').addEventListener('click', () => {
      document.getElementById('addProjectModal').style.display = 'flex';
      document.getElementById('newProjectName').focus();
    });
    document.getElementById('closeProjectModalBtn').addEventListener('click', () => {
      document.getElementById('addProjectModal').style.display = 'none';
    });
    document.getElementById('cancelProjectBtn').addEventListener('click', () => {
      document.getElementById('addProjectModal').style.display = 'none';
    });
    document.getElementById('addProjectForm').addEventListener('submit', (e) => {
      e.preventDefault();
      this.createProject();
    });

    // Manual Log Modal
    document.getElementById('manualLogModalBtn').addEventListener('click', () => {
      document.getElementById('manualDate').value = window.calendarModule.todayDate ? window.calendarModule.todayDate.iso : '';
      this.populateManualProjectSelect();
      document.getElementById('manualLogModal').style.display = 'flex';
    });
    document.getElementById('closeManualModalBtn').addEventListener('click', () => {
      document.getElementById('manualLogModal').style.display = 'none';
    });
    document.getElementById('cancelManualBtn').addEventListener('click', () => {
      document.getElementById('manualLogModal').style.display = 'none';
    });
    document.getElementById('manualLogForm').addEventListener('submit', (e) => {
      e.preventDefault();
      this.saveManualSession();
    });
  }

  async loadProjects() {
    try {
      const res = await fetch('/api/projects');
      this.projects = await res.json();
      this.renderProjectChips();
    } catch (err) {
      console.error('Failed to load projects:', err);
    }
  }

  renderProjectChips() {
    const container = document.getElementById('projectChipsList');
    if (!container) return;
    container.innerHTML = '';

    if (this.projects.length === 0) {
      this.projects = [{ id: '1', name: 'پروژه عمومی', color: '#3b82f6' }];
    }

    this.projects.forEach(p => {
      const chip = document.createElement('button');
      chip.type = 'button';
      chip.className = `chip-btn ${p.name === this.selectedProject ? 'active' : ''}`;
      chip.innerHTML = `
        <span class="chip-dot" style="background-color: ${p.color || '#3b82f6'};"></span>
        <span>${p.name}</span>
      `;
      chip.addEventListener('click', () => {
        this.selectedProject = p.name;
        this.selectedColor = p.color || '#3b82f6';
        document.querySelectorAll('.chip-btn').forEach(c => c.classList.remove('active'));
        chip.classList.add('active');
        document.getElementById('headerTimerProject').textContent = this.selectedProject;
      });
      container.appendChild(chip);
    });

    if (!this.selectedProject && this.projects.length > 0) {
      this.selectedProject = this.projects[0].name;
      this.selectedColor = this.projects[0].color;
    }
    document.getElementById('headerTimerProject').textContent = this.selectedProject;
  }

  populateManualProjectSelect() {
    const sel = document.getElementById('manualProjectSelect');
    sel.innerHTML = '';
    this.projects.forEach(p => {
      const opt = document.createElement('option');
      opt.value = p.name;
      opt.textContent = p.name;
      opt.dataset.color = p.color;
      sel.appendChild(opt);
    });
  }

  async createProject() {
    const name = document.getElementById('newProjectName').value.trim();
    const color = document.getElementById('newProjectColor').value;
    if (!name) return;

    try {
      await fetch('/api/projects', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name, color }),
      });
      document.getElementById('addProjectModal').style.display = 'none';
      document.getElementById('newProjectName').value = '';
      this.selectedProject = name;
      this.selectedColor = color;
      await this.loadProjects();
    } catch (err) {
      console.error('Failed to create project:', err);
    }
  }

  async refreshTimerStatus() {
    try {
      const res = await fetch('/api/timer');
      const data = await res.json();
      this.applyTimerState(data.timer, data.elapsed_seconds);
    } catch (err) {
      console.error('Failed to get timer:', err);
    }
  }

  async startTimer() {
    const desc = document.getElementById('timerDescInput').value.trim();

    try {
      const res = await fetch('/api/timer/start', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          project: this.selectedProject || 'پروژه عمومی',
          description: desc,
          color: this.selectedColor || '#3b82f6',
        }),
      });
      if (res.ok) {
        const timer = await res.json();
        this.applyTimerState(timer, 0);
      }
    } catch (err) {
      console.error('Failed to start timer:', err);
    }
  }

  async pauseTimer() {
    try {
      const res = await fetch('/api/timer/pause', { method: 'POST' });
      if (res.ok) {
        const timer = await res.json();
        this.applyTimerState(timer, this.timer.elapsed_seconds);
      }
    } catch (err) {
      console.error('Failed to pause timer:', err);
    }
  }

  async resumeTimer() {
    try {
      const res = await fetch('/api/timer/resume', { method: 'POST' });
      if (res.ok) {
        const timer = await res.json();
        this.applyTimerState(timer, this.timer.elapsed_seconds);
      }
    } catch (err) {
      console.error('Failed to resume timer:', err);
    }
  }

  async stopTimer() {
    try {
      const res = await fetch('/api/timer/stop', { method: 'POST' });
      if (res.ok) {
        this.applyTimerState({ is_running: false }, 0);
        await this.loadDailySummary();
        if (window.calendarModule) {
          window.calendarModule.loadMonth(window.calendarModule.currentYear, window.calendarModule.currentMonth);
        }
      }
    } catch (err) {
      console.error('Failed to stop timer:', err);
    }
  }

  async saveManualSession() {
    const jalaliDate = document.getElementById('manualDate').value;
    const project = document.getElementById('manualProjectSelect').value;
    const startTimeStr = document.getElementById('manualStartTime').value;
    const endTimeStr = document.getElementById('manualEndTime').value;
    const desc = document.getElementById('manualDesc').value.trim();

    if (!jalaliDate) {
      alert('لطفاً تاریخ را وارد کنید');
      return;
    }
    if (!startTimeStr || !endTimeStr) {
      alert('لطفاً زمان شروع و پایان را وارد کنید');
      return;
    }

    const [sh, sm] = startTimeStr.split(':').map(Number);
    const [eh, em] = endTimeStr.split(':').map(Number);

    const startMinutes = sh * 60 + sm;
    const endMinutes = eh * 60 + em;

    if (endMinutes <= startMinutes) {
      alert('زمان پایان باید بعد از زمان شروع باشد');
      return;
    }

    const durationSec = (endMinutes - startMinutes) * 60;

    try {
      const res = await fetch('/api/worktime/manual', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          project: project,
          description: desc,
          jalali_date: jalaliDate,
          start_time: startTimeStr,
          end_time: endTimeStr,
          duration_seconds: durationSec,
          color: color,
        }),
      });

      if (!res.ok) {
        const errTxt = await res.text();
        alert(errTxt || 'خطا در ثبت سشن دستی');
        return;
      }

      document.getElementById('manualLogModal').style.display = 'none';
      document.getElementById('manualDesc').value = '';
      await this.loadDailySummary();
      if (window.calendarModule) {
        window.calendarModule.loadMonth(window.calendarModule.currentYear, window.calendarModule.currentMonth);
      }
    } catch (err) {
      console.error('Failed to save manual session:', err);
    }
  }

  applyTimerState(timer, elapsedSeconds) {
    this.timer = {
      ...timer,
      elapsed_seconds: elapsedSeconds || 0,
    };

    const headerClock = document.getElementById('headerTimerClock');
    const headerProject = document.getElementById('headerTimerProject');
    const headerDesc = document.getElementById('headerTimerDesc');
    const globalDot = document.getElementById('globalTimerDot');

    const heroDigits = document.getElementById('heroTimerDigits');
    const heroStatus = document.getElementById('heroTimerStatus');

    const startBtn = document.getElementById('timerStartBtn');
    const pauseBtn = document.getElementById('timerPauseBtn');
    const resumeBtn = document.getElementById('timerResumeBtn');
    const stopBtn = document.getElementById('timerStopBtn');

    if (this.timer.is_running) {
      headerProject.textContent = this.timer.project;
      headerClock.textContent = window.toPersianDigits(window.formatSeconds(this.timer.elapsed_seconds));

      startBtn.style.display = 'none';
      stopBtn.style.display = 'inline-flex';

      if (this.timer.is_paused) {
        globalDot.className = 'timer-status-dot paused';
        headerDesc.textContent = 'متوقف شده';
        if (heroStatus) heroStatus.textContent = `متوقف: ${this.timer.project}`;
        pauseBtn.style.display = 'none';
        resumeBtn.style.display = 'inline-flex';
      } else {
        globalDot.className = 'timer-status-dot';
        headerDesc.textContent = this.timer.description || 'در حال کار...';
        if (heroStatus) heroStatus.textContent = `در حال ثبت: ${this.timer.project}`;
        pauseBtn.style.display = 'inline-flex';
        resumeBtn.style.display = 'none';
      }
    } else {
      globalDot.className = 'timer-status-dot idle';
      headerProject.textContent = this.selectedProject || 'پروژه عمومی';
      headerDesc.textContent = 'آماده شروع';
      if (heroStatus) heroStatus.textContent = 'آماده شروع کار';

      startBtn.style.display = 'inline-flex';
      pauseBtn.style.display = 'none';
      resumeBtn.style.display = 'none';
      stopBtn.style.display = 'none';
    }

    const formattedTime = window.toPersianDigits(window.formatSeconds(this.timer.elapsed_seconds));
    headerClock.textContent = formattedTime;
    if (heroDigits) heroDigits.textContent = formattedTime;

    // Update Floating PiP window
    if (window.floatingTimer) {
      window.floatingTimer.updateState(this.timer, this.timer.elapsed_seconds);
    }

    // Local ticker
    clearInterval(this.localTickInterval);
    if (this.timer.is_running && !this.timer.is_paused) {
      this.localTickInterval = setInterval(() => {
        this.timer.elapsed_seconds++;
        const formatted = window.toPersianDigits(window.formatSeconds(this.timer.elapsed_seconds));
        headerClock.textContent = formatted;
        if (heroDigits) heroDigits.textContent = formatted;

        if (window.floatingTimer) {
          window.floatingTimer.updateState(this.timer, this.timer.elapsed_seconds);
        }
      }, 1000);
    }
  }

  async loadDailySummary() {
    const iso = window.calendarModule && window.calendarModule.todayDate ? window.calendarModule.todayDate.iso : '';
    try {
      const res = await fetch(`/api/worktime/daily?date=${iso}`);
      const data = await res.json();

      const hours = Math.floor(data.total_seconds / 3600);
      const mins = Math.floor((data.total_seconds % 3600) / 60);
      const summaryText = `${window.toPersianDigits(String(hours))} ساعت و ${window.toPersianDigits(String(mins))} دقیقه`;
      document.getElementById('worktimeTodayTotal').textContent = summaryText;

      this.renderBreakdown(data.breakdown, data.total_seconds);
      this.renderSessionsTable(data.sessions);
    } catch (err) {
      console.error('Failed to load daily summary:', err);
    }
  }

  renderBreakdown(breakdown, totalSec) {
    const container = document.getElementById('projectBreakdownWrapper');
    if (!container) return;
    if (!breakdown || breakdown.length === 0 || totalSec === 0) {
      container.innerHTML = '<div class="empty-text">کارکردی برای تفکیک پروژه ثبت نشده است</div>';
      return;
    }

    container.innerHTML = breakdown.map(item => {
      const pct = Math.round((item.duration_seconds / totalSec) * 100);
      const hrs = (item.duration_seconds / 3600).toFixed(1);
      return `
        <div style="display:flex; justify-content:space-between; font-size:0.75rem; color:#cbd5e1; margin-bottom:4px;">
          <span style="color:${item.color || '#3b82f6'}; font-weight:700;">${item.project}</span>
          <span>${window.toPersianDigits(hrs)}س (${window.toPersianDigits(String(pct))}٪)</span>
        </div>
      `;
    }).join('');
  }

  renderSessionsTable(sessions) {
    const tbody = document.getElementById('workSessionsTableBody');
    if (!tbody) return;
    if (!sessions || sessions.length === 0) {
      tbody.innerHTML = '<tr><td colspan="4" class="empty-text">هنوز سشنی برای امروز ثبت نشده است</td></tr>';
      return;
    }

    tbody.innerHTML = sessions.map(s => {
      const start = new Date(s.start_time).toLocaleTimeString('fa-IR', { hour: '2-digit', minute: '2-digit' });
      const end = new Date(s.end_time).toLocaleTimeString('fa-IR', { hour: '2-digit', minute: '2-digit' });
      return `
        <tr>
          <td><strong style="color:${s.color || '#3b82f6'};">${s.project}</strong></td>
          <td style="direction:ltr; text-align:right;">${start}-${end}</td>
          <td><strong style="color:#6ee7b7;">${window.toPersianDigits(window.formatSeconds(s.duration_seconds))}</strong></td>
          <td>
            <button class="inline-add-btn" style="color:#f43f5e;" onclick="window.worktimeModule.deleteSession('${s.id}')">×</button>
          </td>
        </tr>
      `;
    }).join('');
  }

  async deleteSession(id) {
    if (!confirm('آیا از حذف این سشن اطمینان دارید؟')) return;
    try {
      const res = await fetch(`/api/worktime/sessions/${encodeURIComponent(id)}`, {
        method: 'DELETE',
      });
      if (res.ok) {
        await this.loadDailySummary();
        if (window.calendarModule) {
          window.calendarModule.loadMonth(window.calendarModule.currentYear, window.calendarModule.currentMonth);
        }
      } else {
        const errTxt = await res.text();
        alert(errTxt || 'خطا در حذف سشن');
      }
    } catch (err) {
      console.error('Failed to delete session:', err);
    }
  }
}

window.worktimeModule = new WorktimeModule();
