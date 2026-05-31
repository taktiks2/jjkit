package ui

import "testing"

// applyLayout が logVP / filesVP / diffVP すべてに正しい内寸を渡すこと。
// バグ防止：3つの viewport のうち1つだけ更新して残りを 0×0 のまま放置する
// 実装ミス（受け入れ基準「リサイズで壊れない」を満たさない）を検出する。
// 内寸 = rect の外寸 - 枠2 - タイトル1。
func TestApplyLayoutSizesAllViewports(t *testing.T) {
	m := New()
	m.width, m.height = 100, 30
	m.applyLayout()

	footerH := lipglossHeight(m.footerView())
	want := computeLayout(100, 30, footerH)

	cases := []struct {
		name       string
		gotW, gotH int
		want       rect
	}{
		{"log", m.logVP.Width(), m.logVP.Height(), want.log},
		{"files", m.filesVP.Width(), m.filesVP.Height(), want.files},
		{"diff", m.diffVP.Width(), m.diffVP.Height(), want.diff},
	}
	for _, c := range cases {
		wantW := max(c.want.w-2, 0)
		wantH := max(c.want.h-3, 0)
		if c.gotW != wantW || c.gotH != wantH {
			t.Errorf("%s viewport: %dx%d, want %dx%d", c.name, c.gotW, c.gotH, wantW, wantH)
		}
	}
}

// 端末サイズが変わってもレイアウトが壊れないこと（issue #2 の受け入れ条件）。
// 契約: (1) どの矩形も負にならない、(2) 左カラム幅 + Diff 幅 = 全幅、
// (3) 左カラム4ペインの高さ合計 = body 高さ（footer を除いた高さ）、
// (4) Diff の高さは body と同じ。
//
// 比率の具体値（左カラム=全幅の2/5 など）は人間レビューで調整する余地があるので
// 不変条件だけを縛る。
func TestComputeLayoutReflows(t *testing.T) {
	for _, tc := range []struct{ w, h, footer int }{
		{100, 30, 1},
		{30, 8, 1},
		{200, 60, 2},
	} {
		l := computeLayout(tc.w, tc.h, tc.footer)
		body := tc.h - tc.footer

		for _, r := range []rect{l.log, l.files, l.bookmarks, l.oplog, l.diff} {
			if r.w < 0 || r.h < 0 {
				t.Fatalf("%dx%d: negative size %+v", tc.w, tc.h, r)
			}
		}
		if l.log.w+l.diff.w != tc.w {
			t.Errorf("%dx%d: left %d + diff %d != width %d", tc.w, tc.h, l.log.w, l.diff.w, tc.w)
		}
		if got := l.log.h + l.files.h + l.bookmarks.h + l.oplog.h; got != body {
			t.Errorf("%dx%d: left column heights sum %d, want body %d", tc.w, tc.h, got, body)
		}
		if l.diff.h != body {
			t.Errorf("%dx%d: diff.h %d, want body %d", tc.w, tc.h, l.diff.h, body)
		}
	}
}
