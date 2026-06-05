package ui

import (
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	"github.com/taktiks2/jjkit/internal/jj"
)

// diffReq は Diff ペインに出すべき内容の識別子。file=="" は change 全体を表す。
// 非同期ロードの stale 判定にも使う。
type diffReq struct {
	change string
	file   string
}

// diffPane は diff 本文を表示するペイン。
type diffPane struct {
	content string
	vp      viewport.Model
}

// diffLoadedMsg は jj diff の結果。req は「何を要求したか」で stale 判定する。
type diffLoadedMsg struct {
	req diffReq
	raw []byte
}

// diffCmd は req の diff を非同期に取りに行く Cmd を返す。change 空なら nil。
func diffCmd(req diffReq) tea.Cmd {
	if req.change == "" {
		return nil
	}
	return func() tea.Msg {
		var (
			raw []byte
			err error
		)
		if req.file == "" {
			raw, err = jj.Diff(req.change)
		} else {
			raw, err = jj.DiffFile(req.change, req.file)
		}
		if err != nil {
			return logErrMsg{err}
		}
		return diffLoadedMsg{req: req, raw: raw}
	}
}

// Apply は diffLoadedMsg を取り込む。expectedReq と一致しない結果は stale で捨てる。
func (p *diffPane) Apply(msg diffLoadedMsg, expectedReq diffReq) {
	if msg.req != expectedReq {
		return
	}
	p.content = string(msg.raw)
	p.refresh()
}

// Resize は applyLayout から呼ぶ。
func (p *diffPane) Resize(r rect) {
	setVP(&p.vp, r)
	p.refresh()
}

// View は viewport の現状を返す。
func (p diffPane) View() string { return p.vp.View() }

// refresh は保持している diff 本文を viewport に流し、先頭へ戻す。
func (p *diffPane) refresh() {
	p.vp.SetContent(p.content)
	p.vp.GotoTop()
}
