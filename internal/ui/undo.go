package ui

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/taktiks2/jjkit/internal/jj"
)

// undoModal は直前 operation を取り消す確認ダイアログ。Confirm で jj undo を発射し、Cancel で閉じる。
// undo は引数を取らないので target フィールドは無い。
type undoModal struct{}

func (u undoModal) HandleKey(msg tea.KeyPressMsg, keys keyMap) (Modal, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Confirm):
		return nil, undoCmd()
	case key.Matches(msg, keys.Cancel):
		return nil, nil
	}
	return u, nil // j/k 等のナビゲーションキーは吸う
}

func (u undoModal) HandleMsg(_ tea.Msg, _ string) (Modal, tea.Cmd, error) {
	return u, nil, nil
}

func (u undoModal) View(width int) string {
	boxW := min(60, width-4)
	title := "Undo last operation?"
	body := "Reverts the most recent jj operation."
	hint := "y/Enter: undo . n/Esc: cancel"
	return modalBoxStyle.Width(boxW).Render(title + "\n\n" + body + "\n\n" + hint)
}

func (u undoModal) FooterHint() string {
	return modalHintStyle.Render("y/Enter: undo . n/Esc: cancel")
}

func (u undoModal) Resize(_ int) Modal { return u }

// undoCmd は jj undo を非同期で実行する Cmd。結果は opResultMsg として届き、
// ホストが opInFlight を解除して全 pane refresh を発射する。
func undoCmd() tea.Cmd {
	return func() tea.Msg {
		return opResultMsg{err: jj.Undo()}
	}
}
