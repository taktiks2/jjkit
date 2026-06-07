// Package jjop は `jj op log` の出力（センチネルマーカー付きテンプレートで描画したもの）を
// operation 単位の行にパースし、画面行を operation へ対応付ける。
//
// jjlog と同じ単発スキャン方式（ADR-0004）を踏襲する。op-log も複数行ノードを持つため、
// 位置依存の2パス突き合わせは脆い。
package jjop

import (
	"slices"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// Sentinel はテンプレートが各 operation に埋め込むマーカー。jjlog のものと衝突しない
// 別文字列にしておく。色ラベルの付かない素のテキストなので jj の --color always でも
// 壊れず、表示前に剥がせる。
const Sentinel = "%%JJKOP%%"

// Template は jj op log に渡す -T 引数。各 op の先頭行に Sentinel で op-id（短縮12桁）を
// 埋め込み、後ろに jj 既定の compact ノードを続ける。op-id は restore の引数になるため
// 12 桁で取る（ローカル衝突する確率が無視できる長さ）。
const Template = `"` + Sentinel + `" ++ id.short(12) ++ "` + Sentinel + `" ++ builtin_op_log_compact`

// Row は1つの operation。op-id とその op に属する全画面行を持つ。
type Row struct {
	ID    string
	Lines []string
}

// OpLog はパース結果。op 行を新しい順に並べたもの（jj op log の既定順序を維持）。
type OpLog struct {
	Rows []Row
}

// Parse は jj op log の生バイトを Row に分割する。センチネルを含む行が新しい Row を始め、
// 含まない行は現在の Row の継続行になる。
func Parse(raw []byte) (*OpLog, error) {
	text := strings.ReplaceAll(string(raw), "\r\n", "\n")
	text = strings.TrimRight(text, "\n")
	ol := &OpLog{}
	if text == "" {
		return ol, nil
	}
	for _, line := range strings.Split(text, "\n") {
		// マーカーが無い行は継続行。直前の op にぶら下げて次の行へ。
		idx := strings.Index(line, Sentinel)
		if idx < 0 {
			appendContinuating(ol, line)
			continue
		}

		// 2 個目のマーカーを探す。見つからない壊れた行は保険として継続行扱い。
		rest := line[idx+len(Sentinel):]
		end := strings.Index(rest, Sentinel)
		if end < 0 {
			appendContinuating(ol, line)
			continue
		}

		// マーカー行は新しい op の先頭。2 個のマーカーに挟まれた payload から
		// op-id を取り出し、マーカーを除いた表示行で新しい Row を作る。
		// payload は色コードを剥がす（実 jj は id.short(12) を色付きで出すため）。
		id := ansi.Strip(rest[:end])
		display := line[:idx] + rest[end+len(Sentinel):]
		ol.Rows = append(ol.Rows, Row{ID: id, Lines: []string{display}})
	}
	return ol, nil
}

// Lines は全 Row の画面行を順番に1本へ平坦化する（viewport へ流す用）。
func (l *OpLog) Lines() []string {
	var out []string
	for _, r := range l.Rows {
		out = append(out, r.Lines...)
	}
	return out
}

// LineRange は row が占める flat 行の範囲 [start, end) を返す。選択 op の全行を
// ハイライトするのに使う。
func (l *OpLog) LineRange(row int) (start, end int) {
	for i := range min(row, len(l.Rows)) {
		start += len(l.Rows[i].Lines)
	}
	end = start
	if row >= 0 && row < len(l.Rows) {
		end = start + len(l.Rows[row].Lines)
	}
	return start, end
}

// RowByID は ID 一致する Row の index を返す。無ければ -1。
// 用途: refresh 後に「前選択していた op ID」を新ログの中から探して選択を維持する。
func (l *OpLog) RowByID(id string) int {
	if id == "" {
		return -1
	}
	return slices.IndexFunc(l.Rows, func(r Row) bool { return r.ID == id })
}

// appendContinuating はマーカー無しの行を現在（最後）の Row に足す。
// 最初のマーカーより前の行は捨てる（jj op log はノードから始まるため）。
func appendContinuating(ol *OpLog, line string) {
	if n := len(ol.Rows); n > 0 {
		ol.Rows[n-1].Lines = append(ol.Rows[n-1].Lines, line)
	}
}
