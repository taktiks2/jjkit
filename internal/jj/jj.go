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
