package ui

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"

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

// Issue #3 (Q 仕様): refresh 後の logSel は「前選択の change ID を新ログから探す → 無ければ @」。
// これにより describe 後や r キーで「場所を失わない」UX になる。

// 前選択の change ID が new log にもあれば、その row に居続ける。
// 例: describe 後の refresh で同じ change を編集していたら、そこに留まる。
func TestLogLoadedPreservesSelection(t *testing.T) {
	m := New()
	m.width, m.height, m.ready = 80, 24, true
	// 前の log: row0=a, row1=b(@). 選択は row0=a。
	m.log = &jjlog.Log{Rows: []jjlog.Row{
		{ChangeID: "a", Lines: []string{"a0"}},
		{ChangeID: "b", IsWC: true, Lines: []string{"b0"}},
	}}
	m.logSel = 0
	// new log: row0=x, row1=a, row2=b(@). a は row1 に移動。
	newLog := &jjlog.Log{Rows: []jjlog.Row{
		{ChangeID: "x", Lines: []string{"x0"}},
		{ChangeID: "a", Lines: []string{"a0"}},
		{ChangeID: "b", IsWC: true, Lines: []string{"b0"}},
	}}
	updated, _ := m.Update(logLoadedMsg{log: newLog})
	if got := updated.(Model).logSel; got != 1 {
		t.Errorf("logSel = %d, want 1 (a moved to row1)", got)
	}
}

// 前選択の change ID が new log に無ければ WC row に fallback（abandon の典型ケース）。
func TestLogLoadedFallsBackToWC(t *testing.T) {
	m := New()
	m.width, m.height, m.ready = 80, 24, true
	m.log = &jjlog.Log{Rows: []jjlog.Row{
		{ChangeID: "gone", Lines: []string{"x"}},
	}}
	m.logSel = 0
	newLog := &jjlog.Log{Rows: []jjlog.Row{
		{ChangeID: "a", Lines: []string{"a0"}},
		{ChangeID: "b", IsWC: true, Lines: []string{"b0"}},
	}}
	updated, _ := m.Update(logLoadedMsg{log: newLog})
	if got := updated.(Model).logSel; got != 1 {
		t.Errorf("logSel = %d, want 1 (WC fallback)", got)
	}
}

