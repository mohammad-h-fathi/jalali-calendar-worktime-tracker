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
    this.projectsSummary = null;
    this.timelineMode = 'selected';
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

    const heroStart = document.getElementById('heroTimerStartBtn');
    const heroPause = document.getElementById('heroTimerPauseBtn');
    const heroResume = document.getElementById('heroTimerResumeBtn');
    const heroStop = document.getElementById('heroTimerStopBtn');
    if (heroStart) heroStart.addEventListener('click', () => this.startTimer());
    if (heroPause) heroPause.addEventListener('click', () => this.pauseTimer());
    if (heroResume) heroResume.addEventListener('click', () => this.resumeTimer());
    if (heroStop) heroStop.addEventListener('click', () => this.stopTimer());

    const filterSelectedBtn = document.getElementById('timelineFilterSelectedBtn');
    const filterAllBtn = document.getElementById('timelineFilterAllBtn');
    if (filterSelectedBtn) {
      filterSelectedBtn.addEventListener('click', () => this.setTimelineMode('selected'));
    }
    if (filterAllBtn) {
      filterAllBtn.addEventListener('click', () => this.setTimelineMode('all'));
    }

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

  getProjectColor(name) {
    const proj = this.projects.find(p => p.name === name);
    return (proj && proj.color) ? proj.color : '#3b82f6';
  }

  syncSelectedChip() {
    document.querySelectorAll('#projectChipsList .chip-btn').forEach(chip => {
      if (chip.dataset.project === this.selectedProject) {
        chip.classList.add('active');
      } else {
        chip.classList.remove('active');
      }
    });
  }

  updateTimerActionControls() {
    const startBtn = document.getElementById('timerStartBtn');
    const heroStartBtn = document.getElementById('heroTimerStartBtn');
    const heroStartLabel = document.getElementById('heroTimerStartLabel');
    const heroActiveRow = document.getElementById('heroTimerActiveRow');
    const heroPauseBtn = document.getElementById('heroTimerPauseBtn');
    const heroResumeBtn = document.getElementById('heroTimerResumeBtn');
    const headerProject = document.getElementById('headerTimerProject');

    const selected = this.selectedProject || 'پروژه عمومی';

    if (this.timer.is_running) {
      const differentProject = selected !== this.timer.project;
      if (startBtn) startBtn.style.display = differentProject ? 'inline-flex' : 'none';
      if (heroStartBtn) heroStartBtn.style.display = differentProject ? 'inline-flex' : 'none';
      if (heroStartLabel) {
        heroStartLabel.textContent = `تغییر و شروع تایمر «${selected}»`;
      }
      if (heroActiveRow) heroActiveRow.style.display = 'grid';
      if (heroPauseBtn) heroPauseBtn.style.display = this.timer.is_paused ? 'none' : 'inline-flex';
      if (heroResumeBtn) heroResumeBtn.style.display = this.timer.is_paused ? 'inline-flex' : 'none';
    } else {
      if (headerProject) headerProject.textContent = selected;
      if (startBtn) startBtn.style.display = 'inline-flex';
      if (heroStartBtn) heroStartBtn.style.display = 'inline-flex';
      if (heroStartLabel) {
        heroStartLabel.textContent = `شروع تایمر «${selected}»`;
      }
      if (heroActiveRow) heroActiveRow.style.display = 'none';
    }
  }

  async selectProject(name, color, persist = true) {
    this.selectedProject = name || 'پروژه عمومی';
    this.selectedColor = color || this.getProjectColor(this.selectedProject);
    this.syncSelectedChip();
    this.updateTimerActionControls();
    this.renderProjectsTotals();
    this.renderProjectTimeline();

    if (persist && !this.timer.is_running) {
      try {
        await fetch('/api/timer/project', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            project: this.selectedProject,
            color: this.selectedColor,
          }),
        });
      } catch (err) {
        console.error('Failed to persist selected project:', err);
      }
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
      chip.dataset.project = p.name;
      chip.className = `chip-btn ${p.name === this.selectedProject ? 'active' : ''}`;
      chip.innerHTML = `
        <span class="chip-dot" style="background-color: ${p.color || '#3b82f6'};"></span>
        <span>${p.name}</span>
      `;
      chip.addEventListener('click', () => {
        this.selectProject(p.name, p.color || '#3b82f6', true);
      });
      container.appendChild(chip);
    });

    if (!this.selectedProject && this.projects.length > 0) {
      this.selectedProject = this.projects[0].name;
      this.selectedColor = this.projects[0].color;
    }
    this.updateTimerActionControls();
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
      await this.loadProjects();
      await this.loadProjectsSummary();
      await this.selectProject(name, color, true);
    } catch (err) {
      console.error('Failed to create project:', err);
    }
  }

  async refreshTimerStatus() {
    try {
      const res = await fetch('/api/timer');
      const data = await res.json();
      if (data.timer && data.timer.project) {
        this.selectedProject = data.timer.project;
        this.selectedColor = data.timer.color || this.getProjectColor(data.timer.project);
      }
      this.applyTimerState(data.timer, data.elapsed_seconds, true);
    } catch (err) {
      console.error('Failed to get timer:', err);
    }
  }

  async startTimer() {
    const desc = document.getElementById('timerDescInput').value.trim();
    const wasRunning = this.timer.is_running;
    const project = this.selectedProject || 'پروژه عمومی';
    const color = this.selectedColor || this.getProjectColor(project);

    try {
      const res = await fetch('/api/timer/start', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          project: project,
          description: desc,
          color: color,
        }),
      });
      if (res.ok) {
        const timer = await res.json();
        this.applyTimerState(timer, 0, true);
        if (wasRunning) {
          await this.loadDailySummary();
          if (window.calendarModule) {
            window.calendarModule.loadMonth(window.calendarModule.currentYear, window.calendarModule.currentMonth);
          }
        }
      } else {
        const errTxt = await res.text();
        alert(errTxt || 'خطا در شروع تایمر');
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
        const stoppedProject = this.timer.project;
        const stoppedColor = this.timer.color;
        this.applyTimerState({ is_running: false, project: stoppedProject, color: stoppedColor }, 0);
        if (this.selectedProject && this.selectedProject !== stoppedProject) {
          await this.selectProject(this.selectedProject, this.selectedColor, true);
        }
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
    const selectedOpt = document.querySelector('#manualProjectSelect option:checked');
    const color = (selectedOpt && selectedOpt.dataset.color) ? selectedOpt.dataset.color : this.getProjectColor(project);

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

  applyTimerState(timer, elapsedSeconds, syncSelection = false) {
    const prevTimerProject = this.timer ? this.timer.project : '';
    const incomingTimer = timer || { is_running: false };

    if (syncSelection || (incomingTimer.project && incomingTimer.project !== prevTimerProject)) {
      if (incomingTimer.project) {
        this.selectedProject = incomingTimer.project;
        this.selectedColor = incomingTimer.color || this.getProjectColor(incomingTimer.project);
        this.syncSelectedChip();
        if (this.projectsSummary) {
          this.renderProjectsTotals();
          this.renderProjectTimeline();
        }
      }
    }

    this.timer = {
      ...incomingTimer,
      elapsed_seconds: elapsedSeconds || 0,
    };

    const headerClock = document.getElementById('headerTimerClock');
    const headerProject = document.getElementById('headerTimerProject');
    const headerDesc = document.getElementById('headerTimerDesc');
    const globalDot = document.getElementById('globalTimerDot');

    const heroDigits = document.getElementById('heroTimerDigits');
    const heroStatus = document.getElementById('heroTimerStatus');

    const pauseBtn = document.getElementById('timerPauseBtn');
    const resumeBtn = document.getElementById('timerResumeBtn');
    const stopBtn = document.getElementById('timerStopBtn');

    if (this.timer.is_running) {
      headerProject.textContent = this.timer.project;
      headerClock.textContent = window.toPersianDigits(window.formatSeconds(this.timer.elapsed_seconds));

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

      pauseBtn.style.display = 'none';
      resumeBtn.style.display = 'none';
      stopBtn.style.display = 'none';
    }

    this.updateTimerActionControls();

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

    await this.loadProjectsSummary();
  }

  escapeHtml(str) {
    return String(str || '')
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#39;');
  }

  formatDurationLabel(sec) {
    const total = Math.max(0, Math.floor(sec || 0));
    const hours = Math.floor(total / 3600);
    const mins = Math.floor((total % 3600) / 60);
    const secs = total % 60;
    if (hours > 0) {
      return `${window.toPersianDigits(String(hours))} ساعت و ${window.toPersianDigits(String(mins))} دقیقه`;
    }
    if (mins > 0) {
      return `${window.toPersianDigits(String(mins))} دقیقه`;
    }
    if (secs > 0) {
      return `${window.toPersianDigits(String(secs))} ثانیه`;
    }
    return '۰ دقیقه';
  }

  setTimelineMode(mode) {
    this.timelineMode = mode === 'all' ? 'all' : 'selected';
    const selBtn = document.getElementById('timelineFilterSelectedBtn');
    const allBtn = document.getElementById('timelineFilterAllBtn');
    if (selBtn) selBtn.classList.toggle('active', this.timelineMode === 'selected');
    if (allBtn) allBtn.classList.toggle('active', this.timelineMode === 'all');
    this.renderProjectTimeline();
  }

  async loadProjectsSummary() {
    try {
      const res = await fetch('/api/worktime/projects');
      if (!res.ok) return;
      this.projectsSummary = await res.json();
      this.renderProjectsTotals();
      this.renderProjectTimeline();
    } catch (err) {
      console.error('Failed to load projects summary:', err);
    }
  }

  renderProjectsTotals() {
    const totalElem = document.getElementById('worktimeAllTimeTotal');
    const container = document.getElementById('allProjectsTotalsList');
    if (!container) return;

    const summary = this.projectsSummary;
    if (!summary) {
      if (totalElem) totalElem.textContent = '۰ ساعت و ۰ دقیقه';
      container.innerHTML = '<div class="empty-text">در حال بارگذاری...</div>';
      return;
    }

    const totalSec = summary.total_seconds || 0;
    if (totalElem) {
      const hours = Math.floor(totalSec / 3600);
      const mins = Math.floor((totalSec % 3600) / 60);
      totalElem.textContent = `${window.toPersianDigits(String(hours))} ساعت و ${window.toPersianDigits(String(mins))} دقیقه`;
    }

    const projects = summary.projects || [];
    if (projects.length === 0) {
      container.innerHTML = '<div class="empty-text">پروژه‌ای تعریف نشده است</div>';
      return;
    }

    container.innerHTML = '';
    projects.forEach(p => {
      const pct = totalSec > 0 ? Math.round((p.total_seconds / totalSec) * 100) : 0;
      const isActive = p.project === this.selectedProject;
      const color = p.color || '#3b82f6';

      const btn = document.createElement('button');
      btn.type = 'button';
      btn.className = `project-total-item ${isActive ? 'active' : ''}`;
      btn.innerHTML = `
        <div class="project-total-top">
          <div class="project-total-name-wrap">
            <span class="chip-dot" style="background-color: ${this.escapeHtml(color)}; flex-shrink: 0;"></span>
            <span class="project-total-name">${this.escapeHtml(p.project)}</span>
          </div>
          <span class="project-total-duration">${this.formatDurationLabel(p.total_seconds)}</span>
        </div>
        <div class="project-total-meta">
          <span>${window.toPersianDigits(String(p.session_count || 0))} سشن • ${window.toPersianDigits(String(p.active_days || 0))} روز</span>
          <span>${window.toPersianDigits(String(pct))}٪</span>
        </div>
        <div class="project-total-bar-bg">
          <div class="project-total-bar-fill" style="width: ${pct}%; background-color: ${this.escapeHtml(color)};"></div>
        </div>
      `;
      btn.addEventListener('click', () => {
        this.timelineMode = 'selected';
        const selBtn = document.getElementById('timelineFilterSelectedBtn');
        const allBtn = document.getElementById('timelineFilterAllBtn');
        if (selBtn) selBtn.classList.add('active');
        if (allBtn) allBtn.classList.remove('active');
        this.selectProject(p.project, p.color, true);
      });
      container.appendChild(btn);
    });
  }

  renderProjectTimeline() {
    const titleElem = document.getElementById('projectTimelineTitle');
    const container = document.getElementById('projectTimelineContainer');
    if (!container) return;

    const selected = this.selectedProject || 'پروژه عمومی';
    if (titleElem) {
      titleElem.textContent = this.timelineMode === 'all'
        ? 'تایم‌لاین همه پروژه‌ها:'
        : `تایم‌لاین: ${selected}`;
    }

    if (!this.projectsSummary || !this.projectsSummary.projects) {
      container.innerHTML = '<div class="empty-text">سابقه‌ای ثبت نشده است</div>';
      return;
    }

    let days = [];
    let dotColor = this.selectedColor || this.getProjectColor(selected);

    if (this.timelineMode === 'selected') {
      const proj = this.projectsSummary.projects.find(p => p.project === selected);
      days = (proj && proj.timeline) ? proj.timeline : [];
      if (proj && proj.color) dotColor = proj.color;
    } else {
      dotColor = '#3b82f6';
      const dayMap = new Map();
      this.projectsSummary.projects.forEach(p => {
        (p.timeline || []).forEach(d => {
          let existing = dayMap.get(d.jalali_date);
          if (!existing) {
            existing = {
              jalali_date: d.jalali_date,
              duration_seconds: 0,
              session_count: 0,
              sessions: [],
            };
            dayMap.set(d.jalali_date, existing);
          }
          existing.duration_seconds += d.duration_seconds || 0;
          existing.session_count += d.session_count || 0;
          existing.sessions.push(...(d.sessions || []));
        });
      });
      days = Array.from(dayMap.values()).sort((a, b) => b.jalali_date.localeCompare(a.jalali_date));
      days.forEach(d => {
        d.sessions.sort((a, b) => new Date(b.start_time) - new Date(a.start_time));
      });
    }

    if (days.length === 0) {
      const emptyMsg = this.timelineMode === 'all'
        ? 'هنوز سشنی ثبت نشده است'
        : `هنوز سشنی برای «${this.escapeHtml(selected)}» ثبت نشده است`;
      container.innerHTML = `<div class="empty-text">${emptyMsg}</div>`;
      return;
    }

    const todayIso = window.calendarModule && window.calendarModule.todayDate ? window.calendarModule.todayDate.iso : '';
    container.innerHTML = '';

    days.forEach(day => {
      const dateStr = window.toPersianDigits((day.jalali_date || '').replace(/-/g, '/'));
      const todayBadge = (day.jalali_date && day.jalali_date === todayIso) ? ' (امروز)' : '';

      const groupEl = document.createElement('div');
      groupEl.className = 'timeline-day-group';

      const dotEl = document.createElement('span');
      dotEl.className = 'timeline-day-dot';
      dotEl.style.backgroundColor = dotColor;
      groupEl.appendChild(dotEl);

      const headerEl = document.createElement('div');
      headerEl.className = 'timeline-day-header';
      headerEl.innerHTML = `
        <span>${dateStr}${todayBadge}</span>
        <span class="timeline-day-total">${this.formatDurationLabel(day.duration_seconds)}</span>
      `;
      groupEl.appendChild(headerEl);

      const listEl = document.createElement('div');
      listEl.className = 'timeline-session-list';

      (day.sessions || []).forEach(s => {
        const start = new Date(s.start_time).toLocaleTimeString('fa-IR', { hour: '2-digit', minute: '2-digit' });
        const end = new Date(s.end_time).toLocaleTimeString('fa-IR', { hour: '2-digit', minute: '2-digit' });
        const sessColor = s.color || this.getProjectColor(s.project);
        const descText = s.description ? this.escapeHtml(s.description) : 'بدون توضیح';

        const itemEl = document.createElement('div');
        itemEl.className = 'timeline-session-item';
        itemEl.innerHTML = `
          <div class="timeline-session-main">
            <div class="timeline-session-top">
              <span class="timeline-session-project" style="color:${this.escapeHtml(sessColor)};">${this.escapeHtml(s.project)}</span>
              <span class="timeline-session-range">${start} - ${end}</span>
            </div>
            <div class="timeline-session-desc">${descText}</div>
          </div>
          <div class="timeline-session-actions">
            <span class="timeline-session-dur">${window.toPersianDigits(window.formatSeconds(s.duration_seconds))}</span>
            <button type="button" class="inline-add-btn timeline-del-btn" style="color:#f43f5e;" title="حذف">×</button>
          </div>
        `;
        const delBtn = itemEl.querySelector('.timeline-del-btn');
        if (delBtn) {
          delBtn.addEventListener('click', () => this.deleteSession(s.id));
        }
        listEl.appendChild(itemEl);
      });

      groupEl.appendChild(listEl);
      container.appendChild(groupEl);
    });
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
