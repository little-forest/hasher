---
title: go-cli 準拠のレイアウトへの再編と hash/xattr のライブラリ公開
created: 2026-08-23
updated: 2026-08-23
status: 実装済み
---

# 要求

````text
このアプリケーションは、 go-cli スキルの規定する構成に沿っていません。違いを調べ、可能な限りgo-cliの構成にあわせる実装計画を立ててください。
また、このアプリケーションの hash 計算と xattr 属性への読み書き処理は、別のアプリケーションへライブラリとして提供する計画があることも加味してください。
````

# 設計: go-cli 準拠のレイアウトへの再編と hash/xattr のライブラリ公開

想定ブランチ名: `refactor/go-cli-alignment`

## 目的

hasher のリポジトリレイアウトと cobra まわりの書き方を `go-cli` スキルの規定に合わせ、同時に **hash 計算と xattr 読み書きを外部アプリから `go get` できる公開パッケージとして切り出す**。

ユーザーから見て変わるのは 3 点だけである。

- 失敗時に **Usage 全文が流れなくなる**（`SPEC-CLI-002` / `SPEC-OUT-005`）
- `update`（`-r` なし）が **失敗したときに終了コード 1 を返すようになる**（現状は 0。課題 B1）
- `Ctrl-C` / `SIGTERM` を受け取る配線が入る（**中断が効くようになるわけではない**。決定 15）

**変わらないもの**: 全サブコマンドの出力書式、TSV / JSON の書式、拡張属性の名前と値、色付け、進捗表示、`version` の出力。これは「構成の付け替え」であり、機能追加ではない。

## 現状（出発点）

### 読んだ既存設計文書

| 文書 | 状態 |
| --- | --- |
| [`ARCHITECTURE.md`](../../ARCHITECTURE.md) | リポジトリ直下に存在。全 447 行を読んだ |
| [`SPECS.md`](../../SPECS.md) | リポジトリ直下に存在。全 701 行を読んだ |

`docs/` 配下には設計文書が無く、`docs/` ディレクトリ自体が存在しなかった。

今回の設計を縛る既存の記述:

- `SPEC-CLI-002`（終了コードの決定規則）— **2 経路あり、片方は Usage 全文が出る**。決定 7 でこれを 1 経路に畳むため、この項目は改訂対象。
- `SPEC-OUT-005`（エラーメッセージの体裁）— 上と同じ理由で改訂対象。
- `SPEC-CLI-110`（`version`）— 出力書式が固定されている。`Run` → `RunE` の書き換えで**書式を変えてはならない**。
- `SPEC-XATTR-001`〜`006` — 属性名は `user.hasher.{アルゴリズム名}` 等。**パッケージ名を変えても属性名は変わってはならない**（属性名は Go のパッケージ構成と無関係なリテラルから組み立てられている）。
- `SPEC-LIMIT-004`（中断・再開に対応しない）— 決定 15 のとおり今回も「対応しない」まま。記述の微修正のみ。
- `SPEC-LIMIT-005`（並列度 1 固定）、`SPEC-LIMIT-003`（端末判定なし）— **今回の設計には効かない**（決定 15 で対象外）。
- `ARCHITECTURE.md` §11 の B1〜B18 — うち B1 / B12 / B15 のみ今回触れる（決定 15〜17）。

### go-cli スキルとの差分

`go-cli` の `SKILL.md`「コアルール」・`project-setup.md`「プロジェクト構成」・`references/cobra-guide.md` §13「導入チェックリスト」を突き合わせた結果。

| # | go-cli の規定 | hasher の現状 | 判定 |
| --- | --- | --- | --- |
| a | モジュールルート = リポジトリルート、`main.go` はルート直下 | `go.mod` と `main.go` が `src/` 配下。`go.mod` は `module github.com/little-forest/hasher` を宣言 | **不一致**。宣言パスと実位置が食い違うため、外部から `go get github.com/little-forest/hasher/core` は解決できない（サブディレクトリのモジュールはパスに `/src` を含む必要がある） |
| b | ロジックは `internal/*` へ落とす | `core` / `common` が公開パッケージ。`internal/` は存在しない | **不一致** |
| c | `internal/*` から `cmd` を import しない | 逆依存なし | 準拠 |
| d | `Execute(ctx context.Context) error` | `func Execute()`（戻り値なし・`cmd/root.go:39`） | **不一致** |
| e | `cmd` パッケージ内で `os.Exit` を呼ばない | `cmd/root.go:42` と `cmd/root.go:44` の 2 箇所で呼ぶ | **不一致** |
| f | `main.go` で `signal.NotifyContext` + `ExecuteContext` | シグナル処理なし（課題 B15） | **不一致** |
| g | 全サブコマンドが `RunE` | `version` だけ `Run`（`cmd/version.go:37`）。他 9 個は `RunE` | **一部不一致** |
| h | 全サブコマンドに `SilenceUsage: true` | 1 箇所も無い | **不一致** |
| i | 実装は `run<Command>` 関数に分離 | 済み。ただし `statusWrapper.RunE(...)` を挟み、実体のシグネチャが `(int, error)` という独自形 | 準拠だが形が違う |
| j | `init()` で `AddCommand` + フラグ定義 | 全ファイルで実施済み | 準拠 |
| k | フラグ名を `Flag_<command>_<Name>` 定数に | 定数化は全フラグで済み。ただし `Flag_root_Verbose` / `Flag_Update_ForceUpdate` / `Flag_DirDiff_showOnlyDifferences` / `Flag_Duplication_Source`（コマンド名は `duplicate`）と大文字小文字・語が揃っていない | **一部不一致** |
| l | フラグ値の受け取り方式をプロジェクト内で統一 | 全て方式 B（`cmd.Flags().GetBool(...)`）で統一されてはいる。ただし root の永続フラグを子コマンドから読む用途があり、go-cli はこの場合に方式 A（変数バインド）を推奨 | 統一済みだが推奨と異なる |
| m | 横断的前処理は root の `PersistentPreRunE` 1 箇所 | `PersistentPreRunE` 自体が存在しない | 該当なし（ロギングが対象外のため今回は不要） |
| n | ログは slog + tint、stderr、`--log-level` / `--log-format` | 独自の `ShowWarn` / `ShowError`（`common/common.go:51-61`）と ANSI 進捗表示 | **不一致（決定 14 により今回対象外）** |
| o | `version` は `RunE` + `versionString()` 1 箇所 + `cmd.OutOrStdout()` | `Run` + `fmt.Printf` 直書き。加えて未使用のデバッグ関数 `testSingleProcess` / `testMultiProgress` / `sleep` が同居（`cmd/version.go:47-159`、約 113 行） | **不一致** |
| p | ldflags の `-X` パスが実 import パスと一致 | 一致している（`github.com/little-forest/hasher/cmd.version`）。ただし goreleaser の実行が `workdir: src` + `-f ../.goreleaser.yml` という変則構成 | 準拠だが移設時に壊しやすい |
| q | `aqua.yaml` に go / goreleaser / task を固定 | `suzuki-shunsuke/github-comment` のみ | **不一致** |
| r | `Taskfile.yml` を置く | 存在しない | **不一致** |
| s | `.gitignore` | 存在（`/hasher`、`/src/dist`） | パス修正のみ |

