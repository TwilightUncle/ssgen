package auto_link

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/TwilightUncle/ssgen/features/access_md"
)

type MdHeaderInfo struct {
	text     string
	pagename string
	id       string
}

type MdAllHeaaderInfo struct {
	// 見出しのヘッダ部分の内容をキーにしたマップ
	idMap map[string]MdHeaderInfo

	// 画面単位でグループ化した物
	pageGroup map[string][]MdHeaderInfo
}

// マークダウン中の見出しパターン
const mD_H_MATCH_PATTERN = `(?m)^#{1,6} +(.+)$`

// マークダウン中のリンク独自記法パターン
const mD_AUTO_LINK_MATCH_PATTERN = `\[\{(.*?)\}\]`

// 後方参照できなかった...
// https://github.com/google/re2/wiki/Syntax
const hTML_H_MATCH_PATTERN = `(?s)<(h\d)( ?[^<>]*)>(.*?)</(h\d)>`

// const hTML_H_MATCH_PATTERN = `(?s)<(h\d)( ?[^<>]*)>(.*?)</\1>`

// マークダウンの見出し情報を全て抽出
func getMdHeaderInfos(mdStr string, pagename string) []MdHeaderInfo {
	exp := regexp.MustCompile(mD_H_MATCH_PATTERN)

	var infos []MdHeaderInfo
	for _, match := range exp.FindAllStringSubmatch(mdStr, -1) {
		infos = append(infos, MdHeaderInfo{
			text:     match[1],
			pagename: pagename,
			id:       url.QueryEscape(match[1]),
		})
	}
	return infos
}

// 受け取った前の見出しの情報を取得
// 返却されるmapのkeyは見出しとして表示される内容
// 複数ページで同じ見出しが使用されている場合、処理において後に出現したものが優先される
// ※ソースの段階では順序不定?
func NewMdAllHeaaderInfo(mdPathes access_md.MdPaths) (MdAllHeaaderInfo, error) {
	groupedInfos, err := mdPathes.MapFileContent(func(path string, data []byte) interface{} {
		return getMdHeaderInfos(string(data), mdPathes.GetPageName(path))
	})

	if err != nil {
		return MdAllHeaaderInfo{}, err
	}

	idMap := map[string]MdHeaderInfo{}
	pageGroup := map[string][]MdHeaderInfo{}
	for _, infos := range groupedInfos {
		for _, hInfo := range infos.([]MdHeaderInfo) {
			idMap[hInfo.text] = hInfo
			pageGroup[hInfo.pagename] = append(pageGroup[hInfo.pagename], hInfo)
		}
	}
	return MdAllHeaaderInfo{idMap: idMap, pageGroup: pageGroup}, nil
}

// 文字列部分、パス部分、ID部分に文字列を分割
func splitStrPathId(target string) (string, string, string) {
	// サブマッチより抽出
	exp := regexp.MustCompile(`^((?:[^#|]+\|)?)([^#|]+)((?:#[^#|]*)?)$`)
	match := exp.FindStringSubmatch(target)

	// 解析対象でない並びの場合何も行わない
	// エラーとするべきか考える
	if match == nil {
		return "", "", ""
	}

	var str, path, id string
	strPartLen := len(match[1])
	idPartLen := len(match[3])

	if idPartLen == 0 {
		str = match[2]
		path = match[2]
	}

	// #のみの場合
	if idPartLen == 1 {
		str = match[2]
		id = match[2]
	}

	if idPartLen > 1 {
		id = match[3][1:]
		if strPartLen > 0 {
			path = match[2]
		} else {
			str = match[2]
		}
	}

	// 文字用のパートがあった場合、最終的に上書き('|'を除去の上)
	if strPartLen > 0 {
		str = match[1][0 : strPartLen-1]
	}
	return str, path, id
}

