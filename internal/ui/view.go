package ui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	// errStyle はエラー行のスタイル（赤文字）。
	errStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	paneStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("248"))
	paneFocused = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("5"))
	titleStyle  = lipgloss.NewStyle().Bold(true)
)

// lipglossHeight は文字列の行数（高さ）を返す。viewport の高さ計算に使う。
func lipglossHeight(s string) int { return lipgloss.Height(s) }

// View は5ペイン（左カラムに Log/Files/Bookmarks/Oplog、右に Diff）+ 下部フッターを返す。
// 起動直後（ready=false）は loading 表示。
func (m Model) View() tea.View {
	content := "loading..."
	if m.ready {
		l := computeLayout(m.width, m.height, lipglossHeight(m.footerView()))
		left := lipgloss.JoinVertical(lipgloss.Left,
			paneBox("Log", m.logVP.View(), l.log, m.focus == paneLog),
			paneBox("Files", m.filesVP.View(), l.files, m.focus == paneFiles),
			paneBox("Bookmarks", "(coming soon)", l.bookmarks, m.focus == paneBookmarks),
			paneBox("Oplog", "(coming soon)", l.oplog, false),
		)
		body := lipgloss.JoinHorizontal(lipgloss.Top, left, paneBox("Diff", m.diffVP.View(), l.diff, false))
		content = body + "\n" + m.footerView()
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

// footerView は通常は help の1行ヒント、エラー時は赤いエラー行を返す。
func (m Model) footerView() string {
	if m.err != nil {
		return errStyle.Render("error: " + m.err.Error())
	}
	return m.help.View(m.keys)
}
