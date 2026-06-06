// Package jj は jj CLI をサブプロセスとして叩き、その出力を返す（ADR-0001）。
package jj

import (
	"fmt"
	"os/exec"
	"strings"
)

// Root はリポジトリのルートを返す。jj リポ外で実行すると jj 自身のエラーメッセージを
// 載せた error を返す（TUI を起動する前のリポ判定に使う）。
func Root() (string, error) {
	out, err := exec.Command("jj", "root").Output()
	if err != nil {
		return "", wrap("jj root", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func logArgs(template string) []string {
	return []string{"log", "--no-pager", "--color", "always", "-T", template}
}

// LogRaw は指定テンプレートで jj log を実行し、生バイト（ANSI 入り）を返す。
func LogRaw(template string) ([]byte, error) {
	out, err := exec.Command("jj", logArgs(template)...).Output()
	if err != nil {
		return nil, wrap("jj log", err)
	}
	return out, nil
}

func diffArgs(rev string) []string {
	return []string{"diff", "--no-pager", "--color", "always", "-r", rev}
}

func diffFileArgs(rev, path string) []string {
	return []string{"diff", "--no-pager", "--color", "always", "-r", rev, path}
}

func diffSummaryArgs(rev string) []string {
	return []string{"diff", "--no-pager", "--summary", "-r", rev}
}

// Diff は change 全体の diff（ANSI 入り）を返す。
func Diff(rev string) ([]byte, error) {
	out, err := exec.Command("jj", diffArgs(rev)...).Output()
	if err != nil {
		return nil, wrap("jj diff", err)
	}
	return out, nil
}

// DiffFile は指定 change の中の1ファイルだけの diff を返す。
func DiffFile(rev, path string) ([]byte, error) {
	out, err := exec.Command("jj", diffFileArgs(rev, path)...).Output()
	if err != nil {
		return nil, wrap("jj diff", err)
	}
	return out, nil
}

// DiffSummary は変更ファイル一覧（`M path` 形式）を返す。
func DiffSummary(rev string) ([]byte, error) {
	out, err := exec.Command("jj", diffSummaryArgs(rev)...).Output()
	if err != nil {
		return nil, wrap("jj diff", err)
	}
	return out, nil
}

// wrap は exec のエラーを読みやすくする。jj が非ゼロ終了したときは stderr の内容を、
// それ以外（jj が見つからない等）はラップしたエラーを返す。
func wrap(what string, err error) error {
	if ee, ok := err.(*exec.ExitError); ok {
		if msg := strings.TrimSpace(string(ee.Stderr)); msg != "" {
			return fmt.Errorf("%s", msg)
		}
	}
	return fmt.Errorf("%s: %w", what, err)
}

func newArgs(change string) []string {
	return []string{"new", "--no-pager", change}
}

// New は <change> を親に持つ空の新 change を作り、@ をそこに移す。
func New(change string) error {
	_, err := exec.Command("jj", newArgs(change)...).Output()
	if err != nil {
		return wrap("jj new", err)
	}
	return nil
}

func editArgs(change string) []string {
	return []string{"edit", "--no-pager", change}
}

// Edit は @ を <change> に移す。
func Edit(change string) error {
	_, err := exec.Command("jj", editArgs(change)...).Output()
	if err != nil {
		return wrap("jj edit", err)
	}
	return nil
}

func describeArgs(change, msg string) []string {
	return []string{"describe", "--no-pager", "-r", change, "-m", msg}
}

// Describe は <change> の description を <msg> に置き換える。空文字も許容。
func Describe(change, msg string) error {
	_, err := exec.Command("jj", describeArgs(change, msg)...).Output()
	if err != nil {
		return wrap("jj describe", err)
	}
	return nil
}

func abandonArgs(change string) []string {
	return []string{"abandon", "--no-pager", change}
}

// Abandon は <change> を破棄する。
func Abandon(change string) error {
	_, err := exec.Command("jj", abandonArgs(change)...).Output()
	if err != nil {
		return wrap("jj abandon", err)
	}
	return nil
}

func descriptionArgs(change string) []string {
	return []string{"log", "--no-pager", "--no-graph", "-T", "description", "-r", change}
}

// Description は <change> の description 文字列を返す（describe modal の seed 用）。
// 末尾改行を trim して返す。空 description でも空文字列が返り、エラーにはならない。
func Description(change string) (string, error) {
	out, err := exec.Command("jj", descriptionArgs(change)...).Output()
	if err != nil {
		return "", wrap("jj log", err)
	}
	return strings.TrimRight(string(out), "\n"), nil
}

func bookmarkListArgs(template string) []string {
	return []string{"bookmark", "list", "--no-pager", "--all-remotes", "-T", template}
}

// BookmarkListRaw は --all-remotes + 指定テンプレートで bookmark list を実行し、生バイトを返す。
func BookmarkListRaw(template string) ([]byte, error) {
	out, err := exec.Command("jj", bookmarkListArgs(template)...).Output()
	if err != nil {
		return nil, wrap("jj bookmark list", err)
	}
	return out, nil
}

func bookmarkMoveArgs(name, to string) []string {
	return []string{"bookmark", "move", "--no-pager", name, "--to", to}
}

// BookmarkMove は bookmark <name> を target に進める（既存 ref を上書き、新規作成はしない）。
// jj が拒否する場合は stderr を error に乗せる（wrap が hook 済み）。
func BookmarkMove(name, to string) error {
	_, err := exec.Command("jj", bookmarkMoveArgs(name, to)...).Output()
	if err != nil {
		return wrap("jj bookmark move", err)
	}
	return nil
}