// pathに合致するmdの場所を検索し、返却
func searchFirstPath(path string, allHeaderInfos MdAllHeaaderInfo) string {
	// まず、完全一致のチェック
	for name := range allHeaderInfos.pageGroup {
		if name == path {
			return name
		}
	}
	// 末尾一致のチェック
	for name := range allHeaderInfos.pageGroup {
		exp := regexp.MustCompile(path + `$`)
		if exp.MatchString(name) {
			return name
		}
	}
	return ""
}

func existsIdInPagename(id string, pagename string, allHeaderInfos MdAllHeaaderInfo) bool {
	group := allHeaderInfos.pageGroup[pagename]
	exists := false
	for _, hInfo := range group {
		if hInfo.id == id {
			exists = true
			break
		}
	}
	return exists
}

// 置換する文字列を生成
func makeReplaceStr(match string, baseUrl string, allHeaderInfos MdAllHeaaderInfo, suffix string) (string, string, bool) {
	str, path, id := splitStrPathId(match)

	if str == "" {
		return match, "", false
	}

	if path == "" {
		hInfo, ok := allHeaderInfos.idMap[id]
		if !ok {
			return str, "", false
		}
		path = hInfo.pagename + suffix + "#" + hInfo.id
	} else {
		pagename := searchFirstPath(path, allHeaderInfos)
		if pagename == "" {
			return str, "", false
		}
		path = pagename + suffix

		if id != "" {
			if !existsIdInPagename(id, pagename, allHeaderInfos) {
				return str, "", false
			}
			path += "#" + id
		}
	}
	return str, baseUrl + "/" + path, true
}

// ページ名及び、独自記法から、マークダウンのリンクで置き換えたマークダウン文字列を返す
// リンクが見つからなかったものは二つ目の引数において内容をカンマ区切り文字列で返却
func MakeLink(baseUrl string, targetMdStr string, allHeaderInfos MdAllHeaaderInfo, suffix string) (string, string) {
	exp := regexp.MustCompile(mD_AUTO_LINK_MATCH_PATTERN)

	// リンク部分の置き換え実施
	replaced := exp.ReplaceAllStringFunc(targetMdStr, func(str string) string {
		str, path, _ := makeReplaceStr(exp.FindStringSubmatch(str)[1], baseUrl, allHeaderInfos, suffix)
		if path == "" {
			return str
		}
		return fmt.Sprintf("[%s](%s)", str, path)
	})

	// リンク指定されているにも関わらず、該当のリンクが存在しない物を検出
	notExistsLinks := []string{}
	for _, match := range exp.FindAllStringSubmatch(targetMdStr, -1) {
		if _, _, ok := makeReplaceStr(match[1], baseUrl, allHeaderInfos, suffix); !ok {
			notExistsLinks = append(notExistsLinks, match[1])
		}
	}
	return replaced, strings.Join(notExistsLinks, ",")
}

// 該当マークダウンの配置位置より、パンくずリストを作成
// 階層に該当する画面が存在しない場合はリンクではないただの文字列として表示する
func MakeBreadCrumbs(baseUrl string, pagename string, allHeaderInfos MdAllHeaaderInfo, suffix string) [][2]string {
	splited := strings.Split(pagename, "/")
	checkPath := make([]string, 0)
	result := make([][2]string, 0)

	for _, name := range splited {
		checkPath = append(checkPath, name)
		str, path, _ := makeReplaceStr(
			name+"|"+strings.Join(checkPath, "/"),
			baseUrl,
			allHeaderInfos,
			suffix,
		)
		result = append(result, [2]string{str, path})
	}
	return result
}

// HTMLの見出し要素にid属性を付与
func AddIdForHtmlH(html string) string {
	exp := regexp.MustCompile(hTML_H_MATCH_PATTERN)
	return exp.ReplaceAllStringFunc(html, func(str string) string {
		match := exp.FindStringSubmatch(str)
		return fmt.Sprintf(
			"<%s%s id=\"%s\">%s</%s>",
			match[1],
			match[2],
			url.QueryEscape(match[3]),
			match[3],
			match[4],
		)
	})
}
