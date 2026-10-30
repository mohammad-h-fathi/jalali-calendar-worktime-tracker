@echo off
echo Building Persian Calendar for Windows (x64)...
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-H windowsgui" -o bin\persian-calendar.exe .\cmd\persian-calendar
if %ERRORLEVEL% EQU 0 (
    echo.
    echo Build Successful: bin\persian-calendar.exe
) else (
    echo.
    echo Build Failed!
)
pause
