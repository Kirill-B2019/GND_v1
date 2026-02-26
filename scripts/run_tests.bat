@echo off
REM Запуск всех тестов перед продом
REM Двойной клик или: scripts\run_tests.bat

cd /d "%~dp0\.."
echo === GND: запуск тестов ===
echo.

where go >nul 2>&1
if errorlevel 1 (
    echo [ERROR] go не найден в PATH.
    echo Добавьте каталог bin установки Go в PATH ^(например C:\Program Files\Go\bin^).
    echo См. docs/deployment-server.md - раздел "Go не найден в PATH".
    exit /b 1
)

echo [1/2] go test -short
go test ./... -count=1 -short
if errorlevel 1 (
    echo FAIL: короткие тесты
    exit /b 1
)

echo.
echo [2/2] go test - полный прогон
go test ./... -count=1
if errorlevel 1 (
    echo FAIL: тесты
    exit /b 1
)

echo.
echo === Все тесты пройдены ===
exit /b 0
