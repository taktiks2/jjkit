package ui

import (
	"slices"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
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

type mode int

const (
	modeNormal            mode = iota
	modeDescribingLoading      // describe seed fetch 中
	modeDescribing             // textinput 編集中
	modeConfirmingAbandon      // abandon 確認待ち
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
	opResultMsg  struct{ err error }
	// descLoadedMsg は describe modal の seed 取得結果。
	// target は要求時点の change ID。loading 中に @ が動いても stale 判定できる。
	descLoadedMsg struct {
		target string
		desc   string
		err    error
	}
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

	// 書き込み操作 (n/e/d/a) 中のフラグ。
	// 真の間は n/e/d/a を吸う（読み込み系 r/Up/Down/Tab は通す）。
	opInFlight bool
	mode       mode
	// describe modal の対象 change ID。stale 判定に使う。
	// loading 中に @ が動いても、descLoadedMsg.target と m.selectedChangeID() の照合で
	// 古い結果を捨てられる。Cancel 時に "" にリセット。
	descTarget string
	// describe modal の text-input。SetValue(seed) → ユーザー編集 → Value() で submit。
	descInput textinput.Model
}

// New は初期化済みの Model を返す。
func New() Model {
	return Model{
		keys:      newKeyMap(),
		help:      help.New(),
		logVP:     viewport.New(),
		filesVP:   viewport.New(),
		diffVP:    viewport.New(),
		descInput: textinput.New(),
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
		m.refreshFiles()
		m.refreshDiff()
		return m, nil
	case logLoadedMsg:
		prevID := m.selectedChangeID()
		m.log = msg.log
		m.err = nil
		if i := m.log.RowByChangeID(prevID); i >= 0 {
			m.logSel = i
		} else {
			m.logSel = m.log.WorkingCopyRow()
		}
		m.refreshLog()
		m.scrollLog()
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
		return m, tea.Batch(loadLog, m.filesCmd(), m.diffCmd())
	case filesLoadedMsg:
		if msg.change != m.selectedChangeID() {
			return m, nil // 別 change 宛の古い結果 -> 捨てる
		}
		m.files = msg.files
		m.fileSel = 0
		m.refreshFiles()
		return m, nil
	case diffLoadedMsg:
		if msg.req != m.currentDiffReq() {
			return m, nil // 選択が動いたあとに届いた古い結果 -> 捨てる
		}
		m.diffContent = string(msg.raw)
		m.refreshDiff()
		return m, nil
	case descLoadedMsg:
		// stale ガード 2 種
		if m.mode != modeDescribingLoading {
			return m, nil // 既にキャンセル済み
		}
		if msg.target != m.selectedChangeID() {
			return m, nil // 選択が動いた
		}
		if msg.err != nil {
			m.mode = modeNormal
			m.descTarget = ""
			m.err = msg.err
			return m, nil
		}
		// 成功: modeDescribing に遷移、textinput に seed セット、Focus blink Cmd を返す。
		m.mode = modeDescribing
		m.descInput.SetValue(msg.desc)
		m.descInput.CursorEnd()
		return m, m.descInput.Focus()
	case tea.KeyPressMsg:
		switch m.mode {
		case modeConfirmingAbandon:
			return m.updateConfirmingAbandon(msg)
		case modeDescribingLoading:
			return m.updateDescribingLoading(msg)
		case modeDescribing:
			return m.updateDescribing(msg)
		default:
			return m.updateNormal(msg)
		}
	}
	return m, nil
}

// jjAbandonCmd は選択中の change を破棄する jj abandon を非同期で実行する Cmd を返す。
func (m Model) jjAbandonCmd() tea.Cmd {
	change := m.selectedChangeID()
	if change == "" {
		return nil
	}
	return func() tea.Msg {
		return opResultMsg{err: jj.Abandon(change)}
	}
}

// updateConfirmingAbandon は abandon 確認 modal 中のキー処理。
// Confirm → jj abandon 発射、Cancel → 戻る、それ以外は吸う。
func (m Model) updateConfirmingAbandon(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Confirm):
		cmd := m.jjAbandonCmd()
		m.mode = modeNormal
		if cmd == nil {
			return m, nil
		}
		m.opInFlight = true
		return m, cmd
	case key.Matches(msg, m.keys.Cancel):
		m.mode = modeNormal
		return m, nil
	}
	return m, nil // それ以外 (j/k 等)は吸う
}

// updateNormal は通常モードのキー処理。既存の KeyPressMsg 処理を切り出したもの。
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
		return m, m.diffCmd()
	case key.Matches(msg, m.keys.New):
		if m.opInFlight {
			return m, nil
		}
		cmd := m.jjNewCmd()
		if cmd == nil {
			return m, nil
		}
		m.opInFlight = true
		return m, cmd
	case key.Matches(msg, m.keys.Edit):
		if m.opInFlight {
			return m, nil
		}
		cmd := m.jjEditCmd()
		if cmd == nil {
			return m, nil
		}
		m.opInFlight = true
		return m, cmd
	case key.Matches(msg, m.keys.Abandon):
		if m.opInFlight {
			return m, nil
		}
		if m.selectedChangeID() == "" {
			return m, nil
		}
		m.mode = modeConfirmingAbandon
		return m, nil
	case key.Matches(msg, m.keys.Describe):
		if m.opInFlight {
			return m, nil
		}
		cmd := m.descCmd()
		if cmd == nil {
			return m, nil
		}
		m.mode = modeDescribingLoading
		m.descTarget = m.selectedChangeID()
		return m, cmd
	case key.Matches(msg, m.keys.Refresh):
		return m, loadLog
	case key.Matches(msg, m.keys.Help):
		m.help.ShowAll = !m.help.ShowAll
		m.applyLayout()
		m.refreshLog()
		m.refreshFiles()
		m.refreshDiff()
		return m, nil
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

	// modal box = min(60, m.width-4)、padding(1, 2) で左右合計 4 引いた内寸。
	boxW := min(60, m.width-4)
	m.descInput.SetWidth(max(boxW-4, 0))
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

