package ui

import (
	"strings"
	"testing"

	"github.com/taktiks2/jjkit/internal/jjop"
)

// テスト用の 2 op log。row0=1行、row1=2行。
func twoRowOpLog() *jjop.OpLog {
	return &jjop.OpLog{Rows: []jjop.Row{
		{ID: "aaaa11112222", Lines: []string{"@  aaaa11112222 1 minute ago"}},
		{ID: "bbbb33334444", Lines: []string{"○  bbbb33334444 2 minutes ago", "│  describe commit"}},
	}}
}

// op log が pane に流れ込み、画面文字列に op-id（短縮）と表示行が出ること。
func TestOplogPaneRender(t *testing.T) {
	p := &oplogPane{}
	p.vp.SetWidth(80)
	p.vp.SetHeight(10)
	p.Apply(twoRowOpLog())
	view := p.View()
	for _, want := range []string{
		"aaaa11112222",
		"bbbb33334444",
		"describe commit",
	} {
		if !strings.Contains(view, want) {
			t.Errorf("View missing %q\n---\n%s", want, view)
		}
	}
}

// SelectedID は選択中 op の id を返す。Move で動かせて、範囲外はクランプ。
func TestOplogPaneMoveAndSelected(t *testing.T) {
	p := &oplogPane{}
	p.vp.SetWidth(80)
	p.vp.SetHeight(10)
	p.Apply(twoRowOpLog())
	if got := p.SelectedID(); got != "aaaa11112222" {
		t.Errorf("SelectedID initial = %q, want aaaa11112222", got)
	}
	p.Move(+1)
	if got := p.SelectedID(); got != "bbbb33334444" {
		t.Errorf("SelectedID after Move(+1) = %q, want bbbb33334444", got)
	}
	p.Move(+100)
	if got := p.SelectedID(); got != "bbbb33334444" {
		t.Errorf("SelectedID after Move(+100) = %q, want bbbb33334444 (clamped)", got)
	}
	p.Move(-100)
	if got := p.SelectedID(); got != "aaaa11112222" {
		t.Errorf("SelectedID after Move(-100) = %q, want aaaa11112222 (clamped)", got)
	}
}

// 空 oplog で Move しても panic しないこと、SelectedID は "" を返すこと。
func TestOplogPaneEmpty(t *testing.T) {
	p := &oplogPane{}
	p.vp.SetWidth(80)
	p.vp.SetHeight(10)
	p.Move(+1)
	if got := p.SelectedID(); got != "" {
		t.Errorf("SelectedID empty = %q, want \"\"", got)
	}
}
