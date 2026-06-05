package jjbookmark

import (
	"reflect"
	"testing"
)

// ローカルのみ・remote のみ・両方持ちが正しくグルーピングされること。
func TestParseGroupsByName(t *testing.T) {
	raw := []byte(
		"main\t\tlvwnovmp\t-\t-\n" +
			"main\tgit\tlvwnovmp\t0\t0\n" +
			"main\torigin\tabcd1234\t2\t0\n" +
			"feature\t\tkqlu5678\t-\t-\n",
	)
	got := Parse(raw)
	want := []Bookmark{
		{
			Name:        "main",
			LocalTarget: &Target{ChangeID: "lvwnovmp"},
			Remotes: []RemoteRef{
				{Remote: "git", Target: Target{ChangeID: "lvwnovmp"}, Ahead: 0, Behind: 0},
				{Remote: "origin", Target: Target{ChangeID: "abcd1234"}, Ahead: 2, Behind: 0},
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
func TestParseLocalAbsent(t *testing.T) {
	raw := []byte(
		"deleted\t\t-\t-\t-\n" +
			"deleted\torigin\tdead0001\t0\t3\n",
	)
	got := Parse(raw)
	want := []Bookmark{
		{
			Name:        "deleted",
			LocalTarget: nil,
			Remotes: []RemoteRef{
				{Remote: "origin", Target: Target{ChangeID: "dead0001"}, Ahead: 0, Behind: 3},
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
