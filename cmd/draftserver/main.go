// draftserver は BAN/PICK オーバーレイのサーバー。
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"

	"moba-ban-tool/internal/config"
	"moba-ban-tool/internal/draft"
	"moba-ban-tool/internal/hub"
	"moba-ban-tool/internal/server"
	"moba-ban-tool/internal/store"
	"moba-ban-tool/web"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "待ち受けアドレス(別PCから操作する場合は 0.0.0.0:8080)")
	dataFlag := flag.String("data", "", "characters.json などを置くディレクトリ(既定: カレントか exe の隣の data)")
	stateFlag := flag.String("state", "", "状態の保存先(既定: data ディレクトリと同じ場所の state.json)")
	open := flag.Bool("open", true, "起動時にブラウザで操作パネルを開く(配布用ビルドのみ)")
	flag.Parse()

	ui := web.Dist()
	embedded := ui != nil

	dataDir, err := resolveDataDir(*dataFlag)
	if err != nil {
		fatal(embedded, err)
	}
	statePath := *stateFlag
	if statePath == "" {
		statePath = filepath.Join(filepath.Dir(dataDir), "state.json")
	}

	cfg, err := config.Load(dataDir)
	if err != nil {
		fatal(embedded, err)
	}
	engine := draft.NewEngine(cfg.DraftOrder, cfg.CharacterIDs())

	st := store.New(statePath)
	state, ok, err := st.Load()
	switch {
	case err != nil:
		log.Printf("%s を読めないため初期状態で起動します: %v", statePath, err)
		state = engine.Initial()
	case ok:
		engine.Normalize(&state)
		log.Printf("%s から状態を復元しました", statePath)
	default:
		state = engine.Initial()
	}

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		if errors.Is(err, syscall.EADDRINUSE) || isAddrInUse(err) {
			err = fmt.Errorf("%s は使用中です。draftserver がすでに起動していないか確認するか、-addr で別のポートを指定してください(例: -addr 127.0.0.1:8081)", *addr)
		}
		fatal(embedded, err)
	}

	h := hub.New(engine, state, st.Save)
	base := "http://" + displayHost(ln.Addr().(*net.TCPAddr))
	log.Printf("データ: %s", dataDir)
	log.Printf("状態の保存先: %s", statePath)
	if ui == nil {
		log.Printf("開発用ビルドです。フロントは npm --prefix web run dev で起動してください")
	} else {
		fmt.Println()
		fmt.Println("  操作パネル  " + base + "/control")
		fmt.Println("  OBS 用      " + base + "/overlay   (ブラウザソース 1920x1080)")
		fmt.Println()
		fmt.Println("  終了するにはこのウィンドウを閉じるか Ctrl+C を押してください")
		fmt.Println()
		if *open {
			openBrowser(base + "/control")
		}
	}
	fatal(embedded, http.Serve(ln, server.New(cfg, h, dataDir, ui)))
}

// resolveDataDir は data ディレクトリを探す。
// exe をダブルクリックした場合とターミナルから起動した場合で
// カレントディレクトリが変わるので、カレント → exe の隣 の順に探す。
func resolveDataDir(flagValue string) (string, error) {
	if flagValue != "" {
		return filepath.Abs(flagValue)
	}
	candidates := []string{"data"}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "data"))
	}
	for _, c := range candidates {
		if fi, err := os.Stat(c); err == nil && fi.IsDir() {
			return filepath.Abs(c)
		}
	}
	return "", fmt.Errorf("data フォルダが見つかりません。exe と同じ場所に data フォルダを置くか、-data で指定してください")
}

// displayHost は表示用のホスト名。0.0.0.0 で待ち受けている場合は localhost と表示する。
func displayHost(a *net.TCPAddr) string {
	if a.IP.IsUnspecified() {
		return fmt.Sprintf("localhost:%d", a.Port)
	}
	return a.String()
}

func isAddrInUse(err error) bool {
	// Windows の WSAEADDRINUSE は syscall.EADDRINUSE と一致しないことがあるため番号で判定する
	var errno syscall.Errno
	return errors.As(err, &errno) && errno == 10048
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		log.Printf("ブラウザを開けませんでした: %v", err)
	}
}

// fatal はエラーを表示して終了する。
// 配布用ビルドは exe をダブルクリックして使うため、すぐにウィンドウが閉じてエラーが読めないことがないよう Enter を待つ。
func fatal(waitEnter bool, err error) {
	log.Print("エラー: ", err)
	if waitEnter {
		fmt.Print("\nEnter キーで終了します...")
		bufio.NewReader(os.Stdin).ReadString('\n')
	}
	os.Exit(1)
}
