package ui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// errStyle はエラー行のスタイル（赤文字）。
var errStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))

// lipglossHeight は文字列の行数（高さ）を返す。viewport の高さ計算に使う。
func lipglossHeight(s string) int { return lipgloss.Height(s) }

// View は viewport を上、フッターを下に並べて返す。v2 は string でなく tea.View を返す。
func (m Model) View() tea.View {
	if !m.ready {
		return tea.NewView("loading...")
	}
	return tea.NewView(m.viewport.View() + "\n" + m.footerView())
}

// footerView は通常は help の1行ヒント、エラー時は赤いエラー行を返す。
func (m Model) footerView() string {
	if m.err != nil {
		return errStyle.Render("error: " + m.err.Error())
	}
	return m.help.View(m.keys)
}