### 既にあって使えるもの

- **フラグ定数はすでに全フラグで定義済み**（`cmd/*.go` の `const Flag_*`）。決定 10 は改名だけで済み、新規に定数を起こす必要は無い。
- **`statusWrapper`（`cmd/status_wrapper.go`、19 行）が終了コードの経路を 1 箇所に閉じている。** 廃止対象だが、「どのコマンドがどの経路を使っているか」はこのファイルと `return N, err` の grep だけで完全に把握できる。
- **`core` の純粋な処理と表示を伴う処理は既にきれいに分かれている。** `UpdateHashStrictly` / `CalcHash` / `GetHash` / `GetXattr` / `SetXattr` / `RemoveXattr` / `ClearXattr` / `HashAlg` / `Hash` / `UpdateError` は画面に何も書かない。表示を伴うのは `UpdateHash`（`core/hasher.go:40` の `ShowWarn`）、`ConcurrentUpdateHash`、`ListHash` / `ListHash2` のみ。**この境界がそのまま公開ライブラリの境界になる。**
- **`core` が `common` から使っているのは 11 種類の識別子だけ**（`WalkDir`×5、`ShowWarn`×5、`OpenFile`×3、`CountAllFiles`×3、`ShowErrorMsg`×2、`Mark_OK`×2、`IsSymbolicLink`×2、`Mark_Updated`、`Mark_Failed`、`Mark_Error`、`CheckFileType`）。うち純粋なのは `OpenFile` / `CheckFileType` / `IsSymbolicLink` / `WalkDir` のみ。
- **公開対象の関数が使う外部依存は `github.com/pkg/xattr` だけ。** `pkg/errors` を使う `errors.As` / `errors.Wrap` は `UpdateHash` と走査処理にしかなく、いずれも非公開側に残る。`morikuni/aec`（色）と `deckarep/golang-set`（差分）も非公開側のみ。
- `.gitignore` には既に `/hasher` があり、go-cli の `Taskfile.yml` が作るシンボリックリンクをそのまま無視できる。
- ldflags の注入先は `{module}/cmd` パッケージ。**`cmd/` はリポジトリルート直下へ移るだけなので、モジュールパスを変えない限り `-X` の文字列は現行のまま有効。**

### 現状のレイアウト

```
/                       ARCHITECTURE.md SPECS.md README.md CLAUDE.md
                        .goreleaser.yml aqua.yaml .pre-commit-config.yaml
  src/                  ← ここがモジュールルート
    go.mod              module github.com/little-forest/hasher
    .golangci.yaml
    main.go             cmd.Execute() を呼ぶだけ
    cmd/                11 ファイル
    core/               11 ファイル（公開パッケージ）
    common/             2 ファイル（公開パッケージ）
```

## 決定

