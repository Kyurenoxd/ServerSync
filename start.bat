@echo off
chcp 65001 >nul
title ServerSync
color 0b
cls

:menu
cls
echo.
echo  ============================
echo    ServerSync v1.0
echo    Author: Kyurenoxd
echo    GitHub: github.com/Kyurenoxd
echo    Website: kyureno.dev
echo  ============================
echo.
echo  [1] Инструкция по получению токена
echo  [2] Запустить ServerSync
echo  [3] Выход
echo.
set /p choice="Выберите опцию (1-3): "

if "%choice%"=="1" goto instructions
if "%choice%"=="2" goto start
if "%choice%"=="3" exit

goto menu

:instructions
cls
echo.
echo  ============================
echo    Как получить токен пользователя:
echo  ============================
echo.
echo  1. Откройте Discord в браузере
echo  2. Нажмите Ctrl+Shift+I
echo  3. Перейдите во вкладку Network
echo  4. В поиске введите "api/v9"
echo  5. Найдите в заголовках Authorization
echo  6. Скопируйте ваш токен
echo.
echo  ВНИМАНИЕ: Никогда и никому не передавайте свой токен!
echo  Это может привести к взлому вашего аккаунта!
echo.
echo  [1] Запустить ServerSync
echo  [2] Вернуться в меню
echo.
set /p choice="Выберите опцию (1-2): "

if "%choice%"=="1" goto start
if "%choice%"=="2" goto menu
goto instructions

:start
cls
echo.
echo  Запуск ServerSync...
echo.
timeout /t 2 >nul

:: Компилируем и запускаем
go build -o serversync.exe
if %errorlevel% neq 0 (
    echo  [!] Ошибка при компиляции
    pause
    goto menu
)

serversync.exe
if %errorlevel% neq 0 (
    echo  [!] Программа завершилась с ошибкой
    pause
)
goto menu 