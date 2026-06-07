package jjop

import "testing"

// 合成 fixture: 2 つの operation。各エントリの先頭行にセンチネル
// "%%JJKOP%%<op-id>%%JJKOP%%"、継続行（"│" で始まる）にはマーカー無し。
// 実際の jj op log は op-id を色付きで出すため、テストでも ANSI を混ぜる。
//
//	op aaaa11112222 … 3 行（先頭 + description + args）
//	op bbbb33334444 … 2 行（先頭 + description）
func fixture() []byte {
	s := "\x1b[1m\x1b[38;5;2m@\x1b[0m  " + Sentinel + "\x1b[38;5;12maaaa11112222\x1b[39m" + Sentinel + " taktiks2@host 1 minute ago\n" +
		"│  new empty commit\n" +
		"│  args: jj new main\n" +
		"○  " + Sentinel + "bbbb33334444" + Sentinel + " taktiks2@host 2 minutes ago\n" +
		"│  describe commit\n"
	return []byte(s)
}

// ★AC「Oplog ペインは操作を識別できる粒度で並べる」本体。センチネル行で
// op を束ね（マーカー無しの継続行は前の op にぶら下がる）、op-id を取り出せるかを検証する。
func TestParseMapsLinesToOps(t *testing.T) {
	ol, err := Parse(fixture())
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(ol.Rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(ol.Rows))
	}
	if ol.Rows[0].ID != "aaaa11112222" {
		t.Errorf("row0.ID = %q, want aaaa11112222", ol.Rows[0].ID)
	}
	if len(ol.Rows[0].Lines) != 3 {
		t.Errorf("row0 lines = %d, want 3 (header + description + args)", len(ol.Rows[0].Lines))
	}
	if ol.Rows[1].ID != "bbbb33334444" {
		t.Errorf("row1.ID = %q, want bbbb33334444", ol.Rows[1].ID)
	}
	if len(ol.Rows[1].Lines) != 2 {
		t.Errorf("row1 lines = %d, want 2", len(ol.Rows[1].Lines))
	}
}

// 表示用 Lines にセンチネルが残らないこと（ユーザー画面に "%%JJKOP%%..." を出さない）。
func TestParseStripsSentinelFromDisplay(t *testing.T) {
	ol, _ := Parse(fixture())
	for _, r := range ol.Rows {
		for _, ln := range r.Lines {
			if containsSentinel(ln) {
				t.Errorf("display line still has sentinel: %q", ln)
			}
		}
	}
}

// op-id は実 jj 出力では ANSI 色コードで包まれるため、Parse は素のテキストに正規化する。
// さもないと `jj op restore <id>` で「Invalid operation ID」になる（jjlog の Task 9 と同種のバグ）。
func TestParseStripsANSIFromOpID(t *testing.T) {
	raw := []byte("@  " + Sentinel + "\x1b[38;5;5m423e4ff24640\x1b[39m" + Sentinel + " taktiks2@host now\n")
	ol, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(ol.Rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(ol.Rows))
	}
	if ol.Rows[0].ID != "423e4ff24640" {
		t.Errorf("ID = %q, want %q (ANSI must be stripped)", ol.Rows[0].ID, "423e4ff24640")
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