| # | 論点 | 決定 | 理由 |
| --- | --- | --- | --- |
| 1 | ライブラリの提供形態 | 同一リポジトリ・単一モジュール。`src/` を廃してリポジトリルートをモジュールルートにする | ユーザー確認済み。現状は宣言モジュールパスと実位置が食い違い**外部から `go get` できない**（差分表 a）。ルートへ移せば go-cli の構成（差分表 a）と `go get` 可能性が同時に解決する |
| 2 | 公開パッケージ名 | `hashcore/`（リポジトリルート直下、`github.com/little-forest/hasher/hashcore`） | `core` のまま公開すると利用側の import 名が一般的すぎて衝突しやすい。`internal/` に入れられない（公開するため）ので `pkg/` を挟まずルート直下に置く |
| 3 | `hashcore` に入れるものの線引き | **画面に何も書かず、進捗通知も持たない処理だけ**。具体的には `Hash` / `HashAlg` / `UpdateError` / `CalcHash` / `UpdateHashStrictly` / `GetHash` / `GetXattr` / `SetXattr` / `RemoveXattr` / `ClearXattr` と `Xattr_*` 定数 | 要求が名指しした「hash 計算と xattr 属性への読み書き」がちょうどこの範囲。ここに限れば外部依存が `pkg/xattr` 1 本になり、色ライブラリも stderr への書き込みも利用側に押し付けずに済む |
| 4 | `common` の行き先 | 3 つに割る。純粋なファイル種別判定（`FileType` / `CheckFileType` / `OpenFile` / `IsDirectory` / `IsSymbolicLink` / `EnsureDirectory` / `EnsureRegularFile`）は **`hashcore` へ**。色・マーク・`Show*` は **`internal/term` へ**。走査・カウント・`CleanPath` は **`internal/fsutil` へ** | `hashcore` の関数は `OpenFile` を必要とする。これを `internal/` に置くと `hashcore` → `internal/term`（aec 依存）まで芋づるで引き込まれ、決定 3 の「依存 1 本」が崩れる |
| 5 | `Err_updateError` グローバル変数 | `hashcore` には入れず、唯一の利用者である `UpdateHash` と一緒に `internal/hasher` に置く | 課題 B17（並行実行時のデータ競合）を公開 API に載せないため。B17 自体は今回直さない（決定 15）が、**公開してしまうと後方互換の縛りが生まれる** |
| 6 | `Execute` のシグネチャ | `func Execute(ctx context.Context) error` にし、`main.go` で `signal.NotifyContext` を張って `rootCmd.ExecuteContext(ctx)` を呼ぶ。`os.Exit` は `main.go` だけ | go-cli コアルール（差分表 d / e / f）。`cmd` がライブラリのまま保たれ、cobra のコマンドをテストから呼べるようになる |
| 7 | `statusWrapper` の扱い | 廃止。全サブコマンドを `RunE: run<Command>`（`func(*cobra.Command, []string) error`）+ `SilenceUsage: true` に統一 | ユーザー確認済み。`SPEC-CLI-002` の 2 経路が 1 経路になり、失敗時に Usage 全文が流れなくなる |
| 8 | 「差分あり」で終了コード 1 を返す経路（`compare` の `return 1, nil`、`dirdiff` の走査失敗） | 無言のセンチネルエラーを返し、該当コマンドにだけ `SilenceErrors: true` を付ける | 決定 7 のまま素直に error を返すと `compare` が差分検出のたびに `Error: ...` を出す。`SPEC-CLI-103`（差分を終了コードで判定できる唯一のコマンド）の画面を変えずに終了コードだけ 1 にするには、この組み合わせしかない。`compare` は課題 B6 のとおり実エラーを返さないので、`SilenceErrors` で本物のエラーを握り潰す危険が無い |
| 9 | フラグ値の受け取り方式 | 方式 A（変数バインド、`BoolVar(&v, ...)`）に統一する | go-cli コアルール。`verbose` / `recursive` は root の永続フラグを子コマンドが読む用途で、go-cli が明示的に方式 A を推す形（cobra-guide §5「選択指針」）に当てはまる |
| 10 | フラグ定数の命名 | `Flag_<コマンド名を小文字で>_<Name>` に統一（`Flag_update_ForceUpdate` / `Flag_dirdiff_ShowOnlyDifferences` / `Flag_duplicate_Source` / `Flag_calc_NoShowPath` / `Flag_find_NoHash` / `Flag_listHash_Out`。`Flag_root_Verbose` / `Flag_root_Recursive` は現行のまま） | go-cli の `Flag_<command>_<Name>`（cobra-guide §4）。`list-hash` はハイフンが識別子に使えないため `listHash` とする。定数化の目的は grep 可能性なので、揺れが残ると効果が薄い |
| 11 | `rootCmd.Version` と `--version` | **設定しない。** `version` サブコマンドのみ維持する | cobra は `Version` が非空だと `--version` を自動追加し、`v` が空いていれば `-v` も奪う。hasher の `-v` は `verbose`（`SPEC-CLI-001`）なので、設定すると**ショートハンド重複で panic する**（cobra-guide §8 の落とし穴） |
| 12 | `cmd/version.go` の未使用デバッグ関数 | `testSingleProcess` / `testMultiProgress` / `sleep`（約 113 行）を削除する | `version` コマンドと無関係な進捗表示の手動確認コードで、呼び出し元はコメントアウトされている。1 コマンド 1 ファイルの原則（go-cli コアルール）に反する。履歴に残るので消して支障が無い |
| 13 | dot import | 全廃する（`. "github.com/little-forest/hasher/common"` × 9 箇所） | 決定 4 で `common` が 3 つに割れるため、どのみち全 import 文を書き直す。`// nolint:staticcheck` が 9 個消え、CLAUDE.md の「`nolint` は理由付きのときだけ」に近づく |
| 14 | ロギング | **今回対象外。** `internal/logging` は作らず、`--log-level` / `--log-format` も足さない。`ShowWarn` / `ShowError` と進捗表示は `internal/term` へ移すだけで実装は変えない | ユーザー確認済み。hasher の ANSI 進捗表示とマークは `SPEC-OUT-001`〜`005` で規定された**コマンドの成果物**であって診断ログではなく、slog 化すると UI の作り直しになる |
| 15 | 既知不具合の扱い | 構成変更が必然的に触る **B1 / B12 / B15 のみ**。B15 は「シグナル context の配線」まで行い、**下位層へ渡して中断を実効化することはしない**。B2 / B3 / B4 / B5 / B6 / B7 / B8 / B9 / B10 / B11 / B13 / B14 / B16 / B17 / B18 は据え置き | ユーザー確認済み。B15 について go-cli（cobra-guide §10）は「`cmd.Context()` を下位層に渡すなら `ExecuteContext` が必須」と言っており、逆に**渡さないなら配線だけで害は無い**。中断を実効化すると「途中まで属性を書いた状態をどこまで巻き戻すか」という別の設計判断が要るため切り離す |
| 16 | 課題 B1（`update` が失敗しても終了コード 0） | 修正する | `RunE` のシグネチャ変更で `runUpdateHash` は全面的に書き直しになり、`err` のシャドウイングと `// nolint:govet` 2 個は**書き直しの過程で必ず消える**。直さずに移すほうが不自然 |
| 17 | 課題 B12（`go.mod` の Go 宣言 1.21 と CI の 1.26.6 の乖離） | `go.mod` を CI に合わせる | `go.mod` はどのみち移設で書き直す。公開ライブラリになると宣言バージョンが**利用側の最低要件**になるため、実際にビルド検証しているバージョンと合わせておかないと利用者に嘘をつくことになる |
| 18 | ビルドツールチェーン | `Taskfile.yml` を新設。`aqua.yaml` に `golang/go` / `goreleaser/goreleaser` / `go-task/task` を追加。`src/.golangci.yaml` をルートへ移動。CI の `working-directory: ./src` と goreleaser の `workdir: src` / `-f ../.goreleaser.yml` を解消。CI の path filter を `src/**` から新レイアウトに直し、**`.goreleaser.yml` を含める** | go-cli `project-setup.md`「ビルドツールチェーン」。path filter は `src/**` が消滅するのでどのみち直す必要があり、そのついでに ARCHITECTURE.md §13 が指摘する「`.goreleaser.yml` の破壊が PR で検出されない」を塞げる |
| 19 | goreleaser の対象 OS | linux / darwin / windows のまま変えない | go-cli は「Windows はデフォルト除外、必要か確認」と言うが、既に windows 向けにリリースしている（`SPEC-OVERVIEW-002`）。**新規プロジェクト向けの既定値を理由に既存の成果物を減らすのは要求の範囲外** |
| 20 | 設計書の置き場所 | `docs/design/go-cli-alignment.md`（`docs/design/` を新設） | リポジトリ直下の `ARCHITECTURE.md` / `SPECS.md` は「常に現状を映す正典」であり、実装前の 1 回限りの設計書とは別種。混ぜると正典の位置づけが薄まる |