// diffCmd は今表示すべき diff（currentDiffReq）を非同期に取りに行く Cmd を返す。
// 別 goroutine で走り、結果は diffLoadedMsg として Update に届く（stale 判定はそこ）。
func (m Model) diffCmd() tea.Cmd {
	req := m.currentDiffReq()
	if req.change == "" {
		return nil
	}
	return func() tea.Msg {
		var (
			raw []byte
			err error
		)
		if req.file == "" {
			raw, err = jj.Diff(req.change)
		} else {
			raw, err = jj.DiffFile(req.change, req.file)
		}
		if err != nil {
			return logErrMsg{err}
		}
		return diffLoadedMsg{req: req, raw: raw}
	}
}

// filesCmd は選択中の change の変更ファイル一覧を取りに行く Cmd を返す。
func (m Model) filesCmd() tea.Cmd {
	change := m.selectedChangeID()
	if change == "" {
		return nil
	}
	return func() tea.Msg {
		raw, err := jj.DiffSummary(change)
		if err != nil {
			return logErrMsg{err}
		}
		return filesLoadedMsg{change: change, files: jjdiff.ParseSummary(raw)}
	}
}

// loadForChange は change が変わったとき（ファイル一覧 + change 全体 diff の両方）に使う。
// tea.Batch は複数の Cmd を並列に走らせる。
func (m Model) loadForChange() tea.Cmd {
	return tea.Batch(m.filesCmd(), m.diffCmd())
}

// loadAfterMove は移動後にフォーカスに応じた再ロードを返す。
// Log を動かしたら change が変わったので files + diff、Files を動かしたら diff だけ。
func (m Model) loadAfterMove() tea.Cmd {
	switch m.focus {
	case paneLog:
		return m.loadForChange()
	case paneFiles:
		return m.diffCmd()
	default:
		return nil
	}
}

// jjNewCmd は選択中の change に対する jj new を非同期で実行する Cmd を返す。
// 結果は opResultMsg として Update に届く。
func (m Model) jjNewCmd() tea.Cmd {
	change := m.selectedChangeID()
	if change == "" {
		return nil
	}
	return func() tea.Msg {
		return opResultMsg{err: jj.New(change)}
	}
}

// jjEditCmd は選択中の change に @ を移す jj edit を非同期で実行する Cmd を返す。
func (m Model) jjEditCmd() tea.Cmd {
	change := m.selectedChangeID()
	if change == "" {
		return nil
	}
	return func() tea.Msg {
		return opResultMsg{err: jj.Edit(change)}
	}
}

// descCmd は選択中の change の description を非同期で取得する Cmd を返す。
// 結果は descLoadedMsg として Update に届く（target で stale 判定）。
func (m Model) descCmd() tea.Cmd {
	change := m.selectedChangeID()
	if change == "" {
		return nil
	}
	return func() tea.Msg {
		desc, err := jj.Description(change)
		return descLoadedMsg{target: change, desc: desc, err: err}
	}
}

// updateDescribingLoading は describe seed fetch 中のキー処理。
// Cancel で modeNormal に戻る（取得中もキャンセル可）。それ以外は吸う。
func (m Model) updateDescribingLoading(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, m.keys.Cancel) {
		m.mode = modeNormal
		m.descTarget = ""
		return m, nil
	}
	return m, nil
}

// updateDescribing は describe modal (textinput 編集中) のキー処理。
// Submit (Enter) → jj describe 発射、Cancel (Esc) → 戻る、それ以外は textinput に流す。
// Submit は "enter" 限定なので、textinput への "y" 入力等は Confirm に吸われない。
func (m Model) updateDescribing(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Submit):
		cmd := m.jjDescribeCmd(m.descInput.Value())
		m.mode = modeNormal
		m.descInput.Reset()
		m.descTarget = ""
		if cmd == nil {
			return m, nil
		}
		m.opInFlight = true
		return m, cmd
	case key.Matches(msg, m.keys.Cancel):
		m.mode = modeNormal
		m.descInput.Reset()
		m.descTarget = ""
		return m, nil
	}
	// それ以外のキーは textinput に流す (j/k/y 等も普通に文字入力になる)。
	var cmd tea.Cmd
	m.descInput, cmd = m.descInput.Update(msg)
	return m, cmd
}

// jjDescribeCmd は <descTarget> の description を <msg> に置き換える jj describe を
// 非同期で実行する Cmd を返す。submit 時点で @ が動いていても、describe 対象は当初
// 押下した change のまま (descTarget) で良い。
func (m Model) jjDescribeCmd(msg string) tea.Cmd {
	change := m.descTarget
	if change == "" {
		return nil
	}
	return func() tea.Msg {
		return opResultMsg{err: jj.Describe(change, msg)}
	}
}
