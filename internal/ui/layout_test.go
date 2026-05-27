package ui

import "testing"

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
