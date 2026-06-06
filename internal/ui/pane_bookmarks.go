package ui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	"github.com/taktiks2/jjkit/internal/jj"
	"github.com/taktiks2/jjkit/internal/jjbookmark"
)

// bookmarksPane は local bookmark + remote-tracking 一覧を表示するペイン。
// 1 bookmark = 1 行。行の中で各 remote の ahead/behind を [↑n↓m] / ✓ で並べる。
type bookmarksPane struct {
	items []jjbookmark.Bookmark
	sel   int
	vp    viewport.Model
}

// bookmarksLoadedMsg は jj bookmark list の成功結果。
type bookmarksLoadedMsg struct{ items []jjbookmark.Bookmark }

// bookmarksCmd は bookmark list を取りに行く Cmd。成功時 bookmarksLoadedMsg、失敗時 logErrMsg。
func bookmarksCmd() tea.Msg {
	raw, err := jj.BookmarkListRaw(jjbookmark.Template)
	if err != nil {
		return logErrMsg{err}
	}
	return bookmarksLoadedMsg{items: jjbookmark.Parse(raw)}
}

// Apply は新リストを取り込み、選択を範囲内にクランプする。
// abandon 等で items が短くなったときの out-of-range 防御。
func (p *bookmarksPane) Apply(items []jjbookmark.Bookmark) {
	p.items = items
	if p.sel >= len(items) {
		p.sel = max(len(items)-1, 0)
	}
	p.refresh()
}

// Move はカーソルを delta だけ動かす（範囲内にクランプ）。
func (p *bookmarksPane) Move(delta int) {
	if len(p.items) == 0 {
		return
	}
	p.sel = max(0, min(p.sel+delta, len(p.items)-1))
	p.refresh()
}

// SelectedName は選択中 bookmark の名前（無ければ ""）。
func (p bookmarksPane) SelectedName() string {
	if p.sel < 0 || p.sel >= len(p.items) {
		return ""
	}
	return p.items[p.sel].Name
}

// Resize は applyLayout から呼ぶ。外寸 rect から枠/タイトル分を引いた内寸を viewport に渡し、
// 新しい幅で再描画する（選択ハイライトは行幅依存）。
func (p *bookmarksPane) Resize(r rect) {
	setVP(&p.vp, r)
	p.refresh()
}

// View は viewport の現状を返す（paneBox の中身に流す）。
func (p bookmarksPane) View() string { return p.vp.View() }

// refresh は items を1行に整形して viewport に流す。選択行ハイライトを乗せる。
func (p *bookmarksPane) refresh() {
	lines := make([]string, len(p.items))
	for i, b := range p.items {
		lines[i] = formatBookmark(b)
	}
	p.vp.SetContent(RenderContent(lines, p.sel, p.sel+1, p.vp.Width()))
}

// formatBookmark は1 bookmark を1行表現に。例:
//
//	feature → kqlu1234 [↑2]
//	main → c9d4e58b ✓
//	wip → omykzzzz                (remote 無し)
func formatBookmark(b jjbookmark.Bookmark) string {
	change := "-"
	if b.LocalTarget != nil {
		change = b.LocalTarget.ChangeID
	}
	parts := []string{fmt.Sprintf("%s → %s", b.Name, change)}
	for _, r := range b.Remotes {
		parts = append(parts, formatRemote(r))
	}
	return strings.Join(parts, " ")
}

// formatRemote は1 remote の divergence マーカー。同期済 ✓、それ以外は [↑n↓m]（0 は省略）。
func formatRemote(r jjbookmark.RemoteRef) string {
	if r.Ahead == 0 && r.Behind == 0 {
		return "✓"
	}
	mark := ""
	if r.Ahead > 0 {
		mark += fmt.Sprintf("↑%d", r.Ahead)
	}
	if r.Behind > 0 {
		mark += fmt.Sprintf("↓%d", r.Behind)
	}
	return "[" + mark + "]"
}