## 変更点

### 目標レイアウト

```
/                        go.mod  go.sum  main.go
                         .golangci.yaml  Taskfile.yml  aqua.yaml  .goreleaser.yml
                         ARCHITECTURE.md  SPECS.md  README.md  CLAUDE.md
  cmd/                   cobra コマンド定義（1 コマンド 1 ファイル）
    root.go              rootCmd / 永続フラグ / Execute(ctx) error
    calc.go clear.go compare.go dirdiff.go duplicate.go find.go
    list_hash.go show.go update.go version.go
    hasher_progress_notifier.go  stdio_progress_notifier.go
  hashcore/              ★公開ライブラリ（外部アプリが go get する）
    hash.go  hashalg.go  hasher.go  xattr.go  update_error.go  fileinfo.go
  internal/
    hasher/              ワーカープール・ListHash・HashStore・差分アルゴリズム
    term/                色定数・マーク・Show*・カーソル制御
    fsutil/              ディレクトリ走査・ファイル数カウント・CleanPath
  docs/design/           本設計書
```

依存の向き:

```
main ─> cmd ─> internal/hasher ─┬─> hashcore ─> pkg/xattr
              internal/term  <──┤
              internal/fsutil <─┘
                └─> hashcore, internal/term
```

`hashcore` は `internal/*` を一切 import しない（決定 4）。これにより外部利用者から見た依存は `github.com/pkg/xattr` 1 本になる。

### 1. `src/` の解体 ─ モジュールをリポジトリルートへ

`git mv src/* .` 相当。`go.mod` の `module` 行は変えない（決定 1）ため、`cmd` パッケージの import パス `github.com/little-forest/hasher/cmd` も変わらず、`.goreleaser.yml` の `-X` 4 行はそのまま有効（決定 18）。

追随が必要なファイル:

| ファイル | 変更 |
| --- | --- |
| `.github/workflows/ci.yaml` | `working-directory: ./src` を 3 箇所削除。path filter を `src/**` から `'**/*.go'` / `go.mod` / `go.sum` / `.goreleaser.yml` / `Taskfile.yml` / `aqua.yaml` / `.github/workflows/**` に差し替え |
| `.github/workflows/release.yml` | `workdir: src` と `-f ../.goreleaser.yml` を削除（`args: release --clean`） |
| `.gitignore` | `/src/dist` → `/dist` |
| `README.md` | ローカルビルド手順の `cd src` を削除し `task build` に。`cobra-cli add` の節も同様 |
| `src/.golangci.yaml` | ルートへ移動（内容は変更なし） |

### 2. `hashcore/` ─ 公開ライブラリの切り出し

`core` から**そのまま**移す（実装は変えない）:

