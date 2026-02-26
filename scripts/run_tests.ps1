# Запуск всех тестов перед продом
# Использование: .\scripts\run_tests.ps1   или   pwsh -File scripts\run_tests.ps1

$ErrorActionPreference = "Stop"
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$root = Split-Path -Parent $scriptDir
if (-not $root) { $root = Get-Location | Select-Object -ExpandProperty Path }
Set-Location $root

Write-Host "=== GND: запуск тестов ===" -ForegroundColor Cyan
Write-Host "Каталог: $root" -ForegroundColor Gray

$goCmd = Get-Command go -ErrorAction SilentlyContinue
if (-not $goCmd) {
    Write-Host "`n[ERROR] go не найден в PATH." -ForegroundColor Red
    Write-Host "Добавьте каталог bin установки Go в PATH (например C:\Program Files\Go\bin)." -ForegroundColor Yellow
    Write-Host "См. docs/deployment-server.md — раздел 'Go не найден в PATH'." -ForegroundColor Gray
    exit 1
}

# Короткие тесты (без длительных операций), затем при необходимости полные
Write-Host "`n[1/2] go test ./... -count=1 -short" -ForegroundColor Yellow
go test ./... -count=1 -short
if ($LASTEXITCODE -ne 0) {
    Write-Host "FAIL: короткие тесты завершились с ошибкой" -ForegroundColor Red
    exit $LASTEXITCODE
}

Write-Host "`n[2/2] go test ./... -count=1 (полный прогон, включая интеграционные)" -ForegroundColor Yellow
go test ./... -count=1
if ($LASTEXITCODE -ne 0) {
    Write-Host "FAIL: тесты завершились с ошибкой" -ForegroundColor Red
    exit $LASTEXITCODE
}

Write-Host "`n=== Все тесты пройдены ===" -ForegroundColor Green
exit 0