// 初回ロード (m.log == nil): selectedChangeID が "" なので必ず WC fallback。
func TestLogLoadedFirstLoadGoesToWC(t *testing.T) {
	m := New()
	m.width, m.height, m.ready = 80, 24, true
	// m.log は nil (初期状態)
	newLog := &jjlog.Log{Rows: []jjlog.Row{
		{ChangeID: "a", Lines: []string{"a0"}},
		{ChangeID: "b", IsWC: true, Lines: []string{"b0"}},
	}}
	updated, _ := m.Update(logLoadedMsg{log: newLog})
	if got := updated.(Model).logSel; got != 1 {
		t.Errorf("logSel = %d, want 1 (WC on first load)", got)
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

// Issue #3: opResultMsg は new/edit/describe/abandon の jj 実行結果。
// 成功と失敗を 1 型に統合することで opInFlight の解除漏れを構造的に防ぐ。

// 成功時: opInFlight を false に戻し、refresh Cmd（loadLog + filesCmd + diffCmd の batch）を返す。
// Cmd の中身は検査せず「非 nil」までを保証する（bubbles の Cmd は不透明な関数）。
func TestOpResultSuccessClearsFlightAndReturnsRefreshCmd(t *testing.T) {
	m := New()
	m.log = &jjlog.Log{Rows: []jjlog.Row{{ChangeID: "a", Lines: []string{"x"}}}}
	m.opInFlight = true
	updated, cmd := m.Update(opResultMsg{err: nil})
	got := updated.(Model)
	if got.opInFlight {
		t.Errorf("opInFlight = true, want false (cleared on success)")
	}
	if cmd == nil {
		t.Errorf("cmd = nil, want non-nil refresh Cmd")
	}
	if got.err != nil {
		t.Errorf("err = %v, want nil (success should clear err)", got.err)
	}
}

// 失敗時: opInFlight を false に戻し、err を立て、refresh は走らせない。
func TestOpResultErrorClearsFlightAndSetsErr(t *testing.T) {
	m := New()
	m.opInFlight = true
	wantErr := errors.New("immutable")
	updated, cmd := m.Update(opResultMsg{err: wantErr})
	got := updated.(Model)
	if got.opInFlight {
		t.Errorf("opInFlight = true, want false (cleared on err)")
	}
	if got.err != wantErr {
		t.Errorf("err = %v, want %v", got.err, wantErr)
	}
	if cmd != nil {
		t.Errorf("cmd = %v, want nil (no refresh on err)", cmd)
	}
}

// keyPress はテスト用に printable rune の tea.KeyPressMsg を組み立てる。
// Code に加えて Text を入れるのが慣用 (key.Matches は KeyPressMsg.String() を見るが、
// String() は Text が非空ならそれを返すため)。
func keyPress(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: r, Text: string(r)}
}

// Issue #3: 'n' (new) キー — 選択 change に対する jj new を非同期発射、opInFlight ガード付き。

// 'n' 押下: opInFlight=true、Cmd 非 nil を返す。
func TestKeyNewFiresOpAndSetsFlight(t *testing.T) {
	m := New()
	m.log = &jjlog.Log{Rows: []jjlog.Row{{ChangeID: "abc", IsWC: true, Lines: []string{"x"}}}}
	m.logSel = 0
	updated, cmd := m.Update(keyPress('n'))
	got := updated.(Model)
	if !got.opInFlight {
		t.Errorf("opInFlight = false, want true")
	}
	if cmd == nil {
		t.Errorf("cmd = nil, want non-nil jj new Cmd")
	}
}

// opInFlight 中の 'n' は無視 (Cmd nil、フラグ変化なし)。
func TestKeyNewIgnoredWhenOpInFlight(t *testing.T) {
	m := New()
	m.log = &jjlog.Log{Rows: []jjlog.Row{{ChangeID: "abc", IsWC: true, Lines: []string{"x"}}}}
	m.logSel = 0
	m.opInFlight = true
	updated, cmd := m.Update(keyPress('n'))
	if cmd != nil {
		t.Errorf("cmd = %v, want nil (gated by opInFlight)", cmd)
	}
	if !updated.(Model).opInFlight {
		t.Errorf("opInFlight = false, want true (unchanged)")
	}
}

// 選択 change が無い（空 log）場合の 'n' は no-op。
func TestKeyNewIgnoredWhenNoSelection(t *testing.T) {
	m := New()
	// m.log は nil
	updated, cmd := m.Update(keyPress('n'))
	if cmd != nil {
		t.Errorf("cmd = %v, want nil (no selection)", cmd)
	}
	if updated.(Model).opInFlight {
		t.Errorf("opInFlight = true, want false (no op fired)")
	}
}

// Issue #3: 'e' (edit) キー — 選択 change に @ を移す。Task 4 と同形パターン。

// 'e' 押下: opInFlight=true、Cmd 非 nil を返す。
func TestKeyEditFiresOpAndSetsFlight(t *testing.T) {
	m := New()
	m.log = &jjlog.Log{Rows: []jjlog.Row{{ChangeID: "abc", IsWC: true, Lines: []string{"x"}}}}
	m.logSel = 0
	updated, cmd := m.Update(keyPress('e'))
	got := updated.(Model)
	if !got.opInFlight {
		t.Errorf("opInFlight = false, want true")
	}
	if cmd == nil {
		t.Errorf("cmd = nil, want non-nil jj edit Cmd")
	}
}

// opInFlight 中の 'e' は無視。
func TestKeyEditIgnoredWhenOpInFlight(t *testing.T) {
	m := New()
	m.log = &jjlog.Log{Rows: []jjlog.Row{{ChangeID: "abc", IsWC: true, Lines: []string{"x"}}}}
	m.logSel = 0
	m.opInFlight = true
	updated, cmd := m.Update(keyPress('e'))
	if cmd != nil {
		t.Errorf("cmd = %v, want nil (gated)", cmd)
	}
	if !updated.(Model).opInFlight {
		t.Errorf("opInFlight changed, want unchanged")
	}
}
