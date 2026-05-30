package ui

import (
	"slices"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	"github.com/taktiks2/jjkit/internal/jj"
	"github.com/taktiks2/jjkit/internal/jjdiff"
	"github.com/taktiks2/jjkit/internal/jjlog"
)

type pane int

const (
	paneLog pane = iota
	paneFiles
	paneDiff
	paneBookmarks
	paneOplog
)

var focusCycle = []pane{paneLog, paneFiles, paneBookmarks}

func (m Model) cycleFocus() Model {
	cur := slices.Index(focusCycle, m.focus)
	if cur < 0 {
		panic("focus not in focusCycle")
	}
	m.focus = focusCycle[(cur+1)%len(focusCycle)]
	return m
}

// loadLog の結果メッセージ。成功は logLoadedMsg、失敗は logErrMsg。
type (
	logLoadedMsg struct{ log *jjlog.Log }
	logErrMsg    struct{ err error }
)

// Model は Log ペインの状態。
type Model struct {
	focus       pane
	keys        keyMap
	help        help.Model
	logVP       viewport.Model
	filesVP     viewport.Model
	diffVP      viewport.Model
	log         *jjlog.Log
	logSel      int
	files       []jjdiff.FileChange
	fileSel     int
	width       int
	height      int
	ready       bool
	diffContent string
	err         error
}

// New は初期化済みの Model を返す。
func New() Model {
	return Model{
		keys:    newKeyMap(),
		help:    help.New(),
		logVP:   viewport.New(),
		filesVP: viewport.New(),
		diffVP:  viewport.New(),
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
		m.applyLayout()
		m.ready = true
		m.refreshLog()
		return m, nil
	case logLoadedMsg:
		m.log = msg.log
		m.err = nil
		m.logSel = m.log.WorkingCopyRow()
		m.refreshLog()
		m.scrollLog()
		return m, nil
	case filesLoadedMsg:
		if msg.change != m.selectedChangeID() {
			return m, nil // 別 change 宛の古い結果 -> 捨てる
		}
		m.files = msg.files
		m.fileSel = 0
		return m, nil
	case logErrMsg:
		m.err = msg.err
		return m, nil
	case diffLoadedMsg:
		if msg.req != m.currentDiffReq() {
			return m, nil // 選択が動いたあとに届いた古い結果 -> 捨てる
		}
		m.diffContent = string(msg.raw)
		return m, nil
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.Tab):
			m = m.cycleFocus()
			return m, nil
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
			m.applyLayout()
			m.refreshLog()
			return m, nil
		}
	}
	return m, nil
}

// moveSelection はカーソルを delta だけ動かす（範囲内にクランプ）。
func moveSelection(m Model, delta int) Model {
	switch m.focus {
	case paneFiles:
		return moveFileSel(m, delta)
	case paneLog:
		return moveLogSel(m, delta)
	default:
		return m
	}
}

func moveLogSel(m Model, delta int) Model {
	if m.log == nil || len(m.log.Rows) == 0 {
		return m
	}
	m.logSel = max(0, min(m.logSel+delta, len(m.log.Rows)-1))
	m.refreshLog()
	m.scrollLog()
	return m
}

func moveFileSel(m Model, delta int) Model {
	if len(m.files) == 0 {
		return m
	}
	m.fileSel = max(0, min(m.fileSel+delta, len(m.files)-1))
	return m
}

// applyLayout は computeLayout の外寸から、各 viewport の内寸（枠2 + タイトル1 を除く）を渡す。
func (m *Model) applyLayout() {
	footerH := lipglossHeight(m.footerView())
	l := computeLayout(m.width, m.height, footerH)
	setVP(&m.logVP, l.log)
	setVP(&m.filesVP, l.files)
	setVP(&m.diffVP, l.diff)
}

// setVP は外寸 rect から枠(2)とタイトル行(1)を引いた内寸を viewport に渡す。
func setVP(vp *viewport.Model, r rect) {
	vp.SetWidth(max(r.w-2, 0))
	vp.SetHeight(max(r.h-3, 0))
}

// refreshLog は Log ペインに、選択行ハイライトを乗せた jj log を流す。
func (m *Model) refreshLog() {
	if m.log == nil {
		return
	}
	start, end := m.log.LineRange(m.logSel)
	m.logVP.SetContent(RenderContent(m.log.Lines(), start, end, m.logVP.Width()))
}

// scrollLog は選択 change が画面内に収まるようスクロールする。
func (m *Model) scrollLog() {
	if m.log == nil || m.logVP.Height() <= 0 {
		return
	}
	start, end := m.log.LineRange(m.logSel)
	m.logVP.EnsureVisible(end-1, 0, 0)
	m.logVP.EnsureVisible(start, 0, 0)
}

// refreshFiles は変更ファイル一覧を「ステータス + パス」の1行に整形し、選択ハイライト付きで流す。
func (m *Model) refreshFiles() {
	lines := make([]string, len(m.files))
	for i, f := range m.files {
		lines[i] = f.Status + " " + f.Path
	}
	m.filesVP.SetContent(RenderContent(lines, m.fileSel, m.fileSel+1, m.filesVP.Width()))
}

// refreshDiff は保持している diff 本文を Diff viewport に流し、先頭へ戻す。
func (m *Model) refreshDiff() {
	m.diffVP.SetContent(m.diffContent)
	m.diffVP.GotoTop()
}

// diffReq は Diff ペインに出すべき内容の識別子。file=="" は change 全体を表す。
// 非同期ロードの stale 判定にも使う
type diffReq struct {
	change string
	file   string
}

// selectedChangeID は今 Log で選んでいる change の id（無ければ ""）。
// 範囲外・nil ガードを集約することで currentDiffReq 等の呼び出し側を綺麗に保つ。
func (m Model) selectedChangeID() string {
	if m.log == nil || m.logSel < 0 || m.logSel >= len(m.log.Rows) {
		return ""
	}
	return m.log.Rows[m.logSel].ChangeID
}

// currentDiffReq は (focus, 選択) から「今 Diff に出すべき内容」を決める。
// Files にフォーカスしていてファイルがあるときだけファイル単位、それ以外は change 全体。
func (m Model) currentDiffReq() diffReq {
	req := diffReq{change: m.selectedChangeID()}
	if m.focus == paneFiles && m.fileSel >= 0 && m.fileSel < len(m.files) {
		req.file = m.files[m.fileSel].Path
	}
	return req
}

// diffLoadedMsg は jj diff の結果。req は「何を要求したか」で stale 判定に使う。
type diffLoadedMsg struct {
	req diffReq
	raw []byte
}

// filesLoadedMsg は jj diff --summary の結果。change で stale 判定する。
type filesLoadedMsg struct {
	change string
	files  []jjdiff.FileChange
}
