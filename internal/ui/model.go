package ui

import (
	"slices"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	"github.com/taktiks2/jjkit/internal/jj"
)

type pane int

const (
	paneLog pane = iota
	paneFiles
	paneDiff
	paneBookmarks
	paneOplog
)

var focusCycle = []pane{paneLog, paneFiles, paneBookmarks, paneOplog}

func (m Model) cycleFocus() Model {
	cur := slices.Index(focusCycle, m.focus)
	if cur < 0 {
		panic("focus not in focusCycle")
	}
	m.focus = focusCycle[(cur+1)%len(focusCycle)]
	return m
}

// 横断的な非同期メッセージ。pane 固有の logLoadedMsg / filesLoadedMsg / diffLoadedMsg
// は各 pane_*.go に居る。descLoadedMsg は describeModal が消費する。
type (
	logErrMsg   struct{ err error }
	opResultMsg struct{ err error }
	// descLoadedMsg は describe modal の seed 取得結果。
	descLoadedMsg struct {
		target string
		desc   string
		err    error
	}
)

// Model は jjkit TUI の状態。pane の内部状態は各 pane struct に閉じ込め、
// Model は focus / modal / opInFlight / err / window size のような横断的な状態だけを持つ。
type Model struct {
	focus pane
	keys  keyMap
	help  help.Model

	log       logPane
	files     filesPane
	diff      diffPane
	bookmarks bookmarksPane
	oplog     oplogPane

	width  int
	height int
	ready  bool
	err    error

	// opInFlight は jj 書き込み操作 (n/e/d/a) を発射してから opResultMsg を受けるまで真。
	// 真の間は n/e/d/a を吸う（読み込み系 r/Up/Down/Tab は通す）。
	opInFlight bool
	// modal == nil なら通常モード。非 nil なら overlay の対話中（abandon / describe）。
	modal Modal
}

// New は初期化済みの Model を返す。
func New() Model {
	return Model{
		keys: newKeyMap(),
		help: help.New(),
	}
}

// Init は起動時に最初の log + bookmarks + oplog 読込を仕掛ける。
func (m Model) Init() tea.Cmd { return tea.Batch(loadLog, bookmarksCmd, oplogCmd) }

// Update はメッセージを受けて状態を更新する（Bubble Tea の中心）。
// modal active 時はキー押下を modal に委譲し、未知の async msg も modal へ流す。
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.help.SetWidth(msg.Width)
		m.applyLayout()
		if m.modal != nil {
			m.modal = m.modal.Resize(m.width)
		}
		m.ready = true
		return m, nil
	case logLoadedMsg:
		prevID := m.log.SelectedChangeID()
		m.err = nil
		m.log.Apply(msg.log, prevID)
		return m, nil
	case logErrMsg:
		m.err = msg.err
		return m, nil
	case opResultMsg:
		m.opInFlight = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.err = nil
		return m, tea.Batch(loadLog, bookmarksCmd, oplogCmd, filesCmd(m.log.SelectedChangeID()), diffCmd(m.currentDiffReq()))
	case filesLoadedMsg:
		m.files.Apply(msg, m.log.SelectedChangeID())
		return m, nil
	case diffLoadedMsg:
		m.diff.Apply(msg, m.currentDiffReq())
		return m, nil
	case bookmarksLoadedMsg:
		m.bookmarks.Apply(msg.items)
		return m, nil
	case oplogLoadedMsg:
		m.err = nil
		m.oplog.Apply(msg.log)
		return m, nil
	case tea.KeyPressMsg:
		if m.modal != nil {
			next, cmd := m.modal.HandleKey(msg, m.keys)
			// next==nil かつ cmd!=nil は「dismiss しつつ jj op を発射」を意味する。
			fired := next == nil && cmd != nil
			m.modal = next
			if fired {
				m.opInFlight = true
			}
			return m, cmd
		}
		return m.updateNormal(msg)
	default:
		// modal 宛の非同期 msg（descLoadedMsg など）を委譲する。
		if m.modal == nil {
			return m, nil
		}
		next, cmd, err := m.modal.HandleMsg(msg, m.log.SelectedChangeID())
		m.modal = next
		if err != nil {
			m.err = err
		}
		return m, cmd
	}
}

