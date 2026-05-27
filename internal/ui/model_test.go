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

// currentDiffReq が (focus, 選択) から「今 Diff に出すべき内容」を正しく導くこと。
// Log focus なら change 全体（file=""）、Files focus なら選択ファイル単位。
func TestCurrentDiffReqByFocus(t *testing.T) {
	m := New()
	m.log = &jjlog.Log{Rows: []jjlog.Row{{ChangeID: "aaaa", IsWC: true, Lines: []string{"x"}}}}
	m.logSel = 0
	m.files = []jjdiff.FileChange{{Status: "M", Path: "src/x.go"}}
	m.fileSel = 0

	m.focus = paneLog
	if got := m.currentDiffReq(); got != (diffReq{change: "aaaa"}) {
		t.Errorf("Log focus: got %+v, want whole-change {aaaa}", got)
	}

	m.focus = paneFiles
	if got := m.currentDiffReq(); got != (diffReq{change: "aaaa", file: "src/x.go"}) {
		t.Errorf("Files focus: got %+v, want file-level {aaaa, src/x.go}", got)
	}
}

// 非同期ロード結果の stale ガード：移動後に届いた古い diff（req が今の選択と一致しない）は
// 捨て、現在の選択と一致するものだけ反映すること。Cmd の発火順とは独立にこのロジックが
// 正しいことを保証する（高速移動・順序逆転シナリオの核）。
func TestStaleDiffResultIgnored(t *testing.T) {
	m := New()
	m.log = &jjlog.Log{Rows: []jjlog.Row{{ChangeID: "aaaa", Lines: []string{"x"}}}}
	m.logSel = 0
	m.focus = paneLog // currentDiffReq == {change: "aaaa"}

	// 別 change 宛の結果は無視される。
	stale, _ := m.Update(diffLoadedMsg{req: diffReq{change: "bbbb"}, raw: []byte("STALE")})
	if got := stale.(Model).diffContent; got != "" {
		t.Errorf("stale result applied: %q", got)
	}

	// 現在の選択と一致する結果は反映される。
	fresh, _ := m.Update(diffLoadedMsg{req: diffReq{change: "aaaa"}, raw: []byte("FRESH")})
	if got := fresh.(Model).diffContent; got != "FRESH" {
		t.Errorf("matching result not applied: %q", got)
	}
}

// filesLoadedMsg: 選択中の change と一致する結果なら files を入れ、fileSel を 0 にリセット。
// 別 change 宛の古い結果は無視（こちらも stale ガード）。change が変わったら前の fileSel が
// 範囲外になる可能性があるので、リセットは正当性のために必要。
func TestFilesLoadedSetsFilesAndResetsSelection(t *testing.T) {
	m := New()
	m.log = &jjlog.Log{Rows: []jjlog.Row{{ChangeID: "aaaa", Lines: []string{"x"}}}}
	m.logSel = 0
	m.fileSel = 3 // 前の change で動かしていた古い位置

	files := []jjdiff.FileChange{{Status: "M", Path: "a"}, {Status: "A", Path: "b"}}
	updated, _ := m.Update(filesLoadedMsg{change: "aaaa", files: files})
	got := updated.(Model)
	if len(got.files) != 2 {
		t.Fatalf("files len = %d, want 2", len(got.files))
	}
	if got.fileSel != 0 {
		t.Errorf("fileSel = %d, want 0 (reset)", got.fileSel)
	}

	// 別 change 宛は無視される。
	stale, _ := m.Update(filesLoadedMsg{change: "zzzz", files: files})
	if len(stale.(Model).files) != 0 {
		t.Errorf("stale files applied: %v", stale.(Model).files)
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
