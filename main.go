// Package main implements jjkit, a terminal UI for Jujutsu (jj) — "lazygit for jj".
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/taktiks2/jjkit/internal/jj"
	"github.com/taktiks2/jjkit/internal/ui"
)

func main() {
	// 事前チェック: jj リポ外なら TUI を起動せず、jj のメッセージを出して exit 1（panic しない）。
	if _, err := jj.Root(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if _, err := tea.NewProgram(ui.New()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
