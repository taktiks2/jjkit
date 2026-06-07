// Package jjbookmark は `jj bookmark list --all-remotes` のテンプレート出力を
// Name でグルーピングし、local / remote-tracking + divergence を持つ Bookmark に変換する。
package jjbookmark

import (
	"strconv"
	"strings"
)

// Template は `jj bookmark list --all-remotes -T` に渡す式。
// 1 行 = 1 ref（local もしくは remote）。タブ区切りで <name> <remote> <change-id> <ahead> <behind>。
//
// jj の tracking_*_count は「remote 視点」(tracking_ahead_count = remote が local より N ahead)。
// 我々の RemoteRef.Ahead/Behind は「local 視点」(local が remote より N ahead/behind) で扱いたいので、
// テンプレート段階で意図的に入れ替えて出力する:
//
//	<ahead> 列  ← tracking_behind_count (remote が local より遅れている = local が ahead)
//	<behind> 列 ← tracking_ahead_count  (remote が local より進んでいる = local が behind)
//
// これで Parse 以降は素直に Ahead = atoi(ahead), Behind = atoi(behind) と書ける。
//
// 削除済 ref / local 自身の ahead-behind は "-" を出す。
const Template = `self.name() ++ "\t" ++ self.remote() ++ "\t" ++ if(self.present(),
	self.normal_target().change_id().shortest(8), "-") ++ "\t" ++ if(self.remote(),
	self.tracking_behind_count().lower() ++ "\t" ++ self.tracking_ahead_count().lower(), "-\t-") ++ "\n"`

// Target は ref が指す change の短縮 ID（無ければ親 *Target が nil）。
type Target struct {
	ChangeID string
}

// RemoteRef は1つの remote-tracking 情報。Ahead/Behind は **local 視点**:
//   - Ahead  = local が remote より N commit 進んでいる (push 必要)
//   - Behind = local が remote より N commit 遅れている (pull 必要)
//
// jj 側は remote 視点で値を返すため、Template 段階で入れ替えて格納している (Template の docstring 参照)。
type RemoteRef struct {
	Remote string
	Target Target
	Ahead  int
	Behind int
}

// Bookmark は同じ Name で集約された local + remote-tracking。
type Bookmark struct {
	Name        string
	LocalTarget *Target
	Remotes     []RemoteRef
}

// Parse はテンプレート出力を行ごとに分解し、Name でグルーピングする。順序は最初に出てきた Name 順を保つ。
func Parse(raw []byte) []Bookmark {
	text := strings.TrimRight(strings.ReplaceAll(string(raw), "\r\n", "\n"), "\n")
	if text == "" {
		return nil
	}
	var out []Bookmark
	idx := map[string]int{}
	for _, line := range strings.Split(text, "\n") {
		fields := strings.Split(line, "\t")
		if len(fields) < 5 {
			continue
		}
		name, remote, change, ahead, behind := fields[0], fields[1], fields[2], fields[3], fields[4]
		i, ok := idx[name]
		if !ok {
			i = len(out)
			idx[name] = i
			out = append(out, Bookmark{Name: name})
		}
		if remote == "" {
			if change != "-" {
				out[i].LocalTarget = &Target{ChangeID: change}
			}
			continue
		}
		out[i].Remotes = append(out[i].Remotes, RemoteRef{
			Remote: remote,
			Target: Target{ChangeID: change},
			Ahead:  atoi(ahead),
			Behind: atoi(behind),
		})
	}
	return out
}

// atoi は "-" や壊れた値を 0 にして読み続ける（テンプレートの稀な空値を握りつぶす）。
func atoi(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}
