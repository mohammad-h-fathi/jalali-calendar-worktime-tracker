/**
 * To-Do & Task Management module
 */
class TodoModule {
  constructor() {
    this.currentFilter = 'today';
    this.tasks = [];
  }

  async init() {
    this.bindEvents();
    await this.loadTasks();
  }

  bindEvents() {
    document.getElementById('createTaskForm').addEventListener('submit', (e) => {
      e.preventDefault();
      this.createTask();
    });

    document.querySelectorAll('.pill-filter').forEach(tab => {
      tab.addEventListener('click', () => {
        document.querySelectorAll('.pill-filter').forEach(t => t.classList.remove('active'));
        tab.classList.add('active');
        this.currentFilter = tab.dataset.filter;
        this.renderTasks();
      });
    });
  }

  async loadTasks() {
    try {
      const res = await fetch('/api/tasks');
      this.tasks = await res.json() || [];
      this.updateCounters();
      this.renderTasks();
    } catch (err) {
      console.error('Failed to load tasks:', err);
    }
  }

  async createTask() {
    const title = document.getElementById('newTaskTitle').value.trim();
    if (!title) return;

    let date = document.getElementById('newTaskJalaliDate').value.trim();
    if (!date && window.calendarModule && window.calendarModule.todayDate) {
      date = window.calendarModule.todayDate.iso;
    }
    if (date) {
      date = date.replace(/\//g, '-');
      const parts = date.split('-');
      if (parts.length !== 3 || parts[0].length !== 4 || isNaN(parts[0]) || isNaN(parts[1]) || isNaN(parts[2])) {
        alert('فرمت تاریخ نامعتبر است (مثال: ۱۴۰۳-۰۸-۰۴)');
        return;
      }
      date = `${parts[0]}-${String(Number(parts[1])).padStart(2, '0')}-${String(Number(parts[2])).padStart(2, '0')}`;
    }
    const priority = document.getElementById('newTaskPriority').value;

    try {
      const res = await fetch('/api/tasks', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          title,
          jalali_date: date,
          priority,
        }),
      });

      if (res.ok) {
        document.getElementById('newTaskTitle').value = '';
        await this.loadTasks();
        if (window.calendarModule) {
          window.calendarModule.loadMonth(window.calendarModule.currentYear, window.calendarModule.currentMonth);
          if (window.calendarModule.selectedDate) {
            window.calendarModule.selectDay(window.calendarModule.selectedDate.day);
          }
        }
      } else {
        const errTxt = await res.text();
        alert(errTxt || 'خطا در ثبت وظیفه');
      }
    } catch (err) {
      console.error('Failed to create task:', err);
    }
  }

  async toggleTask(id) {
    try {
      const res = await fetch(`/api/tasks/${id}/toggle`, { method: 'POST' });
      if (res.ok) {
        await this.loadTasks();
        if (window.calendarModule) {
          window.calendarModule.loadMonth(window.calendarModule.currentYear, window.calendarModule.currentMonth);
          if (window.calendarModule.selectedDate) {
            window.calendarModule.selectDay(window.calendarModule.selectedDate.day);
          }
        }
      }
    } catch (err) {
      console.error('Failed to toggle task:', err);
    }
  }

  async deleteTask(id) {
    try {
      const res = await fetch(`/api/tasks/${id}`, { method: 'DELETE' });
      if (res.ok) {
        await this.loadTasks();
        if (window.calendarModule) {
          window.calendarModule.loadMonth(window.calendarModule.currentYear, window.calendarModule.currentMonth);
          if (window.calendarModule.selectedDate) {
            window.calendarModule.selectDay(window.calendarModule.selectedDate.day);
          }
        }
      }
    } catch (err) {
      console.error('Failed to delete task:', err);
    }
  }

  updateCounters() {
    const todayIso = window.calendarModule && window.calendarModule.todayDate ? window.calendarModule.todayDate.iso : '';
    const pendingToday = this.tasks.filter(t => !t.completed && t.jalali_date === todayIso).length;
    const allPending = this.tasks.filter(t => !t.completed).length;

    const navBadge = document.getElementById('navTodoCount');
    if (navBadge) navBadge.textContent = window.toPersianDigits(String(pendingToday));

    const summaryText = document.getElementById('tasksCountSummary');
    if (summaryText) summaryText.textContent = `${window.toPersianDigits(String(allPending))} کار`;
  }

  renderTasks() {
    const container = document.getElementById('tasksListContainer');
    if (!container) return;
    const todayIso = window.calendarModule && window.calendarModule.todayDate ? window.calendarModule.todayDate.iso : '';

    let filtered = this.tasks;
    if (this.currentFilter === 'today') {
      filtered = this.tasks.filter(t => t.jalali_date === todayIso);
    } else if (this.currentFilter === 'pending') {
      filtered = this.tasks.filter(t => !t.completed);
    } else if (this.currentFilter === 'completed') {
      filtered = this.tasks.filter(t => t.completed);
    }

    if (filtered.length === 0) {
      container.innerHTML = '<div class="empty-text" style="text-align:center; padding: 20px 0;">کاری در این بخش نیست</div>';
      return;
    }

    container.innerHTML = filtered.map(t => {
      const priorityLabel = {
        urgent: 'فوری 🔥',
        high: 'بالا',
        medium: 'متوسط',
        low: 'کم',
      }[t.priority] || '';

      const priorityColor = {
        urgent: '#fda4af',
        high: '#fdba74',
        medium: '#93c5fd',
        low: '#a7f3d0',
      }[t.priority] || '#cbd5e1';

      return `
        <div class="task-item-row ${t.completed ? 'completed' : ''}">
          <div class="task-item-right">
            <input type="checkbox" class="task-chk" ${t.completed ? 'checked' : ''} onchange="window.todoModule.toggleTask('${t.id}')">
            <div>
              <span class="task-label">${t.title}</span>
              <div style="font-size:0.68rem; color:#94a3b8; display:flex; gap:8px;">
                <span>📅 ${window.toPersianDigits(t.jalali_date)}</span>
                <span style="color:${priorityColor}; font-weight:600;">${priorityLabel}</span>
              </div>
            </div>
          </div>
          <button class="inline-add-btn" style="color:#f43f5e; font-size:0.9rem;" title="حذف" onclick="window.todoModule.deleteTask('${t.id}')">×</button>
        </div>
      `;
    }).join('');
  }
}

window.todoModule = new TodoModule();
