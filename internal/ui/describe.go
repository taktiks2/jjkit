package ui

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/taktiks2/jjkit/internal/jj"
)

// describeModal は describe フローを内包する 2 状態機械。
//
//	loading=true : seed fetch 中（descSeedCmd の完了待ち）。Esc でキャンセル可。
//	loading=false: textinput で編集中。Enter で submit、Esc でキャンセル。
//
// target は modal 構築時の change ID。submit 時点で @ が動いていても
// describe 対象は当初の target のままで良い。
type describeModal struct {
	target  string
	loading bool
	input   textinput.Model
}

// newDescribeModal は loading 状態の modal を返す。width は modal box の外寸計算に使う。
func newDescribeModal(target string, width int) describeModal {
	in := textinput.New()
	in.SetWidth(descInputWidth(width))
	return describeModal{target: target, loading: true, input: in}
}

// descInputWidth は modal box（min(60, width-4)）から padding 2*2 を引いた input 内寸。
func descInputWidth(width int) int {
	boxW := min(60, width-4)
	return max(boxW-4, 0)
}

func (d describeModal) HandleKey(msg tea.KeyPressMsg, keys keyMap) (Modal, tea.Cmd) {
	if d.loading {
		if key.Matches(msg, keys.Cancel) {
			return nil, nil
		}
		return d, nil // それ以外は吸う
	}
	switch {
	case key.Matches(msg, keys.Submit):
		return nil, describeCmd(d.target, d.input.Value())
	case key.Matches(msg, keys.Cancel):
		return nil, nil
	}
	var cmd tea.Cmd
	d.input, cmd = d.input.Update(msg)
	return d, cmd
}

func (d describeModal) HandleMsg(msg tea.Msg, currentChange string) (Modal, tea.Cmd, error) {
	loaded, ok := msg.(descLoadedMsg)
	if !ok {
		return d, nil, nil
	}
	if !d.loading {
		return d, nil, nil // 既に編集状態（多重 descLoadedMsg 防御）
	}
	if loaded.target != currentChange {
		return d, nil, nil // stale: 選択が動いた
	}
	if loaded.err != nil {
		return nil, nil, loaded.err // 閉じてホストに err を上げる
	}
	d.loading = false
	d.input.SetValue(loaded.desc)
	d.input.CursorEnd()
	return d, d.input.Focus(), nil
}

func (d describeModal) View(width int) string {
	boxW := min(60, width-4)
	if d.loading {
		body := fmt.Sprintf("loading description for %s...", d.target)
		hint := "Esc: cancel"
		return modalBoxStyle.Width(boxW).Render(body + "\n\n" + hint)
	}
	title := fmt.Sprintf("Edit description for %s", d.target)
	hint := "Enter: save . Esc: cancel"
	return modalBoxStyle.Width(boxW).Render(title + "\n\n" + d.input.View() + "\n\n" + hint)
}

func (d describeModal) FooterHint() string {
	if d.loading {
		return modalHintStyle.Render("Esc: cancel")
	}
	return modalHintStyle.Render("Enter: save . Esc: cancel")
}

func (d describeModal) Resize(width int) Modal {
	d.input.SetWidth(descInputWidth(width))
	return d
}

// descSeedCmd は describe modal の seed として現在の description を非同期取得する Cmd。
// 結果は descLoadedMsg として届く（target で stale 判定）。
func descSeedCmd(target string) tea.Cmd {
	if target == "" {
		return nil
	}
	return func() tea.Msg {
		desc, err := jj.Description(target)
		return descLoadedMsg{target: target, desc: desc, err: err}
	}
}

// describeCmd は target の description を msg に置き換える jj describe を非同期で実行する Cmd。
// 結果は opResultMsg として届く。
func describeCmd(target, msg string) tea.Cmd {
	if target == "" {
		return nil
	}
	return func() tea.Msg {
		return opResultMsg{err: jj.Describe(target, msg)}
	}
}
