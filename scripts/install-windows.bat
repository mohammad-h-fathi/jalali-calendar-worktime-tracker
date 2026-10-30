@echo off
chcp 65001 >nul
echo ========================================================
echo        نصب‌کننده تقویم جلالی و ثبت زمان‌کار
echo ========================================================
echo.

set "TARGET_DIR=%LOCALAPPDATA%\PersianCalendar"
mkdir "%TARGET_DIR%" 2>nul
mkdir "%TARGET_DIR%\scripts" 2>nul

echo [1/3] در حال کپی فایل‌های اجرایی...
copy /Y "persian-calendar.exe" "%TARGET_DIR%\" >nul
if exist "scripts\floating_timer.ps1" (
    copy /Y "scripts\floating_timer.ps1" "%TARGET_DIR%\scripts\" >nul
)

echo [2/3] در حال ایجاد میانبر روی دسکتاپ و منوی استارت...
powershell -NoProfile -Command ^
  "$ws = New-Object -ComObject WScript.Shell; " ^
  "$sDesktop = $ws.CreateShortcut([System.IO.Path]::Combine([Environment]::GetFolderPath('Desktop'), 'تقویم جلالی.lnk')); " ^
  "$sDesktop.TargetPath = '%TARGET_DIR%\persian-calendar.exe'; " ^
  "$sDesktop.Save(); " ^
  "$startMenuPath = [System.IO.Path]::Combine([Environment]::GetFolderPath('StartMenu'), 'Programs', 'تقویم جلالی.lnk'); " ^
  "$sStart = $ws.CreateShortcut($startMenuPath); " ^
  "$sStart.TargetPath = '%TARGET_DIR%\persian-calendar.exe'; " ^
  "$sStart.Save(); " ^
  "$startupPath = [System.IO.Path]::Combine([Environment]::GetFolderPath('Startup'), 'تقویم جلالی (پس‌زمینه).lnk'); " ^
  "$sStartup = $ws.CreateShortcut($startupPath); " ^
  "$sStartup.TargetPath = '%TARGET_DIR%\persian-calendar.exe'; " ^
  "$sStartup.Arguments = '--no-window'; " ^
  "$sStartup.Save();"

echo [3/3] نصب فونت‌ها...
if exist "fonts" (
    powershell -NoProfile -Command ^
      "$fonts = Get-ChildItem -Path 'fonts\*.ttf'; " ^
      "$shellApp = New-Object -ComObject Shell.Application; " ^
      "$fontFolder = $shellApp.Namespace(0x14); " ^
      "foreach ($f in $fonts) { try { $fontFolder.CopyHere($f.FullName, 16) } catch {} }"
)

echo.
echo ========================================================
echo   نصب با موفقیت انجام شد!
echo   میانبر برنامه روی دسکتاپ و منوی استارت ایجاد شد.
echo ========================================================
pause
