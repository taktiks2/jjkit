package jjlog

import "testing"

// 合成 fixture: change を2つ。先頭行にセンチネル "%%JJK%%<id>|<wcフラグ>%%JJK%%"、
// 継続行 "│" にはマーカー無し。\x1b[..m は jj の色付けを模した ANSI。
//
//	change aaaa1111 … 作業コピー(@, フラグ 1)、2行（先頭 + desc-a-line2）
//	change bbbb2222 … 作業コピーでない(フラグ 0)、1行
func fixture() []byte {
	s := "\x1b[1m◉\x1b[0m " + Sentinel + "aaaa1111|1" + Sentinel + " main desc-a-line1\n" +
		"│  desc-a-line2\n" +
		"◉ " + Sentinel + "bbbb2222|0" + Sentinel + " desc-b-line1\n"
	return []byte(s)
}

// ★AC「画面行→change マッピング」本体。センチネル行で change を束ね（マーカー無しの
// 継続行は前の change にぶら下がる）、change-id と作業コピーフラグを抽出できるかを検証する。
// 複数行ノード全体をハイライトできるのは、この束ねが正しいことが前提。
func TestParseMapsLinesToChanges(t *testing.T) {
	log, err := Parse(fixture())
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(log.Rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(log.Rows))
	}
	if log.Rows[0].ChangeID != "aaaa1111" || !log.Rows[0].IsWC {
		t.Errorf("row0 = %+v, want change aaaa1111 IsWC", log.Rows[0])
	}
	if len(log.Rows[0].Lines) != 2 {
		t.Errorf("row0 lines = %d, want 2 (node + continuation)", len(log.Rows[0].Lines))
	}
	if log.Rows[1].ChangeID != "bbbb2222" || log.Rows[1].IsWC {
		t.Errorf("row1 = %+v, want change bbbb2222 not IsWC", log.Rows[1])
	}
}

// 機能を支える2つの計算を固定する。WorkingCopyRow は「r でカーソルを @ にリセット」に、
// LineRange は「選択 change の全行をハイライト」に必須。範囲の境界がズレると隣の change まで
// 塗ってしまうため、flat 行範囲を明示的に検証する。
func TestLineRangeAndWorkingCopy(t *testing.T) {
	log, _ := Parse(fixture())
	if got := log.WorkingCopyRow(); got != 0 {
		t.Errorf("WorkingCopyRow = %d, want 0", got)
	}
	if s, e := log.LineRange(0); s != 0 || e != 2 {
		t.Errorf("LineRange(0) = (%d,%d), want (0,2)", s, e)
	}
	if s, e := log.LineRange(1); s != 2 || e != 3 {
		t.Errorf("LineRange(1) = (%d,%d), want (2,3)", s, e)
	}
}

// @ が先頭でない場合に WorkingCopyRow が正しく動くか。jj では jj edit で @ が log の
// 途中に来るため実際に起きる。スタブ実装(return 0)を許さない＝以前踏んだ「偶然パス」バグを防ぐ。
func TestWorkingCopyRowNotFirst(t *testing.T) {
	raw := []byte(
		"◆  " + Sentinel + "aaaa1111|0" + Sentinel + " main\n" +
			"@  " + Sentinel + "bbbb2222|1" + Sentinel + " wip\n")
	log, _ := Parse(raw)
	if got := log.WorkingCopyRow(); got != 1 {
		t.Errorf("WorkingCopyRow = %d, want 1 (@ is the second row)", got)
	}
}

// 内部メタdata（センチネル）が表示用 Lines に漏れていないことを保証する。
// ユーザーの画面に "%%JJK%%..." が出たら台無しなので、剥がし忘れを防ぐガード。
func TestParseStripsSentinelFromDisplay(t *testing.T) {
	log, _ := Parse(fixture())
	for _, r := range log.Rows {
		for _, ln := range r.Lines {
			if containsSentinel(ln) {
				t.Errorf("display line still has sentinel: %q", ln)
			}
		}
	}
}

// 実 jj log は change_id.short(8) を色付きで出すので、センチネル間に ANSI が混じる。
// Parse は ChangeID を素のテキスト（色コード抜き）で保持しなければならない。
// さもないと `jj diff -r <id>` で「Failed to parse revset」になる（Task 9 で踏んだバグ）。
func TestParseStripsANSIFromChangeID(t *testing.T) {
	raw := []byte("◉  " + Sentinel + "\x1b[38;5;5mpnmvvnyx\x1b[39m|1" + Sentinel + " desc\n")
	log, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(log.Rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(log.Rows))
	}
	if log.Rows[0].ChangeID != "pnmvvnyx" {
		t.Errorf("ChangeID = %q, want %q (ANSI must be stripped)", log.Rows[0].ChangeID, "pnmvvnyx")
	}
	if !log.Rows[0].IsWC {
		t.Errorf("IsWC = false, want true")
	}
}

// containsSentinel はテスト補助: 文字列中にセンチネルが残っていないかを素朴に走査する。
func containsSentinel(s string) bool {
	for i := 0; i+len(Sentinel) <= len(s); i++ {
		if s[i:i+len(Sentinel)] == Sentinel {
			return true
		}
	}
	return false
}
