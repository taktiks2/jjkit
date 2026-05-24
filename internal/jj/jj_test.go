package jj

import (
	"reflect"
	"testing"
)

// logArgs（jj log に渡す引数の組み立て）を固定する。--color always（パイプでも色を強制）と
// --no-pager（ページャ無効）が必ず入ること、テンプレートが -T で渡ることを保証する。
// Root() / LogRaw() の実行部分は実 jj とリポジトリが要るため、Task 7 の手動確認に委ねる。
func TestLogArgs(t *testing.T) {
	got := logArgs("TPL")
	want := []string{"log", "--no-pager", "--color", "always", "-T", "TPL"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("logArgs = %v, want %v", got, want)
	}
}
