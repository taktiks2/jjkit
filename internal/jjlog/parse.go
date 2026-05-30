// Package jjlog は `jj log` の出力（センチネルマーカー付きテンプレートで描画したもの）を
// change 単位の行にパースし、画面行を change へ対応付ける。
package jjlog

import (
	"slices"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// Sentinel はテンプレートが各 change に埋め込むマーカーを囲む文字列。色ラベルの付かない
// 素のテキストなので jj の --color always でも壊れず、表示前に剥がせる。
const Sentinel = "%%JJK%%"

// Template は jj log に渡す -T 引数。各 change の先頭行に Sentinel で change_id（短縮8桁）と
// 作業コピーフラグを埋め込み、後ろに jj 既定の compact ノードを続ける。
const Template = `"` + Sentinel + `" ++ change_id.short(8) ++ "|" ++ if(current_working_copy, "1", "0") ++ "` + Sentinel + `" ++ builtin_log_compact`

// Row は1つの change。id、作業コピー(@)かどうか、その change に属する全画面行を持つ。
type Row struct {
	ChangeID string
	IsWC     bool
	Lines    []string
}

// Log はパース結果。change 行を順序どおりに並べたもの。
type Log struct {
	Rows []Row
}

// Parse は jj log の生バイトを Row に分割する。センチネルを含む行が新しい Row を始め、
// 含まない行は現在の Row の継続行になる。
//
// 入力例（マーカー %%JJK%%<id>|<wc>%%JJK%% はテンプレートが各ノードの先頭行にだけ出すため、
// 継続行にはマーカーが付かない。実際の出力には ANSI 色コードも含まれ、Lines に保持される）:
//
//	◉  %%JJK%%aaaa1111|1%%JJK%% main desc-a-line1
//	│  desc-a-line2
//	◉  %%JJK%%bbbb2222|0%%JJK%% desc-b-line1
//
// は次のようにパースされる:
//
//	Rows[0] = {ChangeID:"aaaa1111", IsWC:true,  Lines:["◉  main desc-a-line1", "│  desc-a-line2"]}
//	Rows[1] = {ChangeID:"bbbb2222", IsWC:false, Lines:["◉  desc-b-line1"]}
func Parse(raw []byte) (*Log, error) {
	text := strings.ReplaceAll(string(raw), "\r\n", "\n")
	text = strings.TrimRight(text, "\n")
	log := &Log{}
	if text == "" {
		return log, nil
	}
	for _, line := range strings.Split(text, "\n") {
		// マーカーが無い行は継続行。直前の change にぶら下げて次の行へ。
		idx := strings.Index(line, Sentinel)
		if idx < 0 {
			appendContinuating(log, line)
			continue
		}

		// 2個目のマーカーを探す。見つからない壊れた行は保険として継続行扱い。
		rest := line[idx+len(Sentinel):]
		end := strings.Index(rest, Sentinel)
		if end < 0 {
			appendContinuating(log, line)
			continue
		}

		// マーカー行は新しい change の先頭。2個のマーカーに挟まれた payload から
		// change-id と作業コピーフラグを取り出し、マーカーを除いた表示行で新しい Row を作る。
		// payload は色コードを剥がす（実 jj は change_id.short(8) を色付きで出すため）。
		payload := ansi.Strip(rest[:end])
		display := line[:idx] + rest[end+len(Sentinel):]
		changeID, wc := payload, false
		if bar := strings.IndexByte(payload, '|'); bar >= 0 {
			changeID = payload[:bar]
			wc = payload[bar+1:] == "1"
		}
		log.Rows = append(log.Rows, Row{ChangeID: changeID, IsWC: wc, Lines: []string{display}})
	}
	return log, nil
}

// Lines は全 Row の画面行を順番に1本へ平坦化する（viewport へ流す用）。
func (l *Log) Lines() []string {
	var out []string
	for _, r := range l.Rows {
		out = append(out, r.Lines...)
	}
	return out
}

// WorkingCopyRow は @ の Row の index を返す。@ が無ければ 0。
func (l *Log) WorkingCopyRow() int {
	if i := slices.IndexFunc(l.Rows, func(r Row) bool { return r.IsWC }); i >= 0 {
		return i
	}
	return 0 // 空 revset 等で @ が無いときの意図的な既定値（go-style.md ルール1の例外）
}

// LineRange は row が占める flat 行の範囲 [start, end) を返す。
func (l *Log) LineRange(row int) (start, end int) {
	for i := range min(row, len(l.Rows)) {
		start += len(l.Rows[i].Lines)
	}
	end = start
	if row >= 0 && row < len(l.Rows) {
		end = start + len(l.Rows[row].Lines)
	}
	return start, end
}

// appendContinuating はマーカー無しの行を現在（最後）の Row に足す。
// 最初のマーカーより前の行は捨てる（jj log はノードから始まるため）。
func appendContinuating(log *Log, line string) {
	if n := len(log.Rows); n > 0 {
		log.Rows[n-1].Lines = append(log.Rows[n-1].Lines, line)
	}
}
