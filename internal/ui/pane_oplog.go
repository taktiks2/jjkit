package ui

import (
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	"github.com/taktiks2/jjkit/internal/jj"
	"github.com/taktiks2/jjkit/internal/jjop"
)

// oplogPane は jj operation log を表示するペイン。
// 1 operation = 1 Row（複数行を含むことがある）。jjlog と同じ sentinel パース方式 (ADR-0004) に乗る。
// 親ペイン: 選択 op は jj op restore の引数になる（undo は選択不要・常に直前 op が対象）。
type oplogPane struct {
	log *jjop.OpLog
	sel int
	vp  viewport.Model
}

// oplogLoadedMsg は jj op log の成功結果。
type oplogLoadedMsg struct{ log *jjop.OpLog }

// oplogCmd は op log を取りに行く Cmd。成功時 oplogLoadedMsg、失敗時 logErrMsg。
func oplogCmd() tea.Msg {
	raw, err := jj.OpLogRaw(jjop.Template)
	if err != nil {
		return logErrMsg{err}
	}
	parsed, err := jjop.Parse(raw)
	if err != nil {
		return logErrMsg{err}
	}
	return oplogLoadedMsg{parsed}
}

// Apply は新 op log を取り込み、前選択 op-id を新ログから探して居続けさせる。
// 無ければ先頭（最新 op）に fallback する。
func (p *oplogPane) Apply(newLog *jjop.OpLog) {
	prevID := p.SelectedID()
	p.log = newLog
	if i := p.log.RowByID(prevID); i >= 0 {
		p.sel = i
	} else {
		p.sel = 0
	}
	p.refresh()
	p.scroll()
}

// Move はカーソルを delta だけ動かす（範囲内にクランプ）。
func (p *oplogPane) Move(delta int) {
	if p.log == nil || len(p.log.Rows) == 0 {
		return
	}
	p.sel = max(0, min(p.sel+delta, len(p.log.Rows)-1))
	p.refresh()
	p.scroll()
}

// SelectedID は今選んでいる op の id（無ければ ""）。restore の引数に使う。
func (p oplogPane) SelectedID() string {
	if p.log == nil || p.sel < 0 || p.sel >= len(p.log.Rows) {
		return ""
	}
	return p.log.Rows[p.sel].ID
}

// Resize は applyLayout から呼ぶ。外寸 rect から枠2+タイトル1 を引いた内寸を viewport に渡し、
// 新しい幅で再描画する（選択ハイライトは行幅依存）。
func (p *oplogPane) Resize(r rect) {
	setVP(&p.vp, r)
	p.refresh()
}

// View は viewport の現状を文字列で返す（paneBox の中身に流す）。
func (p oplogPane) View() string { return p.vp.View() }

// refresh は選択 op の行範囲をハイライトして viewport に流す。
func (p *oplogPane) refresh() {
	if p.log == nil {
		return
	}
	start, end := p.log.LineRange(p.sel)
	p.vp.SetContent(RenderContent(p.log.Lines(), start, end, p.vp.Width()))
}

// scroll は選択 op が画面内に収まるようスクロールする。
func (p *oplogPane) scroll() {
	if p.log == nil || p.vp.Height() <= 0 {
		return
	}
	start, end := p.log.LineRange(p.sel)
	p.vp.EnsureVisible(end-1, 0, 0)
	p.vp.EnsureVisible(start, 0, 0)
}
