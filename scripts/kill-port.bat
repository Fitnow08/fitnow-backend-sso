@echo off
rem Освобождает порт 44044 перед сборкой (убивает зависший процесс при hot reload на Windows)
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":44044 " ^| findstr "LISTENING"') do (
    taskkill /F /PID %%a >nul 2>&1
)
exit /b 0