| 移動元 | 移動先 | 中身 |
| --- | --- | --- |
| `core/hash.go` | `hashcore/hash.go` | `Hash` / `NewHash` / `NewHashFromString` / `String` / `Tsv` / `Json` / `HasSameHashValue` |
| `core/hashalg.go` | `hashcore/hashalg.go` | `HashAlg` / `NewDefaultHashAlg` / `NewHashAlg` / `NewHashAlgFromString` |
| `core/xattr.go` | `hashcore/xattr.go` | `GetXattr` / `SetXattr` / `RemoveXattr` / `ClearXattr` |
| `core/hasher.go` の一部 | `hashcore/hasher.go` | `Xattr_prefix` / `Xattr_size` / `Xattr_modifiedTime` / `Xattr_hashCheckedTime`、`UpdateHashStrictly` / `updateHashCheckedTime` / `CalcHash` / `GetHash` |
| `core/update_error.go` の一部 | `hashcore/update_error.go` | `UpdateError` 型のみ（`Err_updateError` は除く。決定 5） |
| `common/common.go` の一部 | `hashcore/fileinfo.go` | `FileType` と 4 定数、`CheckFileType` / `OpenFile` / `IsDirectory` / `IsSymbolicLink` / `EnsureDirectory` / `EnsureRegularFile`（決定 4） |

`hashcore` に残す import は `crypto` / `fmt` / `io` / `os` / `strconv` / `strings` / `time` と `github.com/pkg/xattr` のみ。

> ⚠️ `Xattr_prefix = "user.hasher"` はリテラルであり、パッケージ名の変更に**追随してはならない**（`SPEC-XATTR-001`〜`006`）。

### 3. `internal/hasher` / `internal/term` / `internal/fsutil`

| 移動元 | 移動先 |
| --- | --- |
| `core/hasher.go` の残り（`UpdateHash` / `UpdateTask` / `UpdateResult` / `ConcurrentUpdateHash` / `adjustNumOfWorkers` / `ListHash` / `ListHash2`） | `internal/hasher/` |
| `core/hashstore.go` `core/dirdiff.go` `core/filediff.go` `core/dirpair.go` `core/progress_notifier.go` | `internal/hasher/` |
| `core/update_error.go` の `Err_updateError` | `internal/hasher/` |
| `common/common.go` の色定数 `C_*` / マーク `Mark_*` / `ShowWarn` / `ShowError` / `ShowErrorMsg` / `ShowCursor` / `HideCursor` | `internal/term/` |
| `common/common.go` の `CleanPath` / `CountAllFiles`、`common/filewalker.go` 全体 | `internal/fsutil/` |
| `core/*_test.go` | 対象コードの移動先に追随（`hasher_test.go` は `hashcore` と `internal/hasher` に分かれる） |

dot import を全廃し（決定 13）、`hashcore` / `term` / `fsutil` として名前付きで import する。

### 4. `main.go` ─ シグナル context と終了コード

```go
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := cmd.Execute(ctx); err != nil {
		os.Exit(1)
	}
}
```

`crypto/sha1` のブランクインポートは維持する（これが無いと `hashAlg.Alg.Available()` が false になり `calc` が "no implementation" で失敗する）。エラー表示は足さない — cobra が stderr に出すため二重になる。

### 5. `cmd/root.go` ─ `Execute(ctx) error` と変数バインド

```go
var (
	verbose   bool
	recursive bool
)

func Execute(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, Flag_root_Verbose, "v", false, "verbose")
	rootCmd.PersistentFlags().BoolVarP(&recursive, Flag_root_Recursive, "r", false, "recursive")
}
```

`os.Exit` は消える（決定 6）。`PersistentPreRunE` は追加しない（決定 14）。`rootCmd.Version` は設定しない（決定 11）。

### 6. `cmd/status_wrapper.go` ─ 削除

各サブコマンドは次の形に揃う（決定 7 / 9 / 10）。

```go
const Flag_update_ForceUpdate = "force-update"

var updateForceUpdate bool

var updateCmd = &cobra.Command{
	Use:          "update",
	Short:        "Calculate file hash and save to extended attribute",
	RunE:         runUpdateHash,
	SilenceUsage: true,
}

func init() {
	updateCmd.Flags().BoolVarP(&updateForceUpdate, Flag_update_ForceUpdate, "f", false, "Force update")
	rootCmd.AddCommand(updateCmd)
}

func runUpdateHash(cmd *cobra.Command, args []string) error { ... }
```

`(int, error)` から `error` への読み替え表:

| 現行 | 移行後 |
| --- | --- |
| `return 0, nil` | `return nil` |
| `return 1, err` / `return -1, err` | `return err`（cobra が `Error: <msg>` を出し、`main` が 1 で終わる） |
| `return 1, nil`（`compare`:44,50、`dirdiff`:78） | `return errSilent`（決定 8）。該当コマンドに `SilenceErrors: true` を追加 |

`errSilent` は `cmd` パッケージ内のセンチネル:

```go
// errSilent reports failure through the exit code without printing anything.
// Commands returning it must set SilenceErrors: true.
var errSilent = errors.New("")
```

`runUpdateHash` は書き直しの過程で `err` のシャドウイングを解消し、`// nolint:govet` 2 個を削除する（決定 16）。ループ内の各パスの失敗を集約し、1 件でも失敗したら最後に error を返す。

### 7. `cmd/version.go`

```go
func versionString() string {
	return fmt.Sprintf("hasher version %s %s built from %s on %s\n", version, osArch, revision, date)
}

func runVersion(cmd *cobra.Command, args []string) error {
	fmt.Fprint(cmd.OutOrStdout(), versionString())
	return nil
}
```

`Run` → `RunE`、`SilenceUsage: true` を追加、`fmt.Printf` → `cmd.OutOrStdout()`（テストから捕捉可能にする）。**書式文字列は 1 文字も変えない**（`SPEC-CLI-110`）。デバッグ関数 3 本を削除（決定 12）。