// updateNormal は通常モード（modal 無し）のキー処理。
func (m Model) updateNormal(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.keys.Up):
		m = moveSelection(m, -1)
		return m, m.loadAfterMove()
	case key.Matches(msg, m.keys.Down):
		m = moveSelection(m, +1)
		return m, m.loadAfterMove()
	case key.Matches(msg, m.keys.Tab):
		m = m.cycleFocus()
		return m, diffCmd(m.currentDiffReq())
	case key.Matches(msg, m.keys.New):
		if m.opInFlight {
			return m, nil
		}
		cmd := jjNewCmd(m.log.SelectedChangeID())
		if cmd == nil {
			return m, nil
		}
		m.opInFlight = true
		return m, cmd
	case key.Matches(msg, m.keys.Edit):
		if m.opInFlight {
			return m, nil
		}
		cmd := jjEditCmd(m.log.SelectedChangeID())
		if cmd == nil {
			return m, nil
		}
		m.opInFlight = true
		return m, cmd
	case key.Matches(msg, m.keys.Abandon):
		if m.opInFlight {
			return m, nil
		}
		target := m.log.SelectedChangeID()
		if target == "" {
			return m, nil
		}
		m.modal = abandonModal{target: target}
		return m, nil
	case key.Matches(msg, m.keys.Advance):
		if m.opInFlight {
			return m, nil
		}
		if m.focus != paneBookmarks {
			return m, nil
		}
		name := m.bookmarks.SelectedName()
		if name == "" {
			return m, nil
		}
		m.opInFlight = true
		return m, jjBookmarkAdvanceCmd(name)
	case key.Matches(msg, m.keys.Describe):
		if m.opInFlight {
			return m, nil
		}
		target := m.log.SelectedChangeID()
		if target == "" {
			return m, nil
		}
		m.modal = newDescribeModal(target, m.width)
		return m, descSeedCmd(target)
	case key.Matches(msg, m.keys.Undo):
		if m.opInFlight {
			return m, nil
		}
		m.modal = undoModal{}
		return m, nil
	case key.Matches(msg, m.keys.Refresh):
		return m, loadLog
	case key.Matches(msg, m.keys.Help):
		m.help.ShowAll = !m.help.ShowAll
		m.applyLayout()
		return m, nil
	}
	return m, nil
}

// moveSelection はカーソルを delta だけ動かす（フォーカスペインのみ）。
func moveSelection(m Model, delta int) Model {
	switch m.focus {
	case paneFiles:
		m.files.Move(delta)
	case paneLog:
		m.log.Move(delta)
	case paneBookmarks:
		m.bookmarks.Move(delta)
	case paneOplog:
		m.oplog.Move(delta)
	}
	return m
}

// applyLayout は computeLayout の外寸から、各 pane に viewport 内寸を伝播する。
// 各 pane の Resize() は viewport サイズ変更後に refresh も呼ぶ（行幅依存のハイライト再計算）。
func (m *Model) applyLayout() {
	footerH := lipglossHeight(m.footerView())
	l := computeLayout(m.width, m.height, footerH)
	m.log.Resize(l.log)
	m.files.Resize(l.files)
	m.diff.Resize(l.diff)
	m.bookmarks.Resize(l.bookmarks)
	m.oplog.Resize(l.oplog)
}

// setVP は外寸 rect から枠(2)とタイトル行(1)を引いた内寸を viewport に渡す。
// 各 pane の Resize から呼ばれる共通ヘルパー。
func setVP(vp *viewport.Model, r rect) {
	vp.SetWidth(max(r.w-2, 0))
	vp.SetHeight(max(r.h-3, 0))
}

// currentDiffReq は (focus, 選択) から「今 Diff に出すべき内容」を決める。
// Files にフォーカスしていてファイルがあるときだけファイル単位、それ以外は change 全体。
func (m Model) currentDiffReq() diffReq {
	req := diffReq{change: m.log.SelectedChangeID()}
	if m.focus == paneFiles {
		if path := m.files.SelectedPath(); path != "" {
			req.file = path
		}
	}
	return req
}

// loadAfterMove は移動後にフォーカスに応じた再ロードを返す。
// Log を動かしたら change が変わったので files + diff、Files を動かしたら diff だけ。
func (m Model) loadAfterMove() tea.Cmd {
	switch m.focus {
	case paneLog:
		return tea.Batch(filesCmd(m.log.SelectedChangeID()), diffCmd(m.currentDiffReq()))
	case paneFiles:
		return diffCmd(m.currentDiffReq())
	default:
		return nil
	}
}

// jjNewCmd は change を親に持つ空の新 change を作る jj new の Cmd。change 空なら nil。
func jjNewCmd(change string) tea.Cmd {
	if change == "" {
		return nil
	}
	return func() tea.Msg {
		return opResultMsg{err: jj.New(change)}
	}
}

// jjEditCmd は @ を change に移す jj edit の Cmd。
func jjEditCmd(change string) tea.Cmd {
	if change == "" {
		return nil
	}
	return func() tea.Msg {
		return opResultMsg{err: jj.Edit(change)}
	}
}

// jjBookmarkAdvanceCmd は <name> を @ に進める Cmd。成功・失敗どちらも opResultMsg に乗せる。
func jjBookmarkAdvanceCmd(name string) tea.Cmd {
	return func() tea.Msg {
		return opResultMsg{err: jj.BookmarkMove(name, "@")}
	}
}
