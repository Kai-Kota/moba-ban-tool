//go:build !windows

package main

import (
	"errors"
	"fmt"
	"os"
)

// runWindow は Windows 以外では使えない。呼び出し側はブラウザで開く動作に切り替える。
func runWindow(url, title string) error {
	return errors.New("専用ウィンドウは Windows でのみ使えます")
}

func hasConsole() bool { return true }

func ensureConsole() {}

func showError(msg string) { fmt.Fprintln(os.Stderr, msg) }
