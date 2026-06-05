package ui

import (
	tea "charm.land/bubbletea/v2"
)

// Modal は overlay 表示する確認 / ローディング / 編集 dialog の seam。
// Model.modal == nil が「modal 無し（通常モード）」を表す。
//
// 新しい modal を足すときは:
//   - 新しい struct 型を作って Modal を満たす
//   - 該当キーを受けた Normal モードのハンドラで Model.modal にセットする
//
// view.go / Update / footer の switch を追加する必要は無い。
type Modal interface {
	// HandleKey は modal が active な間のキー押下を処理する。
	// 戻り値 next==nil は「modal を閉じる」を意味する。
	// next==nil かつ cmd!=nil のとき、ホストは op を発射したとみなして opInFlight を立てる。
	HandleKey(msg tea.KeyPressMsg, keys keyMap) (next Modal, cmd tea.Cmd)

	// HandleMsg はこの modal 宛の非同期メッセージ（例: descLoadedMsg）を処理する。
	// currentChange は stale ガード用にホストの現在選択 change ID を渡す。
	// err!=nil のときホストはエラーを画面に出し、modal は閉じたものとして扱う。
	HandleMsg(msg tea.Msg, currentChange string) (next Modal, cmd tea.Cmd, err error)

	// View は modal 本体（中央の box）を描画する。width は端末幅。
	View(width int) string

	// FooterHint はこの modal が active な間 footer に出すヒント。
	FooterHint() string

	// Resize は端末リサイズに応じて内部 widget の幅を更新する。
	// abandon のように追従不要な modal は自身をそのまま返してよい。
	Resize(width int) Modal
}
