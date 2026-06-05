package ui

import (
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	"github.com/taktiks2/jjkit/internal/jj"
	"github.com/taktiks2/jjkit/internal/jjdiff"
)

// filesPane は選択中 change の変更ファイル一覧。
type filesPane struct {
	items []jjdiff.FileChange
	sel   int
	vp    viewport.Model
}

// filesLoadedMsg は jj diff --summary の結果。change で stale 判定する。
type filesLoadedMsg struct {
	change string
	files  []jjdiff.FileChange
}

// filesCmd は change の変更ファイル一覧を取りに行く Cmd を返す。change 空なら nil。
func filesCmd(change string) tea.Cmd {
	if change == "" {
		return nil
	}
	return func() tea.Msg {
		raw, err := jj.DiffSummary(change)
		if err != nil {
			return logErrMsg{err}
		}
		return filesLoadedMsg{change: change, files: jjdiff.ParseSummary(raw)}
	}
}

// Apply は filesLoadedMsg を取り込む。currentChange と一致しない結果は stale で捨てる。
// 反映時は sel を 0 にリセット（前の change の位置は無効）。
func (p *filesPane) Apply(msg filesLoadedMsg, currentChange string) {
	if msg.change != currentChange {
		return
	}
	p.items = msg.files
	p.sel = 0
	p.refresh()
}

// Move はカーソルを delta だけ動かす（範囲内にクランプ）。
func (p *filesPane) Move(delta int) {
	if len(p.items) == 0 {
		return
	}
	p.sel = max(0, min(p.sel+delta, len(p.items)-1))
	p.refresh()
}

// SelectedPath は今選んでいるファイルのパス（無ければ ""）。
func (p filesPane) SelectedPath() string {
	if p.sel < 0 || p.sel >= len(p.items) {
		return ""
	}
	return p.items[p.sel].Path
}

// Resize は applyLayout から呼ぶ。
func (p *filesPane) Resize(r rect) {
	setVP(&p.vp, r)
	p.refresh()
}

// View は viewport の現状を返す。
func (p filesPane) View() string { return p.vp.View() }

// refresh は変更ファイル一覧を「ステータス + パス」の1行に整形し、選択ハイライト付きで流す。
func (p *filesPane) refresh() {
	lines := make([]string, len(p.items))
	for i, f := range p.items {
		lines[i] = f.Status + " " + f.Path
	}
	p.vp.SetContent(RenderContent(lines, p.sel, p.sel+1, p.vp.Width()))
}
