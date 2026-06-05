package ui

import (
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	"github.com/taktiks2/jjkit/internal/jj"
	"github.com/taktiks2/jjkit/internal/jjlog"
)

// logPane は change graph ペイン。jjlog.Log・選択行・viewport を保持する。
// 親ペイン: ここの選択 change が filesPane と diffPane の取得対象を決める。
type logPane struct {
	log *jjlog.Log
	sel int
	vp  viewport.Model
}

// logLoadedMsg は jj log の成功結果。
type logLoadedMsg struct{ log *jjlog.Log }

// loadLog は jj log を実行してパースする Cmd。成功時 logLoadedMsg、失敗時 logErrMsg。
func loadLog() tea.Msg {
	raw, err := jj.LogRaw(jjlog.Template)
	if err != nil {
		return logErrMsg{err}
	}
	parsed, err := jjlog.Parse(raw)
	if err != nil {
		return logErrMsg{err}
	}
	return logLoadedMsg{parsed}
}

// Apply は新ログを取り込み、前選択 change を新ログから探して居続けさせる。
// 無ければ作業コピー (@) の row に fallback する（abandon の典型ケース）。
func (p *logPane) Apply(newLog *jjlog.Log, prevID string) {
	p.log = newLog
	if i := p.log.RowByChangeID(prevID); i >= 0 {
		p.sel = i
	} else {
		p.sel = p.log.WorkingCopyRow()
	}
	p.refresh()
	p.scroll()
}

// Move はカーソルを delta だけ動かす（範囲内にクランプ）。
func (p *logPane) Move(delta int) {
	if p.log == nil || len(p.log.Rows) == 0 {
		return
	}
	p.sel = max(0, min(p.sel+delta, len(p.log.Rows)-1))
	p.refresh()
	p.scroll()
}

// SelectedChangeID は今選んでいる change の id（無ければ ""）。
func (p logPane) SelectedChangeID() string {
	if p.log == nil || p.sel < 0 || p.sel >= len(p.log.Rows) {
		return ""
	}
	return p.log.Rows[p.sel].ChangeID
}

// Resize は applyLayout から呼ぶ。外寸 rect から枠2+タイトル1 を引いた内寸を viewport に渡し、
// 新しい幅で再描画する（選択ハイライトは行幅依存）。
func (p *logPane) Resize(r rect) {
	setVP(&p.vp, r)
	p.refresh()
}

// View は viewport の現状を文字列で返す（paneBox の中身に流す）。
func (p logPane) View() string { return p.vp.View() }

// refresh は選択行ハイライトを乗せた jj log を viewport に流す。
func (p *logPane) refresh() {
	if p.log == nil {
		return
	}
	start, end := p.log.LineRange(p.sel)
	p.vp.SetContent(RenderContent(p.log.Lines(), start, end, p.vp.Width()))
}

// scroll は選択 change が画面内に収まるようスクロールする。
func (p *logPane) scroll() {
	if p.log == nil || p.vp.Height() <= 0 {
		return
	}
	start, end := p.log.LineRange(p.sel)
	p.vp.EnsureVisible(end-1, 0, 0)
	p.vp.EnsureVisible(start, 0, 0)
}
