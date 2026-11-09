# 開発用: Go サーバーと Vite の dev server を1つのウィンドウでまとめて起動し、操作パネルを開く。
#
#   dev.cmd をダブルクリック
#   または powershell -ExecutionPolicy Bypass -File scripts\dev.ps1
#
# 終了するにはウィンドウを閉じるか Ctrl+C を押す(Go サーバーも一緒に止まる)。

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
$serverExe = Join-Path $root "bin\draftserver-dev.exe"
$server = $null

Push-Location $root
try {
    if (-not (Test-Path "web\node_modules")) {
        Write-Host "==> npm ci" -ForegroundColor Cyan
        npm --prefix web ci
        if ($LASTEXITCODE -ne 0) { throw "npm ci に失敗しました" }
    }

    # go run だと止めるときに子プロセスの exe が残るので、ビルドしてから直接起動する
    Write-Host "==> Go サーバーをビルド" -ForegroundColor Cyan
    go build -o $serverExe ./cmd/draftserver
    if ($LASTEXITCODE -ne 0) { throw "go build に失敗しました" }

    # 同じコンソールで動かすので、ウィンドウを閉じたときにサーバーも終了する
    $server = Start-Process -FilePath $serverExe -NoNewWindow -PassThru

    Write-Host ""
    Write-Host "  操作パネル  http://localhost:5173/control" -ForegroundColor Green
    Write-Host "  OBS 用      http://localhost:5173/overlay" -ForegroundColor Green
    Write-Host ""

    Push-Location web
    try {
        node node_modules\vite\bin\vite.js --open /control
    }
    finally {
        Pop-Location
    }
}
finally {
    if ($server -and -not $server.HasExited) { Stop-Process -Id $server.Id -Force }
    Pop-Location
}
