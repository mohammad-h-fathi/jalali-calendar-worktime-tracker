/**
 * Calendar module: Month view, navigation, day detail inspection
 */
class CalendarModule {
  constructor() {
    this.currentYear = 1405;
    this.currentMonth = 8;
    this.selectedDate = null;
    this.todayDate = null;
    this.monthDaysData = [];
  }

  async init(todayInfo) {
    this.todayDate = {
      year: todayInfo.year,
      month: todayInfo.month,
      day: todayInfo.day,
      iso: todayInfo.iso,
    };
    this.currentYear = todayInfo.year;
    this.currentMonth = todayInfo.month;
    this.selectedDate = { ...this.todayDate };

    this.bindEvents();
    await this.loadMonth(this.currentYear, this.currentMonth);
    await this.selectDay(this.selectedDate.day);
  }

  bindEvents() {
    document.getElementById('calPrevMonthBtn').addEventListener('click', () => {
      this.prevMonth();
    });

    document.getElementById('calNextMonthBtn').addEventListener('click', () => {
      this.nextMonth();
    });

    document.getElementById('calTodayBtn').addEventListener('click', () => {
      this.currentYear = this.todayDate.year;
      this.currentMonth = this.todayDate.month;
      this.loadMonth(this.currentYear, this.currentMonth).then(() => {
        this.selectDay(this.todayDate.day);
      });
    });

    document.getElementById('quickAddTaskBtn').addEventListener('click', () => {
      window.app.switchTab('todo');
      const iso = `${this.currentYear}-${String(this.currentMonth).padStart(2, '0')}-${String(this.selectedDate.day).padStart(2, '0')}`;
      document.getElementById('newTaskJalaliDate').value = iso;
      document.getElementById('newTaskTitle').focus();
    });
  }

  prevMonth() {
    this.currentMonth--;
    if (this.currentMonth < 1) {
      this.currentMonth = 12;
      this.currentYear--;
    }
    this.loadMonth(this.currentYear, this.currentMonth);
  }

  nextMonth() {
    this.currentMonth++;
    if (this.currentMonth > 12) {
      this.currentMonth = 1;
      this.currentYear++;
    }
    this.loadMonth(this.currentYear, this.currentMonth);
  }

  async loadMonth(year, month) {
    try {
      const res = await fetch(`/api/calendar/month?year=${year}&month=${month}`);
      const data = await res.json();
      this.monthDaysData = data.days || [];

      document.getElementById('calendarMonthTitle').textContent = `${data.month_name} ${window.toPersianDigits(String(year))}`;

      this.renderDaysGrid(data);
    } catch (err) {
      console.error('Failed to load month:', err);
    }
  }

  renderDaysGrid(monthData) {
    const grid = document.getElementById('calendarDaysGrid');
    grid.innerHTML = '';

    // Empty cells before first weekday
    for (let i = 0; i < monthData.first_weekday; i++) {
      const blank = document.createElement('div');
      blank.className = 'day-cell other-month';
      grid.appendChild(blank);
    }

    monthData.days.forEach((dayItem) => {
      const cell = document.createElement('div');
      cell.className = 'day-cell';
      cell.dataset.day = dayItem.day;

      if (dayItem.is_today) {
        cell.classList.add('is-today');
      }
      if (dayItem.is_holiday) {
        cell.classList.add('is-holiday');
      }
      if (this.selectedDate && this.selectedDate.year === this.currentYear && this.selectedDate.month === this.currentMonth && this.selectedDate.day === dayItem.day) {
        cell.classList.add('is-selected');
      }

      let dotsHtml = '<div class="day-dots-row">';
      if (dayItem.is_holiday) dotsHtml += '<span class="dot-holiday" title="تعطیل"></span>';
      if (dayItem.task_count > 0) dotsHtml += '<span class="dot-task" title="وظیفه"></span>';
      if (dayItem.work_seconds > 0) dotsHtml += '<span class="dot-work" title="کارکرد"></span>';
      dotsHtml += '</div>';

      cell.innerHTML = `
        <span class="day-number-text">${window.toPersianDigits(String(dayItem.day))}</span>
        ${dotsHtml}
      `;

      cell.addEventListener('click', () => {
        this.selectDay(dayItem.day);
      });

      grid.appendChild(cell);
    });
  }

  async selectDay(day) {
    this.selectedDate = {
      year: this.currentYear,
      month: this.currentMonth,
      day: day,
    };

    document.querySelectorAll('.day-cell').forEach(c => c.classList.remove('is-selected'));
    const targetCell = document.querySelector(`.day-cell[data-day="${day}"]`);
    if (targetCell) {
      targetCell.classList.add('is-selected');
    }

    const iso = `${this.currentYear}-${String(this.currentMonth).padStart(2, '0')}-${String(day).padStart(2, '0')}`;
    const isToday = (this.todayDate.year === this.currentYear && this.todayDate.month === this.currentMonth && this.todayDate.day === day);

    const monthName = document.getElementById('calendarMonthTitle').textContent.split(' ')[0];
    document.getElementById('selectedDateTitle').textContent = `${window.toPersianDigits(String(day))} ${monthName} ${window.toPersianDigits(String(this.currentYear))}`;
    document.getElementById('selectedDateSubtitle').textContent = isToday ? 'امروز' : 'انتخاب شده';

    await this.loadDayDetails(iso, day);
  }

  async loadDayDetails(isoDate, day) {
    const currentDayData = this.monthDaysData.find(d => d.day === day);

    // 1. Render Official Events (from iran.ir)
    const eventsList = document.getElementById('selectedDayEventsList');
    if (currentDayData && currentDayData.events && currentDayData.events.length > 0) {
      eventsList.innerHTML = currentDayData.events.map(ev => `
        <li class="mini-event-item ${ev.is_holiday ? 'is-holiday' : ''}">
          <div style="font-weight:600;">${ev.title}</div>
        </li>
      `).join('');
    } else {
      eventsList.innerHTML = '<li class="empty-text">رویدادی برای این روز نیست</li>';
    }

    // 2. Render Tasks
    try {
      const tasksRes = await fetch(`/api/tasks?date=${isoDate}`);
      const tasks = await tasksRes.json();
      const tasksList = document.getElementById('selectedDayTasksList');

      if (tasks && tasks.length > 0) {
        tasksList.innerHTML = tasks.map(t => `
          <li class="mini-event-item" style="display:flex; align-items:center; gap:8px;">
            <input type="checkbox" style="width:16px; height:16px; accent-color:#3b82f6;" ${t.completed ? 'checked' : ''} onchange="window.todoModule.toggleTask('${t.id}')">
            <span style="font-size:0.8rem; ${t.completed ? 'text-decoration:line-through; opacity:0.6;' : ''}">${t.title}</span>
          </li>
        `).join('');
      } else {
        tasksList.innerHTML = '<li class="empty-text">کاری برای این روز نیست</li>';
      }
    } catch (e) {}

    // 3. Render Worktime
    try {
      const workRes = await fetch(`/api/worktime/daily?date=${isoDate}`);
      const workData = await workRes.json();

      const hours = Math.floor(workData.total_seconds / 3600);
      const mins = Math.floor((workData.total_seconds % 3600) / 60);
      document.getElementById('selectedDayWorkTotal').textContent = `${window.toPersianDigits(String(hours))}س و ${window.toPersianDigits(String(mins))}د`;
    } catch (e) {}
  }
}

window.calendarModule = new CalendarModule();
