package ui

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

const (
	reset = "\x1b[0m"        // ANSI リセット: 色・装飾をすべて初期状態へ戻す
	selBG = "\x1b[48;5;236m" // 選択行の背景色: 48;5;N は「背景を 256色パレットの色 N に」（236=暗いグレー）
)

// highlightLine は1行を選択用の背景色で塗り、幅まで切り詰め／パディングする。
// 行中の reset(\x1b[0m) は背景も消すので、reset の直後に背景 SGR を再注入する（lipgloss #144 回避）。
func highlightLine(line string, width int) string {
	// 表示幅が pane より広い行は ANSI を壊さず切り詰める（第3引数 "" は省略記号を付けない指定）。
	if ansi.StringWidth(line) > width {
		line = ansi.Truncate(line, width, "")
	}
	// 行中の reset をすべて「reset + 背景ON」に置換し、reset で消えた背景を直後に貼り直す。
	body := strings.ReplaceAll(line, reset, reset+selBG)
	// 行末まで帯にするための空白の数。len() ではなく表示幅で計算する（負なら 0）。
	pad := width - ansi.StringWidth(line)
	if pad < 0 {
		pad = 0
	}
	// 先頭で背景ON → 本体 → パディング → 最後に reset で背景OFF、で行全体が1本の帯になる。
	return selBG + body + strings.Repeat(" ", pad) + reset
}

// RenderContent は viewport へ流す表示テキストを組み立てる。
// [selStart, selEnd) の行だけ背景を塗り、それ以外は元のまま通す。
func RenderContent(lines []string, selStart, selEnd, width int) string {
	out := make([]string, len(lines)) // 結果用スライスを行数ぶん確保
	for i, line := range lines {
		if i >= selStart && i < selEnd {
			out[i] = highlightLine(line, width) // 選択範囲の行 → 背景を塗る
		} else {
			out[i] = line // 範囲外の行 → 元のまま（壊さない）
		}
	}
	return strings.Join(out, "\n") // 改行で連結して viewport に渡せる1本の文字列へ
}
