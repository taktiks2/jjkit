package ui

import (
	"testing"

	"github.com/taktiks2/jjkit/internal/jjdiff"
	"github.com/taktiks2/jjkit/internal/jjlog"
)

// テスト用の 2 change ログ。row0=通常(1行)、row1=作業コピー(@, 2行)。
func twoRowLog() *jjlog.Log {
	return &jjlog.Log{Rows: []jjlog.Row{
		{ChangeID: "a", IsWC: false, Lines: []string{"a0"}},
		{ChangeID: "b", IsWC: true, Lines: []string{"b0", "b1"}},
	}}
}

// ログ読込(logLoadedMsg)を受けたら、カーソルが作業コピー(@)の row に合うことを保証する。
// 「起動時・r 再読込で @ にリセット」という決めた挙動の核。
func TestLogLoadedSelectsWorkingCopy(t *testing.T) {
	m := New()
	m.width, m.height, m.ready = 80, 24, true
	updated, _ := m.Update(logLoadedMsg{log: twoRowLog()})
	got := updated.(Model)
	if got.logSel != 1 {
		t.Errorf("logSel = %d, want 1 (the @ row)", got.logSel)
	}
}

// 選択移動が上端・下端でクランプされること（範囲外へ出ない）を保証する。
func TestSelectionClampsAtBounds(t *testing.T) {
	m := New()
	m.width, m.height, m.ready = 80, 24, true
	m.log = twoRowLog()

	m.logSel = 0
	up := moveSelection(m, -1)
	if up.logSel != 0 {
		t.Errorf("up at top: logSel = %d, want 0", up.logSel)
	}

	m.logSel = 1
	down := moveSelection(m, +1)
	if down.logSel != 1 {
		t.Errorf("down at bottom: logSel = %d, want 1", down.logSel)
	}
}

// ↑/↓ が今フォーカスしているペインのカーソルだけを動かすこと（focus=Log なら logSel、
// focus=Files なら fileSel）。これが2階層選択の核。
func TestMoveRoutedByFocus(t *testing.T) {
	m := New()
	m.log = twoRowLog() // 2 rows
	m.files = []jjdiff.FileChange{{Status: "M", Path: "x"}, {Status: "A", Path: "y"}}
	m.logSel, m.fileSel = 0, 0

	m.focus = paneLog
	m = moveSelection(m, +1)
	if m.logSel != 1 || m.fileSel != 0 {
		t.Errorf("Log focus +1: logSel=%d fileSel=%d, want 1,0", m.logSel, m.fileSel)
	}

	m.focus = paneFiles
	m = moveSelection(m, +1)
	if m.fileSel != 1 || m.logSel != 1 {
		t.Errorf("Files focus +1: fileSel=%d logSel=%d, want 1,1", m.fileSel, m.logSel)
	}
}

// Tab が focusCycle（Log → Files → Bookmarks → Log）を巡回し、Diff/Oplog を飛ばすこと。
func TestTabCyclesFocus(t *testing.T) {
	m := New()
	if m.focus != paneLog {
		t.Fatalf("initial focus = %v, want paneLog", m.focus)
	}
	m = m.cycleFocus()
	if m.focus != paneFiles {
		t.Errorf("after 1 Tab: focus = %v, want paneFiles", m.focus)
	}
	m = m.cycleFocus()
	if m.focus != paneBookmarks {
		t.Errorf("after 2 Tabs: focus = %v, want paneBookmarks", m.focus)
	}
	m = m.cycleFocus()
	if m.focus != paneLog {
		t.Errorf("after 3 Tabs: focus = %v, want paneLog (wrap)", m.focus)
	}
}
