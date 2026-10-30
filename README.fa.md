# گاه‌شمار جلالی و زمان‌سنج کارکرد (Jalali Calendar & Worktime Tracker)

<div align="center">

[🇬🇧 Read in English](README.md) • [🇮🇷 مطالعه به زبان فارسی](README.fa.md)

[![Release](https://img.shields.io/github/v/release/mohammad-h-fathi/jalali-calendar-worktime-tracker?style=flat-square&color=3b82f6)](https://github.com/mohammad-h-fathi/jalali-calendar-worktime-tracker/releases)
[![Build & Release](https://github.com/mohammad-h-fathi/jalali-calendar-worktime-tracker/actions/workflows/release.yml/badge.svg)](https://github.com/mohammad-h-fathi/jalali-calendar-worktime-tracker/actions)
[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20Windows%20%7C%20macOS-blueviolet?style=flat-square)](https://github.com/mohammad-h-fathi/jalali-calendar-worktime-tracker/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square)](LICENSE)

</div>

یک اپلیکیشن دسکتاپ بومی، سبک و مدرن برای **لینوکس (اوبونتو ۲۴.۰۴ و ۲۶.۰۴ با GNOME و Wayland)**، **ویندوز (Windows 10/11)** و **مک (macOS)** که با زبان **Go** توسعه یافته است. این برنامه تاریخ خورشیدی و زمان‌سنج زنده را در نوار بالای اوبونتو نمایش می‌دهد، مجهز به یک **تایمر شناور بدون قاب (Borderless Always-on-Top Pill)** است و مدیریت کارهای روزانه (To-Do) و ثبت کارکرد زمانی را با تایپوگرافی چشم‌نواز فونت فارسی **لحظه (Lahzeh)** فراهم می‌کند.

---

## ✨ ویژگی‌های برجسته

### ۱. نشانگر زنده در نوار بالای اوبونتو (GNOME Status Area)
* نمایش دائمی تاریخ خورشیدی جاری در نوار وضعیت بالای اوبونتو (مثلاً `۸ آبان ۱۴۰۵`).
* **تایمر دینامیک:** به محض شروع کار، زمان‌سنج به صورت زنده و ثانیه‌ای در کنار تاریخ به نمایش درمی‌آید:  
  `۸ آبان | ⏱ ۰۰:۲۴:۱۰`
* **منوی کاربردی کلیک راست (DBusMenu):** دسترسی سریع به باز کردن گاه‌شمار، شروع/توقف زمان‌سنج، باز کردن تایمر شناور و خروج از برنامه.

### ۲. تایمر شناور مینیاتوری و بدون قاب (Borderless Floating Action Pill)
* **کاملاً بدون کادر:** بدون نوار عنوان (Titlebar)، بدون دکمه‌های بستن و مینیمایز پنجره، با پس‌زمینه شیشه‌ای تیره (Glassmorphic) و گوشه‌های کاملاً گرد.
* **همیشه رو (Always-on-Top):** روی تمام برنامه‌ها (مرورگرها، IDEها، ترمینال و بازی‌ها) معلق می‌ماند.
* **جابه‌جایی روان (Draggable):** با کلیک و درگ ماوس در هر نقطه از صفحه مانیتور جابه‌جا می‌شود.
* **حذف خودکار با زدن توقف (Auto-Removed on Stop):** به محض کلیک روی دکمه «پایان ⏹»، زمان کارکرد به عنوان سشن ذخیره شده و پنجره شناور **بلافاصله و خودکار از روی صفحه بسته و محو می‌شود**.
* **مستقل و بدون وابستگی:** روی ویندوز از طریق موتور داخلی PowerShell/WPF اجرا شده و به هیچ پیش‌نیاز یا نصب پایتون احتیاج ندارد.

### ۳. تقویم خورشیدی هوشمند و مناسبت‌ها
* پیاده‌سازی دقیق الگوریتم نجومی تقویم جلالی با محاسبه دقیق سال‌های کبیسه.
* **اتصال به سامانه رویدادهای ملی:** دریافت رویدادها و تعطیلات رسمی هر ماه به طور خودکار از API رسمی (`iran.ir/Calendar/GetEvents`) به همراه کش محلی و حالت آفلاین.
* نمایش نشانگر کارهای روز و مجموع ساعات کارکرد ثبت‌شده در تقویم ماهانه.

### ۴. سیستم ثبت زمان کار (Worktime Tracker)
* مدیریت پروژه‌های کاری مختلف با برچسب‌های رنگی دلخواه.
* زمان‌سنج زنده با قابلیت مکث (Pause) و ادامه (Resume).
* امکان ثبت دستی سشن‌های کارکرد گذشته (مشخص کردن ساعت شروع، پایان و توضیحات).
* نمایش خلاصه ساعات کارکرد روزانه و تفکیک درصدی پروژه‌ها.

### ۵. مدیریت کارها و وظایف (To-Do List)
* ثبت وظایف متصل به روزهای مشخص تقویم جلالی.
* تیک زدن و حذف کارها با ذخیره‌سازی فوری.
* شمارنده کارهای باز و ناتمام در نوار ناوبری.

---

## 📥 دانلود و نصب (Installation)

### 🐧 برای اوبونتو و دبیان (Debian / Ubuntu)

#### روش ۱: پکیج استاندارد `.deb` (ساده‌ترین روش)
از صفحه [Releases](https://github.com/mohammad-h-fathi/jalali-calendar-worktime-tracker/releases) آخرین فایل `.deb` را دانلود کنید:
```bash
sudo apt install ./persian-calendar_*.deb
```
> با این دستور، فایل اجرایی، آیکون اختصاصی، فونت‌های لحظه و شورتکات دسکتاپ در مسیرهای سیستمی ثبت می‌شوند.

#### روش ۲: نصب از طریق اسکریپت (بدون نیاز به دسترسی روت)
```bash
git clone https://github.com/mohammad-h-fathi/jalali-calendar-worktime-tracker.git
cd jalali-calendar-worktime-tracker
./scripts/install-desktop.sh
```

---

### 🪟 برای ویندوز (Windows 10 / 11)

1. از صفحه [Releases](https://github.com/mohammad-h-fathi/jalali-calendar-worktime-tracker/releases) فایل `persian-calendar_*_windows_amd64.zip` را دانلود و استخراج (Unzip) کنید.
2. روی فایل **`install-windows.bat`** راست‌کلیک کرده و Run کنید.
3. میانبر برنامه روی دسکتاپ و منوی استارت ایجاد شده و برنامه با بالا آمدن ویندوز در System Tray فعال خواهد بود.

---

### 🍎 برای مک او اس (macOS)

فایل `persian-calendar_*_darwin_arm64.tar.gz` (برای تراشه‌های M1 تا M4) یا `amd64` (برای پردازنده‌های Intel) را دانلود کرده و استخراج کنید:
```bash
tar -xzf persian-calendar_*_darwin_*.tar.gz
./persian-calendar
```

---

## 🛠 کامپایل از روی سورس (Build from Source)

پیش‌نیازها: نصب بودن [Go 1.22+](https://go.dev/)

```bash
# کلون کردن ریپازیتوری
git clone https://github.com/mohammad-h-fathi/jalali-calendar-worktime-tracker.git
cd jalali-calendar-worktime-tracker

# بیلد نسخه لینوکس:
go build -o bin/persian-calendar ./cmd/persian-calendar

# کراس کامپایل برای ویندوز (بدون نیاز به CGO):
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/persian-calendar.exe ./cmd/persian-calendar

# کراس کامپایل برای مک (Apple Silicon):
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o bin/persian-calendar-darwin-arm64 ./cmd/persian-calendar
```

---

## 🚀 اجرای برنامه در خط فرمان

```bash
# اجرای کامل (باز شدن داشبورد در پنجره مستقل):
persian-calendar

# اجرای پس‌زمینه (تنها حضور در نوار بالای اوبونتو / Tray بدون باز شدن پنجره اولیه):
persian-calendar --no-window

# مشخص کردن پورت دلخواه (پیش‌فرض: 41890):
persian-calendar --port=45000
```

---

## 📡 REST API & SSE Endpoints

برنامه یک سرور سبک و سریع محلی اجرا می‌کند:

| آدرس (Endpoint) | متد | شرح عملکرد |
| :--- | :---: | :--- |
| `/api/today` | `GET` | دریافت تاریخ امروز خورشیدی، رویدادها، وضعیت تایمر و تسک‌ها |
| `/api/calendar/month?year=1405&month=8` | `GET` | ماتریس کامل روزهای ماه به همراه رویدادها و ساعت کارکرد |
| `/api/tasks` | `GET / POST` | دریافت لیست وظایف یا ایجاد کار جدید |
| `/api/tasks/{id}/toggle` | `POST` | تغییر وضعیت تکمیل وظیفه |
| `/api/timer` | `GET` | وضعیت زنده تایمر و ثانیه‌های سپری‌شده |
| `/api/timer/start` | `POST` | شروع زمان‌سنج برای یک پروژه |
| `/api/timer/pause` | `POST` | توقف موقت زمان‌سنج |
| `/api/timer/resume` | `POST` | ادامه مجدد زمان‌سنج |
| `/api/timer/stop` | `POST` | پایان و ثبت قطعی سشن کارکرد |
| `/api/timer/floating/open` | `POST` | باز کردن ویجت کپسول شناور بدون قاب |
| `/api/events/stream` | `GET` | استریم بلادرنگ Server-Sent Events (SSE) برای همگام‌سازی لحظه‌ای |

---

## 📂 ساختار ماژولار پروژه

```text
jalali-calendar-worktime-tracker/
├── .github/workflows/
│   └── release.yml          # اکشن خودکار بیلد و ساخت پکیج‌های چندسکویی
├── cmd/
│   └── persian-calendar/
│       └── main.go          # نقطه ورود اصلی برنامه
├── internal/
│   ├── calendar/            # موتور محاسبات نجومی جلالی، سال‌های کبیسه و مناسبت‌ها
│   ├── models/              # تعریف ساختارهای داده Task, WorkSession, Project
│   ├── server/              # سرور HTTP تعبیه‌شده و استریم رویدادهای SSE
│   ├── storage/             # سیستم ذخیره‌سازی اتمیک و امن
│   ├── todo/                # سرویس مدیریت وظایف
│   ├── tracker/             # سرویس زمان‌سنج و تحلیل کارکرد
│   ├── tray/                # هندلر D-Bus StatusNotifierItem نوار بالای لینوکس و Tray ویندوز
│   └── ui/                  # لانچر پنجره و مدیریت تایمر شناور
├── scripts/
│   ├── floating_timer.py    # ویجت کپسول شناور بدون قاب برای لینوکس (GTK 3)
│   ├── floating_timer.ps1   # ویجت کپسول شناور بدون قاب برای ویندوز (WPF)
│   ├── install-desktop.sh   # نصاب دسکتاپ لینوکس
│   ├── install-windows.bat  # نصاب ۱ کلیکی ویندوز
│   └── package-deb.sh       # اسکریپت ساخت پکیج Debian .deb
├── fonts/                   # خانواده فونت فارسی «لحظه» در ۹ وزن
└── web/                     # دارایی‌های تعبیه‌شده وب (HTML5, Vanilla CSS, JS)
```

---

## 📄 لایسنس (License)

این پروژه تحت مجوز **MIT** منتشر شده است و استفاده، توسعه و بازتوزیع آن کاملاً آزاد و رایگان است.
