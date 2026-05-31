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

// Issue #3: new/edit/describe/abandon/description 5関数の引数組み立てを固定する。
// 実行部分（実 jj が要る）は手動確認に委ねる。

// newArgs: jj new --no-pager <change>
func TestNewArgs(t *testing.T) {
	got := newArgs("abc123")
	want := []string{"new", "--no-pager", "abc123"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("newArgs = %v, want %v", got, want)
	}
}

// editArgs: jj edit --no-pager <change>
func TestEditArgs(t *testing.T) {
	got := editArgs("abc123")
	want := []string{"edit", "--no-pager", "abc123"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("editArgs = %v, want %v", got, want)
	}
}

// describeArgs: jj describe --no-pager -r <change> -m <msg>
// -r で対象 change を明示し、-m で description 文字列を渡す（対話モード回避）。
func TestDescribeArgs(t *testing.T) {
	got := describeArgs("abc123", "feat: hello")
	want := []string{"describe", "--no-pager", "-r", "abc123", "-m", "feat: hello"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("describeArgs = %v, want %v", got, want)
	}
}

// abandonArgs: jj abandon --no-pager <change>
func TestAbandonArgs(t *testing.T) {
	got := abandonArgs("abc123")
	want := []string{"abandon", "--no-pager", "abc123"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("abandonArgs = %v, want %v", got, want)
	}
}

// descriptionArgs: jj log --no-pager --no-graph -T description -r <change>
// describe modal の seed 取得用。--no-graph で graph 装飾を抑え、-T description で
// description 文字列だけを評価する。
func TestDescriptionArgs(t *testing.T) {
	got := descriptionArgs("abc123")
	want := []string{"log", "--no-pager", "--no-graph", "-T", "description", "-r", "abc123"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("descriptionArgs = %v, want %v", got, want)
	}
}
