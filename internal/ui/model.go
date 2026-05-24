package ui

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	"github.com/taktiks2/jjkit/internal/jj"
	"github.com/taktiks2/jjkit/internal/jjlog"
)

// loadLog の結果メッセージ。成功は logLoadedMsg、失敗は logErrMsg。
type (
	logLoadedMsg struct{ log *jjlog.Log }
	logErrMsg    struct{ err error }
)

// Model は Log ペインの状態。
type Model struct {
	keys     keyMap
	help     help.Model
	viewport viewport.Model
	log      *jjlog.Log
	selected int
	width    int
	height   int
	ready    bool
	err      error
}

// New は初期化済みの Model を返す。
func New() Model {
	return Model{
		keys:     newKeyMap(),
		help:     help.New(),
		viewport: viewport.New(),
	}
}

// Init は起動時に最初の log 読込を仕掛ける。
func (m Model) Init() tea.Cmd {
	return loadLog
}

// loadLog は jj log を実行してパースする Cmd。Update ループの外（別 goroutine）で走る。
func loadLog() tea.Msg {
	raw, err := jj.LogRaw(jjlog.Template)
	if err != nil {
		return logErrMsg{err}
	}
	parsed, err := jjlog.Parse(raw)
	if err != nil {
		return logErrMsg{err}
	}
	return logLoadedMsg{parsed}
}

// Update はメッセージを受けて状態を更新する（Bubble Tea の中心）。
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.help.SetWidth(msg.Width)
		m.layoutViewport()
		m.ready = true
		m.refreshContent()
		return m, nil
	case logLoadedMsg:
		m.log = msg.log
		m.err = nil
		m.selected = m.log.WorkingCopyRow()
		m.refreshContent()
		m.scrollToSelected()
		return m, nil
	case logErrMsg:
		m.err = msg.err
		return m, nil
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Up):
			return moveSelection(m, -1), nil
		case key.Matches(msg, m.keys.Down):
			return moveSelection(m, +1), nil
		case key.Matches(msg, m.keys.Refresh):
			return m, loadLog
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			m.layoutViewport()
			m.refreshContent()
			return m, nil
		}
	}
	return m, nil
}

// moveSelection はカーソルを delta だけ動かす（範囲内にクランプ）。
func moveSelection(m Model, delta int) Model {
	if m.log == nil || len(m.log.Rows) == 0 {
		return m
	}
	m.selected += delta
	if m.selected < 0 {
		m.selected = 0
	}
	if last := len(m.log.Rows) - 1; m.selected > last {
		m.selected = last
	}
	m.refreshContent()
	m.scrollToSelected()
	return m
}

// layoutViewport は viewport をフッターの上の領域いっぱいに合わせる。
func (m *Model) layoutViewport() {
	h := m.height - lipglossHeight(m.footerView())
	if h < 0 {
		h = 0
	}
	m.viewport.SetWidth(m.width)
	m.viewport.SetHeight(h)
}

// refreshContent は選択範囲を反映した表示テキストを viewport に流し込む。
func (m *Model) refreshContent() {
	if m.log == nil {
		return
	}
	start, end := m.log.LineRange(m.selected)
	m.viewport.SetContent(RenderContent(m.log.Lines(), start, end, m.viewport.Width()))
}

// scrollToSelected は選択 change が画面内に収まるようスクロールする。
func (m *Model) scrollToSelected() {
	if m.log == nil || m.viewport.Height() <= 0 {
		return
	}
	start, end := m.log.LineRange(m.selected)
	m.viewport.EnsureVisible(end-1, 0, 0)
	m.viewport.EnsureVisible(start, 0, 0)
}
