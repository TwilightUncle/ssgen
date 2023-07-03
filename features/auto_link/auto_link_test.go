package auto_link

import (
	"fmt"
	"github.com/TwilightUncle/ssgen/features/access_md"
	"github.com/TwilightUncle/ssgen/helpers/testing_helper"
	"os"
	"path/filepath"
	"testing"
)

// テスト用データ定義
const page1 = `
abc

// headers
# def
##  ghi
### jkl
####    mno
##### pqr
###### stu

// not headers
####### vwx
abc # y
#z
`

const page2 = `
# zyx
# wvu

// リンク情報の上書き確認用
# def
`

func makeTestFileData(t *testing.T) access_md.MdPaths {
	baseDir := filepath.Join(os.TempDir(), "github.com/TwilightUncle/ssgen-auto_link_test-"+testing_helper.MakeRandomStr(32))
	fileDatas := []testing_helper.TestFileData{
		{Path: filepath.Join(baseDir, "page1.md"), Contents: []byte(page1)},
		{Path: filepath.Join(baseDir, "sub", "page2.md"), Contents: []byte(page2)},
		// 無視される想定のフォルダ
		{Path: filepath.Join(baseDir, "layout", "_footer.md")},
	}
	testing_helper.MakeTestFiles(baseDir, fileDatas, t)

	paths, err := access_md.NewMdPaths(
		baseDir,
		[]string{filepath.Join(baseDir, "layout")},
		[]string{".md"},
	)
	if err != nil {
		t.Errorf("collect paths: %v", err)
	}
	return paths
}

func TestGetMdHeaderInfos(t *testing.T) {
	infos := getMdHeaderInfos(page1, "page1")

	// 抽出された数が正しいかテスト
	wantLen := 6
	if len(infos) != wantLen {
		t.Errorf("Want length is %d, but actual %d", len(infos), wantLen)
	}

	// 具体的な取得内容のテスト
	wantInfos := [...]MdHeaderInfo{
		{text: "def", pagename: "page1", id: "def"},
		{text: "ghi", pagename: "page1", id: "ghi"},
		{text: "jkl", pagename: "page1", id: "jkl"},
		{text: "mno", pagename: "page1", id: "mno"},
		{text: "pqr", pagename: "page1", id: "pqr"},
		{text: "stu", pagename: "page1", id: "stu"},
	}
	for i, wantInfo := range wantInfos {
		if infos[i] != wantInfo {
			t.Errorf(
				"Actual {text: %s, pagename: %s, id: %s}, want {text: %s, pagename: %s, id: %s}",
				infos[i].text,
				infos[i].pagename,
				infos[i].id,
				wantInfo.text,
				wantInfo.pagename,
				wantInfo.id,
			)
		}
	}
}

func TestGetAllFileMdHeaderInfos(t *testing.T) {
	// そもそも、mapについて呼び出し順序と記述した順序が守られるのか？
	allHInfos, err := NewMdAllHeaaderInfo(makeTestFileData(t))
	if err != nil {
		t.Errorf("error by getAllFileMdHeaderInfos: %v", err)
	}

	wantInfos := map[string]MdHeaderInfo{
		"ghi": {text: "ghi", pagename: "page1", id: "ghi"},
		"jkl": {text: "jkl", pagename: "page1", id: "jkl"},
		"mno": {text: "mno", pagename: "page1", id: "mno"},
		"pqr": {text: "pqr", pagename: "page1", id: "pqr"},
		"stu": {text: "stu", pagename: "page1", id: "stu"},
		"zyx": {text: "zyx", pagename: "sub/page2", id: "zyx"},
		"wvu": {text: "wvu", pagename: "sub/page2", id: "wvu"},
		// page2のdefで上書きされていること
		"def": {text: "def", pagename: "sub/page2", id: "def"},
	}
	for key, wantInfo := range wantInfos {
		if allHInfos.idMap[key] != wantInfo {
			t.Errorf(
				"Actual {text: %s, pagename: %s, id: %s}, want {text: %s, pagename: %s, id: %s}",
				allHInfos.idMap[key].text,
				allHInfos.idMap[key].pagename,
				allHInfos.idMap[key].id,
				wantInfo.text,
				wantInfo.pagename,
				wantInfo.id,
			)
		}
	}
}

