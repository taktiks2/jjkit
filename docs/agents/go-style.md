# Go style: Go 1.21+ 慣用句で書く

`go.mod` の `go` directive が 1.21 以降のとき、新規・改修コードでは下の慣用句に
従う。理由は AI 含む書き手のスタイル差を吸収し、レビューを「狙い」に集中させるため。

## 必須イディオム

### 1. 要素検索は `slices.Index` / `slices.Contains`

```go
// ❌ 手書き for ループでインデックスを探す
cur := 0
for i, p := range list {
    if p == target {
        cur = i
        break
    }
}

// ✅ slices.Index を使う。見つからないときの -1 は必ず明示処理する。
cur := slices.Index(list, target)
if cur < 0 {
    panic("target not in list") // or return error — 黙って 0 にしない
}
```

### 2. 整数範囲ループは `for i := range n` (Go 1.22+)

```go
// ❌ C-style for
for i := 0; i < n; i++ { ... }

// ✅ range over int
for i := range n { ... }
```

### 3. 大小比較は `min` / `max` builtin (Go 1.21+)

```go
// ❌ if 文で挟む
if x > last { x = last }

// ✅ builtin
x = min(x, last)
```

### 4. ループ変数の `i := i` は書かない (Go 1.22+)

Go 1.22 以降、`for _, v := range xs` の `v` は反復ごとに新スコープなので、
goroutine やクロージャに渡すための再代入は不要。

```go
// ❌
for _, v := range xs {
    v := v
    go func() { use(v) }()
}

// ✅
for _, v := range xs {
    go func() { use(v) }()
}
```

### 5. ソート/比較関数は `slices.SortFunc` + `cmp.Compare`

```go
// ❌ sort.Slice
sort.Slice(rows, func(i, j int) bool { return rows[i].N < rows[j].N })

// ✅ slices.SortFunc + cmp.Compare
slices.SortFunc(rows, func(a, b Row) int { return cmp.Compare(a.N, b.N) })
```

### 6. マップのキー/値取り出しは `maps.Keys` / `maps.Values` (iter)

```go
// ❌ 手書きで詰め直す
keys := make([]string, 0, len(m))
for k := range m { keys = append(keys, k) }

// ✅ slices.Collect + maps.Keys (Go 1.23+ iter)
keys := slices.Collect(maps.Keys(m))
```

### 7. マップ・スライスを空にするのは `clear` builtin

```go
// ❌ 再代入で作り直す（古い参照が残る）
m = map[K]V{}

// ✅ in-place で空にする
clear(m)
```

## 自動チェック

`.golangci.yml` の `intrange` / `copyloopvar` / `usestdlibvars` が一部を機械的に
検出する（ルール 2・4 と stdlib 定数）。**ルール 1・3・5・6・7 は lint で機械検出が
難しいので、レビューで守る**。`just check` を必ず通すこと。

## 例外

- 既存テストヘルパで `sort.Slice` を使っていて読みやすさが勝るときは温存可。
  ただし新規コードでは ✅ 側で書く。
- `slices.Index` で見つからないケースが想定内なら、戻り値 -1 を握りつぶさず
  分岐すること（黙ってのフォールバックは禁止）。
