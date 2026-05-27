package ui

// rect は1区画の外寸（border を含む外側のサイズ）。
// viewport へ渡す内寸は描画時に枠分を引いて作る（Task 9）。
type rect struct {
	w, h int
}

// layout は端末を5ペインに割り付けた結果。
type layout struct {
	log, files, bookmarks, oplog, diff rect
}

func computeLayout(width, height, footerH int) layout {
	body := max(height-footerH, 0)

	leftW := min(width*2/5, width-20)
	leftW = max(leftW, 0)
	diffW := width - leftW

	bookmarksH, oplogH := 5, 5
	if body-bookmarksH-oplogH < 0 {
		bookmarksH, oplogH = body/4, body/4
	}
	rest := max(body-bookmarksH-oplogH, 0)
	logH := rest * 3 / 5
	filesH := rest - logH

	return layout{
		log:       rect{leftW, logH},
		files:     rect{leftW, filesH},
		bookmarks: rect{leftW, bookmarksH},
		oplog:     rect{leftW, oplogH},
		diff:      rect{diffW, body},
	}
}
