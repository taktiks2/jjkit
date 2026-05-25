package jjdiff

import (
	"reflect"
	"testing"
)

// 複数行の summary を status と path に分解できること。
func TestParseSummaryParsesStatusAndPath(t *testing.T) {
	raw := []byte("M src/x.go\nA README.md\nD old.txt\n")
	got := ParseSummary(raw)
	want := []FileChange{
		{Status: "M", Path: "src/x.go"},
		{Status: "A", Path: "README.md"},
		{Status: "D", Path: "old.txt"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseSummary = %v, want %v", got, want)
	}
}

// 空入力（diff 無しの change）で空スライスを返すこと。
func TestParseSummaryEmpty(t *testing.T) {
	if got := ParseSummary([]byte("")); len(got) != 0 {
		t.Errorf("ParseSummary(empty) = %v, want empty", got)
	}
}

// パスに空白が含まれても、最初の空白だけで分割して path を壊さないこと。
func TestParseSummaryPreservesSpacesInPath(t *testing.T) {
	got := ParseSummary([]byte("M my file.go\n"))
	want := []FileChange{{Status: "M", Path: "my file.go"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseSummary = %v, want %v", got, want)
	}
}