func TestSplitStrPathId(t *testing.T) {
	str, path, id := splitStrPathId("")
	if str != "" || path != "" || id != "" {
		t.Errorf("Unexpected results by splitStrPathId: str=%s, path=%s, id=%s", str, path, id)
	}

	if str, path, id = splitStrPathId("str"); str != "str" || path != "str" || id != "" {
		t.Errorf("Unexpected results by splitStrPathId: str=%s, path=%s, id=%s", str, path, id)
	}

	if str, path, id = splitStrPathId("id#"); str != "id" || path != "" || id != "id" {
		t.Errorf("Unexpected results by splitStrPathId: str=%s, path=%s, id=%s", str, path, id)
	}

	if str, path, id = splitStrPathId("str|path"); str != "str" || path != "path" || id != "" {
		t.Errorf("Unexpected results by splitStrPathId: str=%s, path=%s, id=%s", str, path, id)
	}

	if str, path, id = splitStrPathId("str#id"); str != "str" || path != "" || id != "id" {
		t.Errorf("Unexpected results by splitStrPathId: str=%s, path=%s, id=%s", str, path, id)
	}

	if str, path, id = splitStrPathId("str|path/to/page#id"); str != "str" || path != "path/to/page" || id != "id" {
		t.Errorf("Unexpected results by splitStrPathId: str=%s, path=%s, id=%s", str, path, id)
	}

	if str, path, id = splitStrPathId("#nomatch"); str != "" || path != "" || id != "" {
		t.Errorf("Unexpected results by splitStrPathId: str=%s, path=%s, id=%s", str, path, id)
	}

	if str, path, id = splitStrPathId("|nomatch"); str != "" || path != "" || id != "" {
		t.Errorf("Unexpected results by splitStrPathId: str=%s, path=%s, id=%s", str, path, id)
	}

	if str, path, id = splitStrPathId("nomatch|"); str != "" || path != "" || id != "" {
		t.Errorf("Unexpected results by splitStrPathId: str=%s, path=%s, id=%s", str, path, id)
	}
}

func TestMakeLink(t *testing.T) {
	mdPaths := makeTestFileData(t)
	const baseUrl = "http://hostname.test/root"

	// エラーメッセージなし
	const makeLinkTarget1 = `
abc[{def#}]ghi[{jkl#}]
[{zyx|sub/page2#zyx}]
[{def|page1#def}]
[{def2|sub/page2#def}]
[{page2}]
`
	want1 := fmt.Sprintf(
		`
abc[def](%s/sub/page2#def)ghi[jkl](%s/page1#jkl)
[zyx](%s/sub/page2#zyx)
[def](%s/page1#def)
[def2](%s/sub/page2#def)
[page2](%s/sub/page2)
`,
		baseUrl,
		baseUrl,
		baseUrl,
		baseUrl,
		baseUrl,
		baseUrl,
	)
	allHInfos, err1 := NewMdAllHeaaderInfo(mdPaths)
	if err1 != nil {
		t.Error(err1)
	}
	actual1, notExists1 := MakeLink(baseUrl, makeLinkTarget1, allHInfos, "")
	if want1 != actual1 {
		t.Errorf("Actual [%s], want [%s]", actual1, want1)
	}
	if notExists1 != "" {
		t.Errorf("Actual [%s], want empty", notExists1)
	}

	// エラーメッセージ確認も含めて
	const makeLinkTarget2 = `
[{abc#}][{def#}][{ghij#}][{jkl#}]
[{zyx#}]
	`
	want2 := fmt.Sprintf(
		`
abc[def](%s/sub/page2#def)ghij[jkl](%s/page1#jkl)
[zyx](%s/sub/page2#zyx)
	`,
		baseUrl,
		baseUrl,
		baseUrl,
	)
	const wantNotExists = "abc#,ghij#"
	actual2, notExists2 := MakeLink(baseUrl, makeLinkTarget2, allHInfos, "")
	if want2 != actual2 {
		t.Errorf("Actual [%s], want [%s]", actual2, want2)
	}
	if notExists2 != wantNotExists {
		t.Errorf("Actual [%s], want [%s]", notExists2, wantNotExists)
	}
}

func TestMakeBreadCrumbs(t *testing.T) {
	mdPaths := makeTestFileData(t)
	const baseUrl = "http://hostname.test/root"

	allHInfos, err1 := NewMdAllHeaaderInfo(mdPaths)
	if err1 != nil {
		t.Error(err1)
	}

	breadCrumbs := MakeBreadCrumbs(baseUrl, "sub/page2", allHInfos, "")

	if breadCrumbs[0] != [2]string{"sub", ""} {
		t.Errorf("Actual [%+v], want [%+v]", breadCrumbs[0], [2]string{"sub", ""})
	}

	want := [2]string{"page2", baseUrl + "/sub/page2"}
	if breadCrumbs[1] != want {
		t.Errorf("Actual [%+v], want [%+v]", breadCrumbs[1], want)
	}
}
