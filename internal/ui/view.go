package ui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	// errStyle はエラー行のスタイル（赤文字）。
	errStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	paneStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("248"))
	paneFocused = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("5"))
	titleStyle  = lipgloss.NewStyle().Bold(true)

	modalBoxStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("5")).Padding(1, 2)
	modalHintStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

// lipglossHeight は文字列の行数（高さ）を返す。viewport の高さ計算に使う。
func lipglossHeight(s string) int { return lipgloss.Height(s) }

// View は5ペイン + footer を返す。mode != normal なら modal を中央に重ねる。
// 起動直後（ready=false）は loading 表示。
func (m Model) View() tea.View {
	content := "loading..."
	if m.ready {
		body := m.bodyView()
		if m.mode == modeNormal {
			content = body
		} else {
			modalText := m.modalView()
			mx, my := centerOffset(m.width, m.height, modalText)
			base := lipgloss.NewLayer(body)
			modal := lipgloss.NewLayer(modalText).X(mx).Y(my).Z(1)
			content = lipgloss.NewCompositor(base, modal).Render()
		}
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// paneBox は枠 + タイトル行 + 本体を外寸 r に収めて描く。フォーカス中は枠色を変える。
// 内寸 = 外寸 - 枠2 - タイトル1 を Width/Height に渡す。
func paneBox(title, body string, r rect, focused bool) string {
	style := paneStyle
	if focused {
		style = paneFocused
	}
	innerW := max(r.w-2, 0)
	content := titleStyle.Width(innerW).Render(title)
	if r.h-3 > 0 {
		content += "\n" + body
	}
	return style.Width(r.w).Height(r.h).Render(content)
}

// footerView は通常は help、modal 中はキーヒントを返す（赤帯エラーが最優先）。
func (m Model) footerView() string {
	if m.err != nil {
		return errStyle.Render("error: " + m.err.Error())
	}
	switch m.mode {
	case modeConfirmingAbandon:
		return modalHintStyle.Render("y/Enter: abandon . n/Esc: cancel")
	default:
		return m.help.View(m.keys)
	}
}

// bodyView は5ペイン + footer を組み立てて返す。modal 無しのフル画面。
func (m Model) bodyView() string {
	l := computeLayout(m.width, m.height, lipglossHeight(m.footerView()))
	left := lipgloss.JoinVertical(lipgloss.Left,
		paneBox("Log", m.logVP.View(), l.log, m.focus == paneLog),
		paneBox("Files", m.filesVP.View(), l.files, m.focus == paneFiles),
		paneBox("Bookmarks", "(coming soon)", l.bookmarks, m.focus == paneBookmarks),
		paneBox("Oplog", "(coming soon)", l.oplog, false),
	)
	body := lipgloss.JoinHorizontal(lipgloss.Top, left, paneBox("Diff", m.diffVP.View(), l.diff, false))
	return body + "\n" + m.footerView()
}

// modalView は現在の mode に応じた modal 文字列を返す。modeNormal では "" を返す。
func (m Model) modalView() string {
	switch m.mode {
	case modeConfirmingAbandon:
		boxW := min(60, m.width-4)
		title := "Abandon change?"
		body := fmt.Sprintf("Abandon %s?", m.selectedChangeID())
		hint := "y/Enter: abandon . n/Esc: cancel"
		return modalBoxStyle.Width(boxW).Render(title + "\n\n" + body + "\n\n" + hint)
	}
	return ""
}

// centerOffset は modal を画面中央に置くための左上座標を返す（純粋関数）。
func centerOffset(screenW, screenH int, modal string) (x, y int) {
	mw := lipgloss.Width(modal)
	mh := lipgloss.Height(modal)
	x = max(0, (screenW-mw)/2)
	y = max(0, (screenH-mh)/2)
	return
}
