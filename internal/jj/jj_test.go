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

// diff 3種の引数組み立てを固定する。Diff/DiffFile は --color always で色を強制し、
// DiffSummary は色不要なので --summary のみ（--color を付けない）。
// 実行部分（Diff/DiffFile/DiffSummary）は実 jj が要るため手動確認に委ねる。
func TestDiffArgs(t *testing.T) {
	got := diffArgs("abc123")
	want := []string{"diff", "--no-pager", "--color", "always", "-r", "abc123"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("diffArgs = %v, want %v", got, want)
	}
}

func TestDiffFileArgs(t *testing.T) {
	got := diffFileArgs("abc123", "src/x.go")
	want := []string{"diff", "--no-pager", "--color", "always", "-r", "abc123", "src/x.go"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("diffFileArgs = %v, want %v", got, want)
	}
}

func TestDiffSummaryArgs(t *testing.T) {
	got := diffSummaryArgs("abc123")
	want := []string{"diff", "--no-pager", "--summary", "-r", "abc123"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("diffSummaryArgs = %v, want %v", got, want)
	}
}