### 8. `Taskfile.yml`（新規）と `aqua.yaml`

`task build` / `build-all` / `release-snapshot` / `test` / `lint` を定義し、`build` は `link-binary` を呼んでルートに `./hasher` のシンボリックリンクを張る（`.gitignore` に既にある）。`aqua.yaml` に `golang/go` / `goreleaser/goreleaser` / `go-task/task` を追加する（決定 18）。

## 触らないもの

| 触らないもの | 理由 |
| --- | --- |
| 全サブコマンドの出力書式・色・マーク（`SPEC-OUT-001`〜`005`） | 構成の付け替えであって UI の変更ではない。移行前後の出力が完全一致することが検証の柱（下記「検証」） |
| 進捗表示の実装（`hasher_progress_notifier.go` / `stdio_progress_notifier.go`） | 決定 14。`cmd` に置かれたままで go-cli の構成に反しない |
| 拡張属性の名前・値・書き込み順序 | `SPEC-XATTR-001`〜`006`。パッケージ構成と無関係であり、変えると既存のキャッシュが全て無効になる |
| TSV / JSON の書式（`SPEC-FMT-001`〜`004`） | 同上。`Json()` は課題 B18（エスケープ無し）を抱えたまま `hashcore` へ移す — 直すのは決定 15 の範囲外 |
| ハッシュアルゴリズムの選択（sha1 固定） | CLI フラグを足すのは機能追加であり要求の外 |
| goreleaser の対象 OS / アーキテクチャ / アーカイブ名 | 決定 19。`SPEC-OVERVIEW-002` が変わってしまう |
| `.pre-commit-config.yaml` の hook 構成 | ルートで動く設定のまま。`src/` の消滅で `golangci-lint-mod` の対象がむしろ素直になる |
| B2 / B3（シンボリックリンク判定） | 決定 15。`SPEC-LIMIT-001` の実挙動が変わるため、走査系の統合とセットで別途扱う |
| B4 / B14 / B17（並行処理まわり） | 決定 15。`hashcore` には入らず `internal/hasher` に留まるので、公開 API の互換性を気にせず後から直せる |
| B16（端末判定・`NO_COLOR`） | 決定 15。`SPEC-LIMIT-003` の変更になる |
| B10（README のダウンロード URL 不整合） | 決定 15。今回 README のビルド手順は直すが、リリース成果物名の不整合は別件 |

## フェーズ

| | 内容 | これだけで何が変わるか |
| --- | --- | --- |
| **P1** | `src/` を解体しモジュールをリポジトリルートへ（変更点 1）。パッケージ名・コードは一切変えない。CI / release ワークフロー / `.gitignore` / `.golangci.yaml` / README のビルド手順を追随させる | ユーザーから見た挙動は**何も変わらない**。開発者から見て `cd src` が不要になり、CI とリリースが新レイアウトで通る。**`.goreleaser.yml` の変更が PR で検証されるようになる** |
| **P2** | `hashcore/` と `internal/{hasher,term,fsutil}` へのパッケージ再編（変更点 2・3）。dot import 全廃、`Err_updateError` の移動、テストの追随 | 挙動は**変わらない**。`hashcore` が `pkg/xattr` だけに依存する自己完結パッケージになり、別アプリから `go get github.com/little-forest/hasher/hashcore` で hash 計算と xattr 読み書きが使えるようになる（要タグ付け） |
| **P3** | `cmd` 層の go-cli 準拠（変更点 4・5・6）。`Execute(ctx) error`、`main` で `signal.NotifyContext`、`statusWrapper` 廃止、`RunE` + `SilenceUsage` 統一、`errSilent`、B1 修正 | **失敗時に Usage 全文が出なくなる。** `update`（`-r` なし）が失敗時に終了コード 1 を返すようになる。`compare` / `dirdiff` の終了コードと画面は現状のまま。`Ctrl-C` / `SIGTERM` を受け取る配線が入る（中断が効くようになるわけではない） |
| **P4** | フラグ規約の統一（決定 9・10）と `cmd/version.go` の整理（変更点 7） | 挙動は**変わらない**。`--help` の表示もフラグ名・説明とも現状のまま。未使用コード 113 行が消え、`version` の出力がテストから捕捉できるようになる |
| **P5** | `Taskfile.yml` 新設、`aqua.yaml` へのツール追加、`go.mod` の Go 宣言を 1.26 系へ（決定 17・18） | 開発者が `task build` / `task test` / `task lint` を使えるようになり、ビルドに必要なツールが aqua で固定される。`hashcore` を使う側に対して必要な Go バージョンを正しく宣言できる |
| **P6（やらない）** | ロギングの slog 化（決定 14）、中断の実効化（決定 15）、B2 / B3 / B4 / B5 / B6 / B7 / B8 / B9 / B10 / B11 / B13 / B14 / B16 / B17 / B18 の修正、`hashcore` の v1 タグ付けと API ドキュメント | いずれもユーザーの決定で今回の範囲外。ただし **`hashcore` を v1 として公開する前には B18（`Json()` のエスケープ無し）を片付けること** — 公開 API に載せた後では書式を直せなくなる |

P1〜P5 は上から順に実施する。P1 を先に済ませないと import パスの書き換えが二度手間になり、P3 を P2 より先にやると `cmd` の import 文を 2 回書き直すことになる。

## 検証

### 自動

