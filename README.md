# github.com/TwilightUncle/ssgen
静的サイト生成。  
主にマークダウンをHTMLに変換。  

## 利用モジュール
- github.com/gin-gonic/gin
- github.com/russross/blackfriday

## パッケージダウンロード
```sh
go mod tidy
```

## サイト内リンク用のMD記法

```
1. <str>を表示し、<str>に該当する画面(マークダウン)のリンクを生成
[{str}]

2. <str>を表示し、<str>に合致する見出し要素へのリンクを生成
[{str#}]

3. <str>を表示し、<path>に合致する画面(マークダウン)のリンクを生成
[{str|path}]

4. <str>を表示し、<id>に合致する見出し要素へのリンクを生成
[{str#id}]

5. <str>を表示し、<path>に該当する画面(マークダウン)中の<id>の見出しへのリンクを生成
[{str|path#id}]
```

## オプション

- -preview - サーバーを起動し、動的サーバーとしてレンダリングを行う
- -preview-static - 静的サイトの構築と出力を行ったうえで、静的ファイルを返すだけのサーバーを起動する
- オプションなし - 静的サイトの構築と出力のみ
