//go:build windows

package main

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"unsafe"

	"github.com/jchv/go-webview2"
)

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	user32           = syscall.NewLazyDLL("user32.dll")
	procGetConsole   = kernel32.NewProc("GetConsoleWindow")
	procAllocConsole = kernel32.NewProc("AllocConsole")
	procMessageBox   = user32.NewProc("MessageBoxW")
	procSetDpiCtx    = user32.NewProc("SetProcessDpiAwarenessContext")
	procGetDpi       = user32.NewProc("GetDpiForSystem")
)

// enableDPIAwareness は高 DPI 環境で文字がぼやけないよう、プロセスを Per-Monitor DPI 対応にする。
// 対応していない場合 Windows がウィンドウをビットマップで引き伸ばすため、表示がにじむ。
// 戻り値は拡大率(100% のとき 1.0)。
func enableDPIAwareness() float64 {
	const perMonitorAwareV2 = ^uintptr(3) // DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2 (-4)
	if procSetDpiCtx.Find() != nil || procGetDpi.Find() != nil {
		return 1 // Windows 10 1703 より前
	}
	procSetDpiCtx.Call(perMonitorAwareV2)
	dpi, _, _ := procGetDpi.Call()
	if dpi == 0 {
		return 1
	}
	return float64(dpi) / 96
}

func init() {
	// WebView2 のウィンドウとメッセージループは同じ OS スレッドで動かす必要がある。
	// main ゴルーチンをメインスレッドに固定しておく。
	runtime.LockOSThread()
}

// runWindow は操作パネルを WebView2 の専用ウィンドウで開き、閉じられるまで戻らない。
// WebView2 ランタイムがない場合などはウィンドウを作らずにエラーを返す。
func runWindow(url, title string) error {
	scale := enableDPIAwareness()
	px := func(n float64) uint { return uint(n * scale) }
	opts := webview2.WebViewOptions{
		AutoFocus: true,
		// 既定では %AppData%\<exe名> に作られるため、キャッシュ置き場として LocalAppData に置く
		DataPath: webviewDataPath(),
		WindowOptions: webview2.WindowOptions{
			Title:  title,
			Width:  px(1280),
			Height: px(800),
			Center: true,
		},
	}
	w := webview2.NewWithOptions(opts)
	if w == nil {
		return errors.New("WebView2 のウィンドウを作成できませんでした。Microsoft Edge WebView2 ランタイムがインストールされているか確認してください")
	}
	defer w.Destroy()
	w.SetSize(int(px(960)), int(px(600)), webview2.HintMin)
	w.Navigate(url)
	w.Run()
	return nil
}

func webviewDataPath() string {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		return ""
	}
	return filepath.Join(base, "moba-ban-tool", "WebView2")
}

// hasConsole はコンソールウィンドウがあるかどうか。
// 配布用ビルドは -H windowsgui でコンソールなしの exe になる。
func hasConsole() bool {
	h, _, _ := procGetConsole.Call()
	return h != 0
}

// ensureConsole はコンソールがなければ作り、標準入出力をそこへつなぎ直す。
// ウィンドウを使わずにサーバーだけで動かす場合に、URL の表示と終了操作のために使う。
func ensureConsole() {
	if hasConsole() {
		return
	}
	if r, _, _ := procAllocConsole.Call(); r == 0 {
		return
	}
	if out, err := os.OpenFile("CONOUT$", os.O_RDWR, 0); err == nil {
		os.Stdout = out
		os.Stderr = out
	}
	if in, err := os.OpenFile("CONIN$", os.O_RDWR, 0); err == nil {
		os.Stdin = in
	}
	redirectLog()
}

// showError はエラーをメッセージボックスで表示する。コンソールがないときに使う。
func showError(msg string) {
	text, _ := syscall.UTF16PtrFromString(msg)
	caption, _ := syscall.UTF16PtrFromString(appTitle)
	const mbOK, mbIconError = 0x0, 0x10
	procMessageBox.Call(0, uintptr(unsafe.Pointer(text)), uintptr(unsafe.Pointer(caption)), mbOK|mbIconError)
}
