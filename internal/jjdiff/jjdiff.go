// Package jjdiff は `jj diff --summary` の出力を変更ファイル一覧へパースする。
// diff 本文自体は ADR-0002 のとおり生のまま表示するのでパースしない。
package jjdiff

import "strings"

// FileChange は1ファイルの変更。Status は "M"/"A"/"D" 等、Path はリポジトリ相対パス。
type FileChange struct {
	Status string
	Path   string
}

// ParseSummary は `M path` 形式の各行を FileChange に変換する。
// 最初の空白までを Status、残り（パス内の空白は保持）を Path とする。
func ParseSummary(raw []byte) []FileChange {
	text := strings.TrimRight(strings.ReplaceAll(string(raw), "\r\n", "\n"), "\n")
	if text == "" {
		return nil
	}
	var out []FileChange
	for _, line := range strings.Split(text, "\n") {
		if line == "" {
			continue
		}
		sp := strings.IndexByte(line, ' ')
		if sp < 0 {
			continue // 想定外の形は捨てる
		}
		out = append(out, FileChange{Status: line[:sp], Path: line[sp+1:]})
	}
	return out
}
