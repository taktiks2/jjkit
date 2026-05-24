package ui

import (
	"strings"
	"testing"
)

// 選択行が幅いっぱいに背景で塗られること（行末までパディングされること）を保証する。
// lazygit 風の「行全体ハイライト」の見た目の核。selBG/reset は実装側の定数を参照するので、
// 実際の色コードを変えてもこのテストは壊れない（記号で照合している）。
func TestHighlightLinePadsToWidth(t *testing.T) {
	got := highlightLine("abc", 5)
	want := selBG + "abc" + "  " + reset
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// ★Task 3 の技術リスク本体。行の途中に jj 由来の reset(\x1b[0m) があると、そこで背景色も
// 消えてしまう（lipgloss #144）。reset の直後に背景 SGR を再注入して、行末まで背景が
// 途切れないことを保証する。
func TestHighlightLineReassertsBgAfterInnerReset(t *testing.T) {
	got := highlightLine("a"+reset+"b", 2)
	if !strings.Contains(got, reset+selBG) {
		t.Errorf("bg not re-asserted after inner reset: %q", got)
	}
}

// 選択範囲 [selStart, selEnd) の行だけを塗り、範囲外の行は元のまま（壊さない）ことを保証する。
// 範囲がズレて隣の change まで塗る／選択行を塗り忘れる、を防ぐ。
func TestRenderContentHighlightsOnlySelectedRange(t *testing.T) {
	lines := []string{"l0", "l1", "l2"}
	got := RenderContent(lines, 1, 2, 4)
	parts := strings.Split(got, "\n")
	if parts[0] != "l0" {
		t.Errorf("line0 changed: %q", parts[0])
	}
	if !strings.HasPrefix(parts[1], selBG) {
		t.Errorf("line1 not highlighted: %q", parts[1])
	}
	if parts[2] != "l2" {
		t.Errorf("line2 changed: %q", parts[2])
	}
}