| 対象 | 確認すること |
| --- | --- |
| `go build ./...`（ルート） | 全フェーズ後に通る。`src/` 配下を参照する記述が残っていない |
| `go vet ./...` / `golangci-lint run`（ルート） | `govet` が全チェック有効の状態で通る。**`// nolint:govet` と `// nolint:staticcheck` が 1 つも残っていない**（決定 13・16） |
| `go test ./...` | 既存テスト（`dirdiff_test.go` の 6 シナリオ、`filediff_test.go`、`hasher_test.go`）が移動先パッケージで全て通る |
| `hashcore` の依存 | `go list -deps ./hashcore` の出力に `aec` / `golang-set` / `pkg/errors` / `internal/` が**現れない**こと（決定 3・4） |
| 拡張属性名（新規テスト） | `hashcore` のテストで `NewDefaultHashAlg().AttrName == "user.hasher.sha1"`、`Xattr_size == "user.hasher.size"` 等を固定。パッケージ改名が属性名に漏れていないことを機械的に押さえる |
| 課題 B1 の回帰テスト（新規、`cmd`） | `rootCmd.SetArgs([]string{"update", "<存在しないパス>"})` → `Execute` が非 nil を返す。CLAUDE.md の「B を直すときは回帰テストを足す」に従う |
| `version` の出力（新規、`cmd`） | `SetOut` で捕捉した文字列が `hasher version dev unknown built from dev on unknown\n` と一致（`SPEC-CLI-110` の書式が変わっていない） |
| ldflags 注入 | `task release-snapshot` 後のバイナリで `hasher version` が `dev` を返さない（`-X` パスが生きている。cobra-guide §8 の検証方法） |

### 手動

移行前の `main` でビルドしたバイナリを `/tmp/hasher-before` に退避してから始める。

1. `cd src && go build -o /tmp/hasher-before .` で移行前バイナリを作る。
2. テスト用ツリーを 1 つ用意する（xattr に対応した実ファイルシステム上。**RAM ディスク上では属性が保存されず検証にならない** — `ARCHITECTURE.md` §12 参照）。サブディレクトリ・同名ファイル・リネーム相当・重複内容のファイルを含める。
3. 移行前バイナリで次を実行し、stdout / stderr を別々にファイルへ保存する（`-v` 付きと無し両方）。
   `calc`、`update -r`、`show -r`、`list-hash`、`find -e`、`find -n`、`duplicate -s A -t B`、`dirdiff A B`、`dirdiff -d A B`、`clear -r`、`version`
4. 属性をすべて消した同一のツリーを再作成し、P5 まで完了したバイナリで同じ順序・同じ引数で実行し、同様に保存する。
5. **stdout / stderr を `diff` で突き合わせ、`update` の進捗表示（実行時刻に依存する部分）以外が完全一致すること**を確認する。ANSI エスケープを含めて一致させる（`cat -v` で比較する）。
6. 各コマンドの終了コード（`echo $?`）を移行前後で比較する。**`compare` で同一ファイル → 0、異なるファイル → 1、かつ両方とも標準出力・標準エラーに何も出ないこと**を個別に確認する（決定 8 が効いているか）。
7. 移行後バイナリで存在しないパスを `update` に渡し、**終了コードが 1 になり、かつ Usage 全文が出ないこと**を確認する（B1 と決定 7）。
8. 移行後バイナリで `dirdiff` を引数 1 個で実行し、`Error: accepts 2 arg(s), received 1` の後に **Usage 全文が出ないこと**を確認する。
9. 移行後バイナリで `hasher bogus` を実行し、`unknown command` と使い方の案内が**従来どおり出る**ことを確認する（`SilenceUsage` を root に付けていないこと）。
10. 大きめのツリーに対し `update -r -v` を実行して途中で `Ctrl-C` を押し、**プロセスが終了すること**を確認する（配線が入っただけなので、途中まで書かれた属性が残るのは現状どおりで正しい）。
11. `--help` および各サブコマンドの `--help` を移行前後で比較し、フラグ名・ショートハンド・説明文が一致することを確認する（決定 9・10 が表示に漏れていないこと）。`-v` が `verbose` のままで `--version` が生えていないことも見る（決定 11）。
12. 別ディレクトリに使い捨てのモジュールを作り、`go mod edit -replace` を使わずに `hashcore` を `require` できるか（タグ前は疑似バージョンで）確認する。`CalcHash` と `GetXattr` を呼ぶ数行のプログラムがビルド・実行できること。

## 見送った案

