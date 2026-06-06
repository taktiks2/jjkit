package ui

import (
	"strings"
	"testing"

	"github.com/taktiks2/jjkit/internal/jjbookmark"
)

// 各 bookmark が1行で「name → change-id [marker]」形式で出ること。
// 同期済は ✓、 ahead/behind は ↑n↓m、ローカルのみは divergence セクション無し。
func TestBookmarksPaneRender(t *testing.T) {
	p := &bookmarksPane{}
	p.vp.SetWidth(60)
	p.vp.SetHeight(10)
	p.Apply([]jjbookmark.Bookmark{
		{
			Name:        "feature",
			LocalTarget: &jjbookmark.Target{ChangeID: "kqlu1234"},
			Remotes: []jjbookmark.RemoteRef{
				{Remote: "origin", Target: jjbookmark.Target{ChangeID: "abcd5678"}, Ahead: 2, Behind: 0},
			},
		},
		{
			Name:        "main",
			LocalTarget: &jjbookmark.Target{ChangeID: "c9d4e58b"},
			Remotes: []jjbookmark.RemoteRef{
				{Remote: "origin", Target: jjbookmark.Target{ChangeID: "c9d4e58b"}, Ahead: 0, Behind: 0},
			},
		},
		{
			Name:        "wip",
			LocalTarget: &jjbookmark.Target{ChangeID: "omykzzzz"},
		},
	})
	view := p.View()
	for _, want := range []string{
		"feature → kqlu1234 [↑2]",
		"main → c9d4e58b ✓",
		"wip → omykzzzz",
	} {
		if !strings.Contains(view, want) {
			t.Errorf("View missing %q\n---\n%s", want, view)
		}
	}
}

// 選択を Move で動かせること + SelectedName が選択名を返すこと。
func TestBookmarksPaneMove(t *testing.T) {
	p := &bookmarksPane{}
	p.vp.SetWidth(60)
	p.vp.SetHeight(10)
	p.Apply([]jjbookmark.Bookmark{
		{Name: "a", LocalTarget: &jjbookmark.Target{ChangeID: "11111111"}},
		{Name: "b", LocalTarget: &jjbookmark.Target{ChangeID: "22222222"}},
		{Name: "c", LocalTarget: &jjbookmark.Target{ChangeID: "33333333"}},
	})
	if got := p.SelectedName(); got != "a" {
		t.Errorf("SelectedName initial = %q, want %q", got, "a")
	}
	p.Move(2)
	if got := p.SelectedName(); got != "c" {
		t.Errorf("SelectedName after Move(2) = %q, want %q", got, "c")
	}
	p.Move(-1)
	if got := p.SelectedName(); got != "b" {
		t.Errorf("SelectedName after Move(-1) = %q, want %q", got, "b")
	}
	// 範囲外クランプ
	p.Move(100)
	if got := p.SelectedName(); got != "c" {
		t.Errorf("SelectedName after Move(100) = %q, want %q (clamped)", got, "c")
	}
}

// 空 bookmark で Move しても panic しないこと、SelectedName が "" を返すこと。
func TestBookmarksPaneEmpty(t *testing.T) {
	p := &bookmarksPane{}
	p.vp.SetWidth(60)
	p.vp.SetHeight(10)
	p.Move(1)
	if got := p.SelectedName(); got != "" {
		t.Errorf("SelectedName empty = %q, want \"\"", got)
	}
}
