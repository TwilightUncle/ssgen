package middleware

import (
	"fmt"

	"github.com/TwilightUncle/ssgen/features/access_md"
	"github.com/TwilightUncle/ssgen/features/auto_link"
	"github.com/TwilightUncle/ssgen/features/md_parse"
)

type Middleware func(metaData md_parse.MetaData, bytes []byte) (md_parse.MetaData, []byte, error)

type MiddlewareList struct {
	list []Middleware
}

// 実行順の先頭にミドルウェアを追加
func (mList *MiddlewareList) PushFront(fns ...Middleware) {
	mList.list = append(fns, mList.list...)
}

// 実行順の末尾にミドルウェアを追加
func (mList *MiddlewareList) Append(fns ...Middleware) {
	mList.list = append(mList.list, fns...)
}

// 登録されているすべてのミドルウェアを実行
func (mList *MiddlewareList) Apply(metaData md_parse.MetaData, bytes []byte) (md_parse.MetaData, []byte, error) {
	var err error
	for _, m := range mList.list {
		metaData, bytes, err = m(metaData, bytes)
		if err != nil {
			break
		}
	}
	return metaData, bytes, err
}

// リンク作成ミドルウェアを返す
func MakeMdAutoLink(baseUrl string, mdPathes access_md.MdPaths, suffix string) Middleware {
	// 見出しデータはあらかじめ収集の上、キャプチャしておく
	allHInfos, err := auto_link.NewMdAllHeaaderInfo(mdPathes)
	return func(metaData md_parse.MetaData, bytes []byte) (md_parse.MetaData, []byte, error) {
		if err != nil {
			return metaData, bytes, fmt.Errorf("Failed to make middleware 'MdAutoLink': %v", err)
		}
		mdStr, _ := auto_link.MakeLink(baseUrl, string(bytes), allHInfos, suffix)
		return metaData, []byte(mdStr), nil
	}
}

// h要素がリンクのアンカーとなるようIDを設定するミドルウェアを返す
func MakeHtmlAutoLinker() Middleware {
	return func(metaData md_parse.MetaData, bytes []byte) (md_parse.MetaData, []byte, error) {
		htmlStr := auto_link.AddIdForHtmlH(string(bytes))
		return metaData, []byte(htmlStr), nil
	}
}
