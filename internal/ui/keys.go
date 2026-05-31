// Package ui は Bubble Tea による Log ペイン（起動・描画・change 単位ナビ）を実装する。
package ui

import (
	"charm.land/bubbles/v2/key"
)

// keyMap は Log ペインのキーバインド一式。help パッケージに渡すと
// フッターやヘルプ一覧を自動生成してくれる。
type keyMap struct {
	Up      key.Binding
	Down    key.Binding
	Refresh key.Binding
	Help    key.Binding
	Tab     key.Binding
	Quit    key.Binding
	New     key.Binding
	Edit    key.Binding
	Abandon key.Binding
	Confirm key.Binding
	Cancel  key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		Up:      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:    key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Refresh: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
		Help:    key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Tab:     key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "focus")),
		Quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		New:     key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new")),
		Edit:    key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
		Abandon: key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "abandon")),
		Confirm: key.NewBinding(key.WithKeys("enter", "y"), key.WithHelp("enter/y", "confirm")),
		Cancel:  key.NewBinding(key.WithKeys("esc", "n"), key.WithHelp("esc/n", "cancel")),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.New, k.Edit, k.Abandon, k.Refresh, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},
		{k.New, k.Edit, k.Abandon},
		{k.Refresh, k.Tab},
		{k.Help, k.Quit},
	}
}
