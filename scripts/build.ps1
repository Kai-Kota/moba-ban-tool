# 配布用の単一 exe をビルドする。
#
#   powershell -ExecutionPolicy Bypass -File scripts\build.ps1
#
# 出力:
#   release\moba-ban-tool\moba-ban-tool.exe   フロントエンドを埋め込んだサーバー
#   release\moba-ban-tool\data\               キャラ・ドラフト順・画像(差し替え用)
#   release\moba-ban-tool\README.md
#   release\moba-ban-tool.zip                 配布用(state.json を含まないまっさらな状態)
#
# release\moba-ban-tool をそのまま使っている場合に備え、再ビルドしても
# state.json(ドラフトの状態)と data(差し替えたキャラや画像)は消さずに残す。
# 置き換えるのは exe と README だけ。

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
$out = Join-Path $root "release\moba-ban-tool"
$zip = Join-Path $root "release\moba-ban-tool.zip"
$stageRoot = Join-Path $root "release\_stage"
$stage = Join-Path $stageRoot "moba-ban-tool"
$exeName = "moba-ban-tool.exe"

Push-Location $root
try {
    $running = Get-Process moba-ban-tool -ErrorAction SilentlyContinue |
        Where-Object { $_.Path -eq (Join-Path $out $exeName) }
    if ($running) { throw "$out の exe が起動中です。コンソールウィンドウを閉じてから実行してください" }

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
    # zip 用の作業フォルダはまっさらな状態から作る
    if (Test-Path $stageRoot) { Remove-Item -Recurse -Force $stageRoot }
    New-Item -ItemType Directory -Force $stage | Out-Null
    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    $env:CGO_ENABLED = "0"
    go build -tags embed -trimpath -ldflags "-s -w" -o (Join-Path $stage $exeName) ./cmd/draftserver
    if ($LASTEXITCODE -ne 0) { throw "go build に失敗しました" }
    Copy-Item -Recurse "data" (Join-Path $stage "data")
    Copy-Item "README.md" $stage

    Write-Host "==> zip を作成" -ForegroundColor Cyan
    if (Test-Path $zip) { Remove-Item -Force $zip }
    Compress-Archive -Path $stage -DestinationPath $zip

    Write-Host "==> $out を更新(state.json と data は残す)" -ForegroundColor Cyan
    New-Item -ItemType Directory -Force $out | Out-Null
    Copy-Item -Force (Join-Path $stage $exeName) $out
    Copy-Item -Force (Join-Path $stage "README.md") $out
    if (-not (Test-Path (Join-Path $out "data"))) {
        Copy-Item -Recurse (Join-Path $stage "data") (Join-Path $out "data")
    }
    Remove-Item -Recurse -Force $stageRoot

    $size = [math]::Round((Get-Item (Join-Path $out $exeName)).Length / 1MB, 1)
    Write-Host ""
    Write-Host "完了: $out (exe $size MB)" -ForegroundColor Green
    Write-Host "      $zip" -ForegroundColor Green
}
finally {
    Remove-Item Env:GOOS, Env:GOARCH, Env:CGO_ENABLED -ErrorAction SilentlyContinue
    Pop-Location
}
