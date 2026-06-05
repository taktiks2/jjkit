package ui

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/taktiks2/jjkit/internal/jj"
)

// abandonModal は破棄前の確認ダイアログ。Confirm で jj abandon を発射し、Cancel で閉じる。
// target は modal 構築時の change ID を保持する（modal が key を吸うので走行中に動くことはない）。
type abandonModal struct {
	target string
}

func (a abandonModal) HandleKey(msg tea.KeyPressMsg, keys keyMap) (Modal, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Confirm):
		return nil, abandonCmd(a.target)
	case key.Matches(msg, keys.Cancel):
		return nil, nil
	}
	return a, nil // j/k 等のナビゲーションキーは吸う
}

func (a abandonModal) HandleMsg(_ tea.Msg, _ string) (Modal, tea.Cmd, error) {
	return a, nil, nil
}

func (a abandonModal) View(width int) string {
	boxW := min(60, width-4)
	title := "Abandon change?"
	body := fmt.Sprintf("Abandon %s?", a.target)
	hint := "y/Enter: abandon . n/Esc: cancel"
	return modalBoxStyle.Width(boxW).Render(title + "\n\n" + body + "\n\n" + hint)
}

func (a abandonModal) FooterHint() string {
	return modalHintStyle.Render("y/Enter: abandon . n/Esc: cancel")
}

func (a abandonModal) Resize(_ int) Modal { return a }

// abandonCmd は jj abandon を非同期で実行する Cmd。target が空なら nil。
// 結果は opResultMsg として届く（ホストが opInFlight を解除し refresh を発射する）。
func abandonCmd(target string) tea.Cmd {
	if target == "" {
		return nil
	}
	return func() tea.Msg {
		return opResultMsg{err: jj.Abandon(target)}
	}
}