| 案 | 見送った理由 |
| --- | --- |
| ライブラリをマルチモジュール（リポジトリ内に 2 本目の `go.mod`）にする | ユーザーが単一モジュールを選択。タグが `hashcore/v1.2.3` 形式になり、ローカル開発で `replace` 運用が要る割に、利用者が 1 アプリの現時点では独立バージョニングの利点が薄い |
| ライブラリを別リポジトリへ切り出す | ユーザーが単一モジュールを選択。境界は最も明確だが、リポジトリ 2 本の同時開発になりこの計画の範囲が大きく広がる |
| ロギングを slog + tint に全面移行する | ユーザーが「今回対象外」を選択。`SPEC-OUT-001`〜`003` の進捗表示とマークは CLI の成果物であり、slog 化は UI の作り直しになる |
| 診断メッセージ（`ShowWarn` / `ShowError`）だけ slog 化する | 同上。`internal/term` への移動だけに留め、実装は現状維持とした |
| `exitError{code, err}` 型で終了コードを細分化する | ユーザーが `error` 一本化を選択。現状の終了コードは 0 / 1 の 2 値（`SPEC-CLI-002`）で、細分化の要求が出ていない。cobra-guide §7 も「2 値で足りるうちは導入しない」としている |
| `statusWrapper` を維持したまま cobra まわりだけ整える | ユーザーが廃止を選択。維持すると `cmd` 内の `os.Exit` が残り、go-cli のチェックリストを構造的に満たせない |
| 公開パッケージを `core` の名前のまま公開する | 利用側の import 名として一般的すぎ、他ライブラリと衝突しやすい。`hashcore` なら別名を付けずに済む |
| 公開パッケージを `pkg/hashcore` に置く | `pkg/` は Go 公式のレイアウト指針にも go-cli にも無い慣習で、階層が 1 つ深くなるだけ |
| `OpenFile` / `CheckFileType` を `internal/fsutil` に置き `hashcore` から import する | 同一モジュール内なので import 自体は合法だが、`internal/fsutil` は `CountAllFiles` 経由で `internal/term`（aec）に依存するため、公開ライブラリの依存グラフに色ライブラリと stderr への書き込みが混入する |
| `Json()`（課題 B18）を公開範囲から外す、あるいは今回直す | 決定 15 の範囲外。ただし公開 API に載る以上、v1 タグ前には片付ける必要がある（P6 に明記） |
| 走査の 3 実装（`ARCHITECTURE.md` §10）を今回統合する | `SPEC-LIMIT-001` の実挙動（B2 / B3）が変わる。「振る舞いを変えない再編」という今回の性質から外れる |
| 設計書を `ARCHITECTURE.md` に直接書き足す | `ARCHITECTURE.md` は「実際にそう動いている」ことを書く正典であり、未実装の計画を混ぜると信頼できなくなる。反映は実装完了後に行う |

## 実装後に更新する文書

**この設計の時点では反映しない。**

| 文書 | 追記する内容 |
| --- | --- |
| `SPECS.md` §2 `SPEC-CLI-002` | **改訂。** 終了コード 1 に至る経路が 1 本になり、「Usage 全文が出る」経路が消えることを反映。表の 3 行を書き直す |
| `SPECS.md` §6 `SPEC-OUT-005` | **改訂。** `Error: {メッセージ} + Usage 全文` の行を「Usage は出ない」に修正 |
| `SPECS.md` §3 `SPEC-CLI-102`（`update`） | **改訂。** `-r` なしで失敗した場合に終了コード 1 を返すようになったことを反映し、「⚠️ 既知の不具合（B1）」の注記を削除 |
| `SPECS.md` §3 `SPEC-CLI-103`（`compare`） | 終了コード・画面ともに変わらないことを確認のうえ、**変更が無いなら更新不要**。決定 8 の実装が意図どおりかを見てから判断する |
| `SPECS.md` §7 `SPEC-LIMIT-004` | **改訂。** シグナルを受け取る配線は入ったが、安全な中断には依然対応しないことを明記 |
| `SPECS.md` に新領域 `SPEC-LIB-001`（新規項目。採番は実装時） | **新設。** 公開パッケージ `github.com/little-forest/hasher/hashcore` が外部インタフェースになったことを記述する。公開範囲（決定 3 の関数・型・定数）、外部依存が `pkg/xattr` のみであること、`internal/*` は公開契約に含まれないこと、要求する Go バージョン |
| `SPECS.md` §1 `SPEC-OVERVIEW-002` | 成果物名・対象プラットフォームは変えない（決定 19）ため **更新不要**。ビルド手順の変更は README 側 |
| `ARCHITECTURE.md` §1 | **改訂。** 全体像の図と依存表を新レイアウト（`cmd` / `hashcore` / `internal/{hasher,term,fsutil}`）へ。`main.go` が `signal.NotifyContext` を張るようになった旨 |
| `ARCHITECTURE.md` §2 | **改訂。** パッケージ構成のファイル表を全面的に置き換え。`hashcore` を「公開ライブラリ」として位置づけ、`SPEC-LIB-001` を ID で参照する（振る舞いの事実は繰り返さない）。**なぜ `OpenFile` / `CheckFileType` を `hashcore` 側に置いたか**（決定 4 の理由）を残す |
| `ARCHITECTURE.md` §9 | **改訂。** `status_wrapper.go` による二経路の記述を削除し、`RunE` + `SilenceUsage` + `errSilent`（決定 8）の構造に差し替える |
| `ARCHITECTURE.md` §11 | **改訂。** B1 / B12 の行を削除。B15 を「配線済み。ただし下位層へ context を渡していないため中断は実効化されていない」に書き換え。B17 について「公開 API には露出していない」を追記 |
| `ARCHITECTURE.md` §12 | **改訂。** テストの所在（`core/` → `hashcore/` と `internal/hasher/`）と、新設した回帰テスト（属性名の固定、B1、`version` の書式）を反映 |
| `ARCHITECTURE.md` §13 | **改訂。** `src/` の消滅、`Taskfile.yml` の導入、CI の path filter が `.goreleaser.yml` を含むようになったこと（「リリース設定の破壊が PR で検出されない」という記述の削除） |
| `README.md` | ローカルビルド手順を `task build` に。`cobra-cli add` の節から `cd src` を削除。**`hashcore` をライブラリとして使う場合の import 例を追加**。Install 節の URL 不整合（B10）は今回触らない |
| `CLAUDE.md` | 構成図・ビルドコマンド・テストの所在を新レイアウトへ。「Known pitfalls」から B1 と B15 の項目を落とし、「`update` without `-r` swallows failures」「シンボリックリンク」等の残る項目は位置（ファイルパス）を更新。**`hashcore` は公開 API なので互換性を壊す変更は慎重に**という注意を追加 |
