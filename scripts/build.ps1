# 配布用の単一 exe をビルドする。
#
#   powershell -ExecutionPolicy Bypass -File scripts\build.ps1
#
# 出力:
#   release\moba-ban-tool\moba-ban-tool.exe   フロントエンドを埋め込んだサーバー
#   release\moba-ban-tool\data\               キャラ・ドラフト順・画像(差し替え用)
#   release\moba-ban-tool\README.md
#   release\moba-ban-tool.zip                 上記をまとめたもの

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
$out = Join-Path $root "release\moba-ban-tool"
$zip = Join-Path $root "release\moba-ban-tool.zip"

Push-Location $root
try {
    Write-Host "==> フロントエンドをビルド" -ForegroundColor Cyan
    if (-not (Test-Path "web\node_modules")) {
        npm --prefix web ci
        if ($LASTEXITCODE -ne 0) { throw "npm ci に失敗しました" }
    }
    npm --prefix web run build
    if ($LASTEXITCODE -ne 0) { throw "フロントエンドのビルドに失敗しました" }

    Write-Host "==> テスト" -ForegroundColor Cyan
    go test ./internal/...
    if ($LASTEXITCODE -ne 0) { throw "go test に失敗しました" }

    Write-Host "==> exe をビルド" -ForegroundColor Cyan
    if (Test-Path $out) { Remove-Item -Recurse -Force $out }
    New-Item -ItemType Directory -Force $out | Out-Null
    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    $env:CGO_ENABLED = "0"
    go build -tags embed -trimpath -ldflags "-s -w" -o (Join-Path $out "moba-ban-tool.exe") ./cmd/draftserver
    if ($LASTEXITCODE -ne 0) { throw "go build に失敗しました" }

    Write-Host "==> data と README をコピー" -ForegroundColor Cyan
    Copy-Item -Recurse "data" (Join-Path $out "data")
    Copy-Item "README.md" $out

    if (Test-Path $zip) { Remove-Item -Force $zip }
    Compress-Archive -Path $out -DestinationPath $zip

    $size = [math]::Round((Get-Item (Join-Path $out "moba-ban-tool.exe")).Length / 1MB, 1)
    Write-Host ""
    Write-Host "完了: $out (exe $size MB)" -ForegroundColor Green
    Write-Host "      $zip" -ForegroundColor Green
}
finally {
    Remove-Item Env:GOOS, Env:GOARCH, Env:CGO_ENABLED -ErrorAction SilentlyContinue
    Pop-Location
}
