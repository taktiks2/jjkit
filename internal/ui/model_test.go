package ui

import (
	"errors"
	"strings"
	"testing"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/taktiks2/jjkit/internal/jjbookmark"
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
func TestLogLoadedPreservesSelection(t *testing.T) {
	m := New()
	m.width, m.height, m.ready = 80, 24, true
	m.log.log = &jjlog.Log{Rows: []jjlog.Row{
		{ChangeID: "a", Lines: []string{"a0"}},
		{ChangeID: "b", IsWC: true, Lines: []string{"b0"}},
	}}
	m.log.sel = 0
	newLog := &jjlog.Log{Rows: []jjlog.Row{
		{ChangeID: "x", Lines: []string{"x0"}},
		{ChangeID: "a", Lines: []string{"a0"}},
		{ChangeID: "b", IsWC: true, Lines: []string{"b0"}},
	}}
	updated, _ := m.Update(logLoadedMsg{log: newLog})
	if got := updated.(Model).log.sel; got != 1 {
		t.Errorf("logSel = %d, want 1 (a moved to row1)", got)
	}
}

// 前選択の change ID が new log に無ければ WC row に fallback。
func TestLogLoadedFallsBackToWC(t *testing.T) {
	m := New()
	m.width, m.height, m.ready = 80, 24, true
	m.log.log = &jjlog.Log{Rows: []jjlog.Row{
		{ChangeID: "gone", Lines: []string{"x"}},
	}}
	m.log.sel = 0
	newLog := &jjlog.Log{Rows: []jjlog.Row{
		{ChangeID: "a", Lines: []string{"a0"}},
		{ChangeID: "b", IsWC: true, Lines: []string{"b0"}},
	}}
	updated, _ := m.Update(logLoadedMsg{log: newLog})
	if got := updated.(Model).log.sel; got != 1 {
		t.Errorf("logSel = %d, want 1 (WC fallback)", got)
	}
}

// 初回ロード (m.log == nil): selectedChangeID が "" なので必ず WC fallback。
func TestLogLoadedFirstLoadGoesToWC(t *testing.T) {
	m := New()
	m.width, m.height, m.ready = 80, 24, true
	newLog := &jjlog.Log{Rows: []jjlog.Row{
		{ChangeID: "a", Lines: []string{"a0"}},
		{ChangeID: "b", IsWC: true, Lines: []string{"b0"}},
	}}
	updated, _ := m.Update(logLoadedMsg{log: newLog})
	if got := updated.(Model).log.sel; got != 1 {
		t.Errorf("logSel = %d, want 1 (WC on first load)", got)
	}
}

// 選択移動が上端・下端でクランプされること。
func TestSelectionClampsAtBounds(t *testing.T) {
	m := New()
	m.width, m.height, m.ready = 80, 24, true
	m.log.log = twoRowLog()

	m.log.sel = 0
	up := moveSelection(m, -1)
	if up.log.sel != 0 {
		t.Errorf("up at top: logSel = %d, want 0", up.log.sel)
	}

	m.log.sel = 1
	down := moveSelection(m, +1)
	if down.log.sel != 1 {
		t.Errorf("down at bottom: logSel = %d, want 1", down.log.sel)
	}
}

// ↑/↓ がフォーカスペインのカーソルだけを動かす。
func TestMoveRoutedByFocus(t *testing.T) {
	m := New()
	m.log.log = twoRowLog()
	m.files.items = []jjdiff.FileChange{{Status: "M", Path: "x"}, {Status: "A", Path: "y"}}
	m.log.sel, m.files.sel = 0, 0

	m.focus = paneLog
	m = moveSelection(m, +1)
	if m.log.sel != 1 || m.files.sel != 0 {
		t.Errorf("Log focus +1: logSel=%d fileSel=%d, want 1,0", m.log.sel, m.files.sel)
	}

	m.focus = paneFiles
	m = moveSelection(m, +1)
	if m.files.sel != 1 || m.log.sel != 1 {
		t.Errorf("Files focus +1: fileSel=%d logSel=%d, want 1,1", m.files.sel, m.log.sel)
	}
}

// currentDiffReq が (focus, 選択) から「今 Diff に出すべき内容」を正しく導く。
func TestCurrentDiffReqByFocus(t *testing.T) {
	m := New()
	m.log.log = &jjlog.Log{Rows: []jjlog.Row{{ChangeID: "aaaa", IsWC: true, Lines: []string{"x"}}}}
	m.log.sel = 0
	m.files.items = []jjdiff.FileChange{{Status: "M", Path: "src/x.go"}}
	m.files.sel = 0

	m.focus = paneLog
	if got := m.currentDiffReq(); got != (diffReq{change: "aaaa"}) {
		t.Errorf("Log focus: got %+v, want whole-change {aaaa}", got)
	}

	m.focus = paneFiles
	if got := m.currentDiffReq(); got != (diffReq{change: "aaaa", file: "src/x.go"}) {
		t.Errorf("Files focus: got %+v, want file-level {aaaa, src/x.go}", got)
	}
}

// 非同期 diff 結果の stale ガード。
func TestStaleDiffResultIgnored(t *testing.T) {
	m := New()
	m.log.log = &jjlog.Log{Rows: []jjlog.Row{{ChangeID: "aaaa", Lines: []string{"x"}}}}
	m.log.sel = 0
	m.focus = paneLog

	stale, _ := m.Update(diffLoadedMsg{req: diffReq{change: "bbbb"}, raw: []byte("STALE")})
	if got := stale.(Model).diff.content; got != "" {
		t.Errorf("stale result applied: %q", got)
	}

	fresh, _ := m.Update(diffLoadedMsg{req: diffReq{change: "aaaa"}, raw: []byte("FRESH")})
	if got := fresh.(Model).diff.content; got != "FRESH" {
		t.Errorf("matching result not applied: %q", got)
	}
}

// filesLoadedMsg: change 一致なら files を入れて fileSel リセット。
func TestFilesLoadedSetsFilesAndResetsSelection(t *testing.T) {
	m := New()
	m.log.log = &jjlog.Log{Rows: []jjlog.Row{{ChangeID: "aaaa", Lines: []string{"x"}}}}
	m.log.sel = 0
	m.files.sel = 3

	files := []jjdiff.FileChange{{Status: "M", Path: "a"}, {Status: "A", Path: "b"}}
	updated, _ := m.Update(filesLoadedMsg{change: "aaaa", files: files})
	got := updated.(Model)
	if len(got.files.items) != 2 {
		t.Fatalf("files len = %d, want 2", len(got.files.items))
	}
	if got.files.sel != 0 {
		t.Errorf("fileSel = %d, want 0 (reset)", got.files.sel)
	}

	stale, _ := m.Update(filesLoadedMsg{change: "zzzz", files: files})
	if len(stale.(Model).files.items) != 0 {
		t.Errorf("stale files applied: %v", stale.(Model).files.items)
	}
}

// Tab が focusCycle（Log → Files → Bookmarks → Log）を巡回。
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

// 成功時: opInFlight false、refresh Cmd 非 nil。
func TestOpResultSuccessClearsFlightAndReturnsRefreshCmd(t *testing.T) {
	m := New()
	m.log.log = &jjlog.Log{Rows: []jjlog.Row{{ChangeID: "a", Lines: []string{"x"}}}}
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

// 失敗時: opInFlight false、err セット、refresh は走らせない。
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
func keyPress(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: r, Text: string(r)}
}

// 'n' 押下: opInFlight=true、Cmd 非 nil。
func TestKeyNewFiresOpAndSetsFlight(t *testing.T) {
	m := New()
	m.log.log = &jjlog.Log{Rows: []jjlog.Row{{ChangeID: "abc", IsWC: true, Lines: []string{"x"}}}}
	m.log.sel = 0
	updated, cmd := m.Update(keyPress('n'))
	got := updated.(Model)
	if !got.opInFlight {
		t.Errorf("opInFlight = false, want true")
	}
	if cmd == nil {
		t.Errorf("cmd = nil, want non-nil jj new Cmd")
	}
}

// opInFlight 中の 'n' は無視。
func TestKeyNewIgnoredWhenOpInFlight(t *testing.T) {
	m := New()
	m.log.log = &jjlog.Log{Rows: []jjlog.Row{{ChangeID: "abc", IsWC: true, Lines: []string{"x"}}}}
	m.log.sel = 0
	m.opInFlight = true
	updated, cmd := m.Update(keyPress('n'))
	if cmd != nil {
		t.Errorf("cmd = %v, want nil (gated by opInFlight)", cmd)
	}
	if !updated.(Model).opInFlight {
		t.Errorf("opInFlight = false, want true (unchanged)")
	}
}

// 選択 change 無しの 'n' は no-op。
func TestKeyNewIgnoredWhenNoSelection(t *testing.T) {
	m := New()
	updated, cmd := m.Update(keyPress('n'))
	if cmd != nil {
		t.Errorf("cmd = %v, want nil (no selection)", cmd)
	}
	if updated.(Model).opInFlight {
		t.Errorf("opInFlight = true, want false (no op fired)")
	}
}

// 'e' 押下: opInFlight=true、Cmd 非 nil。
func TestKeyEditFiresOpAndSetsFlight(t *testing.T) {
	m := New()
	m.log.log = &jjlog.Log{Rows: []jjlog.Row{{ChangeID: "abc", IsWC: true, Lines: []string{"x"}}}}
	m.log.sel = 0
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
	m.log.log = &jjlog.Log{Rows: []jjlog.Row{{ChangeID: "abc", IsWC: true, Lines: []string{"x"}}}}
	m.log.sel = 0
	m.opInFlight = true
	updated, cmd := m.Update(keyPress('e'))
	if cmd != nil {
		t.Errorf("cmd = %v, want nil (gated)", cmd)
	}
	if !updated.(Model).opInFlight {
		t.Errorf("opInFlight changed, want unchanged")
	}
}

// ---------------------------------------------------------------------------
// Modal seam: abandon flow
// ---------------------------------------------------------------------------

// 'a' 押下: modal=abandonModal{target}、Cmd は nil（まだ何も発射しない）。
func TestKeyAbandonEntersConfirming(t *testing.T) {
	m := New()
	m.log.log = &jjlog.Log{Rows: []jjlog.Row{{ChangeID: "abc", IsWC: true, Lines: []string{"x"}}}}
	m.log.sel = 0
	updated, cmd := m.Update(keyPress('a'))
	got := updated.(Model)
	am, ok := got.modal.(abandonModal)
	if !ok {
		t.Fatalf("modal = %T, want abandonModal", got.modal)
	}
	if am.target != "abc" {
		t.Errorf("modal.target = %q, want %q", am.target, "abc")
	}
	if cmd != nil {
		t.Errorf("cmd = %v, want nil (no op yet, just modal)", cmd)
	}
}

// abandon modal 中の Enter: jj abandon 発射、modal 閉じる、opInFlight=true。
func TestConfirmYesFiresAbandonCmd(t *testing.T) {
	m := New()
	m.log.log = &jjlog.Log{Rows: []jjlog.Row{{ChangeID: "abc", IsWC: true, Lines: []string{"x"}}}}
	m.log.sel = 0
	m.modal = abandonModal{target: "abc"}
	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	got := updated.(Model)
	if got.modal != nil {
		t.Errorf("modal = %T, want nil (back to normal after confirm)", got.modal)
	}
	if !got.opInFlight {
		t.Errorf("opInFlight = false, want true (op fired)")
	}
	if cmd == nil {
		t.Errorf("cmd = nil, want non-nil jj abandon Cmd")
	}
}

// abandon modal 中の Esc: modal 閉じる、opInFlight 変化なし、Cmd nil。
func TestConfirmNoReturnsToNormal(t *testing.T) {
	m := New()
	m.log.log = &jjlog.Log{Rows: []jjlog.Row{{ChangeID: "abc", IsWC: true, Lines: []string{"x"}}}}
	m.log.sel = 0
	m.modal = abandonModal{target: "abc"}
	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	got := updated.(Model)
	if got.modal != nil {
		t.Errorf("modal = %T, want nil (Esc cancels)", got.modal)
	}
	if got.opInFlight {
		t.Errorf("opInFlight = true, want false (cancel doesn't trigger op)")
	}
	if cmd != nil {
		t.Errorf("cmd = %v, want nil", cmd)
	}
}

// abandon modal 中の j (Down): logSel は動かさない。modal がナビキーを吸う。
func TestNormalKeysIgnoredInConfirmingMode(t *testing.T) {
	m := New()
	m.log.log = &jjlog.Log{Rows: []jjlog.Row{
		{ChangeID: "a", Lines: []string{"x"}},
		{ChangeID: "b", IsWC: true, Lines: []string{"y"}},
	}}
	m.log.sel = 0
	m.modal = abandonModal{target: "a"}
	updated, _ := m.Update(keyPress('j'))
	if got := updated.(Model).log.sel; got != 0 {
		t.Errorf("logSel = %d, want 0 (j ignored in confirming mode)", got)
	}
}

// abandon modal 中の footer はヒントを出す。
func TestFooterShowsConfirmHint(t *testing.T) {
	m := New()
	m.modal = abandonModal{target: "abc"}
	got := m.footerView()
	if !strings.Contains(got, "abandon") {
		t.Errorf("footer = %q, want to contain 'abandon'", got)
	}
}

// ---------------------------------------------------------------------------
// Modal seam: describe flow
// ---------------------------------------------------------------------------

// 'd' 押下: modal=describeModal{target, loading: true}、Cmd 非 nil (descSeedCmd)。
func TestKeyDescribeEntersLoadingAndFiresDescCmd(t *testing.T) {
	m := New()
	m.log.log = &jjlog.Log{Rows: []jjlog.Row{{ChangeID: "abc", IsWC: true, Lines: []string{"x"}}}}
	m.log.sel = 0
	updated, cmd := m.Update(keyPress('d'))
	got := updated.(Model)
	dm, ok := got.modal.(describeModal)
	if !ok {
		t.Fatalf("modal = %T, want describeModal", got.modal)
	}
	if !dm.loading {
		t.Errorf("loading = false, want true (just entered)")
	}
	if dm.target != "abc" {
		t.Errorf("target = %q, want %q", dm.target, "abc")
	}
	if cmd == nil {
		t.Errorf("cmd = nil, want descSeedCmd")
	}
}

// descLoadedMsg: target が現在の selectedChangeID と違えば stale で捨てる。
func TestDescLoadedStaleByTargetIgnored(t *testing.T) {
	m := New()
	m.log.log = &jjlog.Log{Rows: []jjlog.Row{{ChangeID: "abc", IsWC: true, Lines: []string{"x"}}}}
	m.log.sel = 0
	m.modal = newDescribeModal("abc", 80)
	updated, _ := m.Update(descLoadedMsg{target: "stale", desc: "old", err: nil})
	got := updated.(Model)
	dm, ok := got.modal.(describeModal)
	if !ok {
		t.Fatalf("modal = %T, want describeModal (stale ignored)", got.modal)
	}
	if !dm.loading {
		t.Errorf("loading = false, want true (stale should not advance state)")
	}
}

// describe modal が無い (loading 中に Esc 済み) なら descLoadedMsg は捨てる。
func TestDescLoadedStaleByModeIgnored(t *testing.T) {
	m := New()
	m.log.log = &jjlog.Log{Rows: []jjlog.Row{{ChangeID: "abc", IsWC: true, Lines: []string{"x"}}}}
	m.log.sel = 0
	// modal == nil（既にキャンセル済み）
	updated, _ := m.Update(descLoadedMsg{target: "abc", desc: "x", err: nil})
	got := updated.(Model)
	if got.modal != nil {
		t.Errorf("modal = %T, want nil (kept after stale msg)", got.modal)
	}
}

// descLoadedMsg(err): modal 閉じる + err セット。
func TestDescLoadedErrorReturnsToNormal(t *testing.T) {
	m := New()
	m.log.log = &jjlog.Log{Rows: []jjlog.Row{{ChangeID: "abc", IsWC: true, Lines: []string{"x"}}}}
	m.log.sel = 0
	m.modal = newDescribeModal("abc", 80)
	wantErr := errors.New("not found")
	updated, _ := m.Update(descLoadedMsg{target: "abc", err: wantErr})
	got := updated.(Model)
	if got.modal != nil {
		t.Errorf("modal = %T, want nil (err recovers)", got.modal)
	}
	if got.err != wantErr {
		t.Errorf("err = %v, want %v", got.err, wantErr)
	}
}

// describing loading 中の Esc: modal 閉じる。
func TestKeyEscCancelsDescribingLoading(t *testing.T) {
	m := New()
	m.log.log = &jjlog.Log{Rows: []jjlog.Row{{ChangeID: "abc", IsWC: true, Lines: []string{"x"}}}}
	m.log.sel = 0
	m.modal = newDescribeModal("abc", 80)
	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	got := updated.(Model)
	if got.modal != nil {
		t.Errorf("modal = %T, want nil", got.modal)
	}
}

// descLoadedMsg(成功): loading=false へ遷移、input に seed セット、Cmd は Focus blink。
func TestDescLoadedTransitionsToDescribingWithSeed(t *testing.T) {
	m := New()
	m.log.log = &jjlog.Log{Rows: []jjlog.Row{{ChangeID: "abc", IsWC: true, Lines: []string{"x"}}}}
	m.log.sel = 0
	m.modal = newDescribeModal("abc", 80)
	updated, cmd := m.Update(descLoadedMsg{target: "abc", desc: "feat: hello"})
	got := updated.(Model)
	dm, ok := got.modal.(describeModal)
	if !ok {
		t.Fatalf("modal = %T, want describeModal", got.modal)
	}
	if dm.loading {
		t.Errorf("loading = true, want false (advanced to editing)")
	}
	if dm.input.Value() != "feat: hello" {
		t.Errorf("input.Value() = %q, want %q", dm.input.Value(), "feat: hello")
	}
	if cmd == nil {
		t.Errorf("cmd = nil, want non-nil (Focus blink)")
	}
}

// describing 中の Enter: jj describe 発射、modal 閉じる、opInFlight=true。
func TestDescribingSubmitFiresDescribeCmd(t *testing.T) {
	m := New()
	m.log.log = &jjlog.Log{Rows: []jjlog.Row{{ChangeID: "abc", IsWC: true, Lines: []string{"x"}}}}
	m.log.sel = 0
	input := textinput.New()
	input.SetValue("new description")
	m.modal = describeModal{target: "abc", loading: false, input: input}
	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	got := updated.(Model)
	if got.modal != nil {
		t.Errorf("modal = %T, want nil", got.modal)
	}
	if !got.opInFlight {
		t.Errorf("opInFlight = false, want true")
	}
	if cmd == nil {
		t.Errorf("cmd = nil, want jj describe Cmd")
	}
}

// describing 中の Esc: modal 閉じる、opInFlight 変化なし。
func TestDescribingEscCancels(t *testing.T) {
	m := New()
	m.log.log = &jjlog.Log{Rows: []jjlog.Row{{ChangeID: "abc", IsWC: true, Lines: []string{"x"}}}}
	m.log.sel = 0
	input := textinput.New()
	input.SetValue("draft")
	m.modal = describeModal{target: "abc", loading: false, input: input}
	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	got := updated.(Model)
	if got.modal != nil {
		t.Errorf("modal = %T, want nil", got.modal)
	}
	if got.opInFlight {
		t.Errorf("opInFlight = true, want false (cancel)")
	}
}

// describing 中の 'j' は textinput に流れる → 値に追加される。
func TestDescribingNormalKeysGoToInput(t *testing.T) {
	m := New()
	m.log.log = &jjlog.Log{Rows: []jjlog.Row{{ChangeID: "abc", IsWC: true, Lines: []string{"x"}}}}
	m.log.sel = 0
	input := textinput.New()
	input.Focus()
	m.modal = describeModal{target: "abc", loading: false, input: input}
	updated, _ := m.Update(keyPress('j'))
	got := updated.(Model)
	dm, ok := got.modal.(describeModal)
	if !ok {
		t.Fatalf("modal = %T, want describeModal", got.modal)
	}
	if dm.input.Value() != "j" {
		t.Errorf("input.Value() = %q, want %q (j typed into input)", dm.input.Value(), "j")
	}
}

// WindowSize 後に describe modal を開くと、input が現在幅で正しくサイズされる。
// modal box = min(60, m.width-4)、padding(1, 2) で左右 4 引いた内寸が input 幅。
func TestDescribeModalSizesInputFromCurrentWidth(t *testing.T) {
	m := New()
	m.log.log = &jjlog.Log{Rows: []jjlog.Row{{ChangeID: "abc", IsWC: true, Lines: []string{"x"}}}}
	m.log.sel = 0
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = updated.(Model)
	updated, _ = m.Update(keyPress('d'))
	m = updated.(Model)
	dm, ok := m.modal.(describeModal)
	if !ok {
		t.Fatalf("modal = %T, want describeModal", m.modal)
	}
	// boxW = min(60, 100-4) = 60, padding 2*2 = 4, input 幅 = 56。
	if w := dm.input.Width(); w != 56 {
		t.Errorf("input.Width() = %d, want 56", w)
	}
}

// 既に open している describe modal も WindowSize で resize される。
func TestWindowSizeResizesActiveDescribeModal(t *testing.T) {
	m := New()
	m.modal = newDescribeModal("abc", 40) // 初期幅 40 → input width 32
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	got := updated.(Model)
	dm, ok := got.modal.(describeModal)
	if !ok {
		t.Fatalf("modal = %T, want describeModal", got.modal)
	}
	if w := dm.input.Width(); w != 56 {
		t.Errorf("input.Width() = %d, want 56 after resize", w)
	}
}

// Issue #4: bookmarksLoadedMsg を流すと Model.bookmarks に反映され、View に出ること。
func TestBookmarksLoadedMsgAppliesToPane(t *testing.T) {
	m := New()
	m.width, m.height = 120, 40
	m.ready = true
	m.applyLayout()
	updated, _ := m.Update(bookmarksLoadedMsg{items: []jjbookmark.Bookmark{
		{Name: "feature", LocalTarget: &jjbookmark.Target{ChangeID: "kqlu1234"}},
	}})
	got := updated.(Model)
	if got.bookmarks.SelectedName() != "feature" {
		t.Errorf("bookmarks.SelectedName = %q, want %q", got.bookmarks.SelectedName(), "feature")
	}
	if !strings.Contains(got.bodyView(), "feature → kqlu1234") {
		t.Errorf("bodyView missing bookmark line\n%s", got.bodyView())
	}
}

// Issue #4: focus が paneBookmarks のとき j/k が bookmarks の選択を動かし、log は動かさないこと。
func TestMoveSelectionRoutesToBookmarks(t *testing.T) {
	m := New()
	m.log.log = twoRowLog()
	m.bookmarks.Apply([]jjbookmark.Bookmark{
		{Name: "a", LocalTarget: &jjbookmark.Target{ChangeID: "11111111"}},
		{Name: "b", LocalTarget: &jjbookmark.Target{ChangeID: "22222222"}},
	})
	m.log.sel = 0
	m.focus = paneBookmarks
	m = moveSelection(m, +1)
	if got := m.bookmarks.SelectedName(); got != "b" {
		t.Errorf("Bookmarks focus +1: bookmarks.SelectedName = %q, want %q", got, "b")
	}
	if m.log.sel != 0 {
		t.Errorf("Bookmarks focus +1: log.sel = %d, want 0 (log untouched)", m.log.sel)
	}
}

// Issue #4: '.' で focus==paneBookmarks のとき opInFlight が立ち、Cmd が返ること。
func TestAdvanceKeyFiresOpAndSetsInFlight(t *testing.T) {
	m := New()
	m.bookmarks.Apply([]jjbookmark.Bookmark{
		{Name: "feature", LocalTarget: &jjbookmark.Target{ChangeID: "kqlu1234"}},
	})
	m.focus = paneBookmarks
	updated, cmd := m.Update(keyPress('.'))
	got := updated.(Model)
	if !got.opInFlight {
		t.Error("opInFlight = false, want true after advance key")
	}
	if cmd == nil {
		t.Error("cmd = nil, want non-nil Cmd that fires jj bookmark move")
	}
}

// focus が paneLog のとき '.' は no-op。
func TestAdvanceKeyNoOpOutsideBookmarks(t *testing.T) {
	m := New()
	m.focus = paneLog
	updated, cmd := m.Update(keyPress('.'))
	got := updated.(Model)
	if got.opInFlight {
		t.Error("opInFlight = true, want false (advance only fires from bookmarks pane)")
	}
	if cmd != nil {
		t.Error("cmd != nil, want nil")
	}
}

// 選択 bookmark が空（items 空）なら no-op。
func TestAdvanceKeyNoOpEmpty(t *testing.T) {
	m := New()
	m.focus = paneBookmarks
	updated, cmd := m.Update(keyPress('.'))
	got := updated.(Model)
	if got.opInFlight {
		t.Error("opInFlight = true on empty bookmarks")
	}
	if cmd != nil {
		t.Error("cmd != nil on empty bookmarks")
	}
}
