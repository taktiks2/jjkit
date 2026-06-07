package ui

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/taktiks2/jjkit/internal/jj"
)

// restoreModal は jj op restore <op-id> の確認ダイアログ。
// target は modal 構築時に固定した op-id を保持する（modal が key を吸うため走行中に動くことはない）。
type restoreModal struct {
	target string
}

func (r restoreModal) HandleKey(msg tea.KeyPressMsg, keys keyMap) (Modal, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Confirm):
		return nil, opRestoreCmd(r.target)
	case key.Matches(msg, keys.Cancel):
		return nil, nil
	}
	return r, nil // j/k 等のナビゲーションキーは吸う
}

func (r restoreModal) HandleMsg(_ tea.Msg, _ string) (Modal, tea.Cmd, error) {
	return r, nil, nil
}

func (r restoreModal) View(width int) string {
	boxW := min(60, width-4)
	title := "Restore to operation?"
	body := fmt.Sprintf("Rewind the repository to operation %s.", r.target)
	hint := "y/Enter: restore . n/Esc: cancel"
	return modalBoxStyle.Width(boxW).Render(title + "\n\n" + body + "\n\n" + hint)
}

func (r restoreModal) FooterHint() string {
	return modalHintStyle.Render("y/Enter: restore . n/Esc: cancel")
}

func (r restoreModal) Resize(_ int) Modal { return r }

// opRestoreCmd は jj op restore <target> を非同期で実行する Cmd。target 空なら nil。
func opRestoreCmd(target string) tea.Cmd {
	if target == "" {
		return nil
	}
	return func() tea.Msg {
		return opResultMsg{err: jj.OpRestore(target)}
	}
}
