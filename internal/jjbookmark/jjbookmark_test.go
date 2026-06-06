package jjbookmark

import (
	"reflect"
	"testing"
)

// ローカルのみ・remote のみ・両方持ちが正しくグルーピングされること。
// テンプレートは local 視点に正規化済みで出力する (詳細は jjbookmark.go の Template 参照)。
// 4列目 = local ahead (local が remote より N 進んでいる), 5列目 = local behind (local が remote より N 遅れている)。
func TestParseGroupsByName(t *testing.T) {
	raw := []byte(
		"main\t\tlvwnovmp\t-\t-\n" +
			"main\tgit\tlvwnovmp\t0\t0\n" +
			// "main origin abcd1234 0 2" = origin は local より 2 commit 進んでいる (local 2 behind)
			"main\torigin\tabcd1234\t0\t2\n" +
			"feature\t\tkqlu5678\t-\t-\n",
	)
	got := Parse(raw)
	want := []Bookmark{
		{
			Name:        "main",
			LocalTarget: &Target{ChangeID: "lvwnovmp"},
			Remotes: []RemoteRef{
				{Remote: "git", Target: Target{ChangeID: "lvwnovmp"}, Ahead: 0, Behind: 0},
				{Remote: "origin", Target: Target{ChangeID: "abcd1234"}, Ahead: 0, Behind: 2},
			},
		},
		{
			Name:        "feature",
			LocalTarget: &Target{ChangeID: "kqlu5678"},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Parse = %#v, want %#v", got, want)
	}
}

// remote はあるがローカルが削除された bookmark は LocalTarget=nil で出ること。
// "deleted origin dead0001 3 0" = local が origin より 3 commit 進んでいる (local 3 ahead)。
func TestParseLocalAbsent(t *testing.T) {
	raw := []byte(
		"deleted\t\t-\t-\t-\n" +
			"deleted\torigin\tdead0001\t3\t0\n",
	)
	got := Parse(raw)
	want := []Bookmark{
		{
			Name:        "deleted",
			LocalTarget: nil,
			Remotes: []RemoteRef{
				{Remote: "origin", Target: Target{ChangeID: "dead0001"}, Ahead: 3, Behind: 0},
			},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Parse = %#v, want %#v", got, want)
	}
}

// 空入力で空スライス。
func TestParseEmpty(t *testing.T) {
	if got := Parse([]byte("")); len(got) != 0 {
		t.Errorf("Parse(empty) = %#v, want empty", got)
	}
}
