package ssgen

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/TwilightUncle/ssgen/features/access_md"
	"github.com/TwilightUncle/ssgen/features/auto_link"
	"github.com/TwilightUncle/ssgen/features/md_parse"
	"github.com/TwilightUncle/ssgen/middleware"

	"github.com/gin-gonic/gin"
	"github.com/russross/blackfriday"
)

type LayoutBuilder func(metaData md_parse.MetaData, convertedHtml template.HTML) gin.H

type Core struct {
	// htmlへ返還前のマークダウンに対して行う処理リスト
	MdMiddlewareList middleware.MiddlewareList
	// 変換後のhtmlに対して行う処理リスト
	HtmlMiddlewareList middleware.MiddlewareList
	MdPaths            access_md.MdPaths
	LayoutBuilder      LayoutBuilder

	BaseUrl          string
	AssetsPath       string
	TemplateDir      string
	TemplateHtmlName string
	OutputDir        string
	UrlSuffix        string

	initialized bool
	previewFlag int
}

var c Core

const (
	buildOnly = iota
	preview
	previewStatic
)

// デフォルトの設定。引数にはマークダウンを格納しているディレクトリと、cssやjs等の静的ファイルを格納するPATH(ディレクトリ,URL共用)を指定
func Default(baseUrl string, mdBaseDir string, assetsPath string, templateDir string, outputDir string) error {
	return Initialize(func(core *Core) error {
		suffix := ""
		if c.previewFlag != preview {
			suffix = ".html"
		}

		var err error
		mdLayoutDir := mdBaseDir + "/layout"
		// マークダウンパス取得
		core.MdPaths, err = access_md.NewMdPaths(
			mdBaseDir,
			[]string{mdLayoutDir},
			[]string{".md"},
		)
		if err != nil {
			return err
		}

		// ミドルウェア登録
		core.MdMiddlewareList.Append(middleware.MakeMdAutoLink(baseUrl, c.MdPaths, suffix))
		core.HtmlMiddlewareList.Append(middleware.MakeHtmlAutoLinker())

		// その他
		core.BaseUrl = baseUrl
		core.AssetsPath = assetsPath
		core.TemplateDir = templateDir
		core.TemplateHtmlName = "index.html"
		core.OutputDir = outputDir
		core.UrlSuffix = suffix
		core.LayoutBuilder, err = MakeDefaultLayoutBuilder(baseUrl, assetsPath, mdLayoutDir)
		return err
	})
}

// 任意の処理による初期化
func Initialize(fn func(core *Core) error) error {
	// 初期化済みの場合エラー
	if c.initialized {
		return fmt.Errorf("already initialized. cannot be call")
	}

	previewFlag := flag.Bool("preview", false, "run preview server")
	previewStaticFlag := flag.Bool("preview-static", false, "run preview static")
	flag.Parse()

	switch {
	case *previewStaticFlag:
		c.previewFlag = previewStatic
	case *previewFlag:
		c.previewFlag = preview
	default:
		c.previewFlag = buildOnly
	}

	if err := fn(&c); err != nil {
		return err
	}

	c.initialized = true
	return nil
}

// テンプレートの組み上げ
func MakeDefaultLayoutBuilder(baseUrl string, assetsPath string, mdLayoutDir string) (LayoutBuilder, error) {
	layoutComponentPathes := map[string]string{
		"header":  mdLayoutDir + "/_header.md",
		"sidebar": mdLayoutDir + "/_sidebar.md",
		"footer":  mdLayoutDir + "/_footer.md",
	}

	// html部品のマークダウンをhtml化
	var err error
	ginH := gin.H{}
	for key, path := range layoutComponentPathes {
		// マークダウンのバイト列取得
		bytes, readErr := os.ReadFile(path)
		if readErr != nil {
			err = readErr
			continue
		}
		// マークダウンにミドルウェア適用
		if _, bytes, err = c.MdMiddlewareList.Apply(md_parse.MetaData{}, bytes); err != nil {
			fmt.Println(err)
			continue
		}
		htmlBytes := blackfriday.MarkdownCommon(bytes)
		ginH[key] = template.HTML(htmlBytes)
	}

	// そのほか、htmlへ埋め込む変数
	ginH["base_url"] = baseUrl
	ginH["assets_path"] = baseUrl + "/" + assetsPath

	allHInfos, _ := auto_link.NewMdAllHeaaderInfo(c.MdPaths)

	// 関数構築
	return func(metaData md_parse.MetaData, convertedHtml template.HTML) gin.H {
		ginH["title"] = metaData.Title
		ginH["overview"] = metaData.Overview
		ginH["breadcrumbs"] = auto_link.MakeBreadCrumbs(baseUrl, metaData.PageName, allHInfos, c.UrlSuffix)
		for i := 1; i <= 6; i++ {
			ginH["idlinks"+strconv.Itoa(i)] = auto_link.MakePageInnerPaths(baseUrl, metaData.PageName, i, allHInfos, c.UrlSuffix)
		}
		ginH["content"] = convertedHtml
		return ginH
	}, err
}

// マークダウンをHTMLへ変換
func convertToHtml(mdFilePath string) (md_parse.MetaData, []byte, error) {
	// マークダウンのバイト列取得
	bytes, readErr := os.ReadFile(mdFilePath)
	if readErr != nil {
		return md_parse.MetaData{}, []byte{}, readErr
	}
	metaData, mdBytes, err := md_parse.ParseFileBytes(bytes)
	if err != nil {
		return metaData, []byte{}, err
	}

	// マークダウンにミドルウェア適用
	if metaData, mdBytes, err = c.MdMiddlewareList.Apply(metaData, mdBytes); err != nil {
		return metaData, []byte{}, err
	}

	metaData.PageName = c.MdPaths.GetPageName(mdFilePath)

	// HTMLに変換の上ミドルウェア適用
	return c.HtmlMiddlewareList.Apply(
		metaData,
		blackfriday.MarkdownCommon(mdBytes),
	)
}

// preview の場合はプレビュー用のサーバーを起動する
func Build() error {
	switch c.previewFlag {
	case preview:
		return RunPreviewServer()
	case previewStatic:
		return RunPreviewStatic()
	}
	return BuildStaticSite()
}

// 静的サイトを出力
func BuildStaticSite() error {
	if !c.initialized {
		return fmt.Errorf("Prease Call the function 'Default' or 'Initialize' beforehand.")
	}

	// 既に存在している場合、出力先ディレクトリを作り直す
	_, err := os.Stat(c.OutputDir)
	if !os.IsNotExist(err) {
		if err = os.RemoveAll(c.OutputDir); err != nil {
			return err
		}
	}

	if err = copyAssetsAll(); err != nil {
		return err
	}
	return outputHtmlAll()
}

// アセッツのコピー
func copyAssetsAll() error {
	assetsPaths, err := access_md.NewMdPaths(
		c.AssetsPath,
		[]string{},
		[]string{".css", ".js"},
	)

	if err != nil {
		return err
	}

	// outputDir/assetsへファイルをコピー
	for _, path := range assetsPaths.GetAll() {
		if err := copyAssets(assetsPaths, path); err != nil {
			return err
		}
	}
	return nil
}

// 一つのファイルをコピー
func copyAssets(assetsPaths access_md.MdPaths, filePath string) error {
	assetsOutputDir := c.OutputDir + "/" + c.AssetsPath + "/"
	if err := os.MkdirAll(assetsOutputDir+assetsPaths.GetPageDir(filePath), 0777); err != nil {
		return err
	}

	relPath, _ := filepath.Rel(assetsPaths.GetBaseDirPath(), filePath)
	dest, err := os.Create(assetsOutputDir + relPath)
	if err != nil {
		return err
	}

	src, err := os.Open(filePath)
	if err != nil {
		return err
	}

	io.Copy(dest, src)
	return nil
}

// 配置されたmdよりすべてのHTMLファイルを出力する
func outputHtmlAll() error {
	// htmlテンプレート取得
	t, err := template.ParseFiles(c.TemplateDir + "/" + c.TemplateHtmlName)
	if err != nil {
		return err
	}

	for _, mdPath := range c.MdPaths.GetAll() {
		if err := outputHtml(t, mdPath); err != nil {
			return err
		}
	}
	return nil
}

// HTMLファイルを出力
func outputHtml(t *template.Template, mdPath string) error {
	metaData, htmlBytes, err := convertToHtml(mdPath)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err = t.Execute(&buf, c.LayoutBuilder(metaData, template.HTML(htmlBytes))); err != nil {
		return err
	}

	if err = os.MkdirAll(c.OutputDir+"/"+c.MdPaths.GetPageDir(mdPath), 0777); err != nil {
		return err
	}

	if err = os.WriteFile(c.OutputDir+"/"+c.MdPaths.GetPageName(mdPath)+".html", buf.Bytes(), 0777); err != nil {
		return err
	}
	return nil
}

// 静的ファイル生成の上、プレビュー
func RunPreviewStatic() error {
	if !c.initialized {
		return fmt.Errorf("Prease Call the function 'Default' or 'Initialize' beforehand.")
	}

	if err := BuildStaticSite(); err != nil {
		return err
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.Static("/", c.OutputDir)

	// run
	router.Run(":8080")
	return nil
}

// プレビュー用サーバー起動
func RunPreviewServer() error {
	if !c.initialized {
		return fmt.Errorf("Prease Call the function 'Default' or 'Initialize' beforehand.")
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.Static("/"+c.AssetsPath, c.AssetsPath)
	router.LoadHTMLGlob(c.TemplateDir + "/*.html")

	// マークダウンのファイル配置よりroute作成
	for _, mdPath := range c.MdPaths.GetAll() {
		handler := makePreviewHandler(mdPath)
		pagename := c.MdPaths.GetPageName(mdPath)
		router.GET("/"+pagename, handler)
		if pagename == "index" {
			router.GET("/", handler)
		}
	}

	// run
	router.Run(":8080")
	return nil
}

// ルーティングのハンドラ作成
func makePreviewHandler(mdPath string) func(con *gin.Context) {
	metaData, htmlBytes, err := convertToHtml(mdPath)
	return func(con *gin.Context) {
		if err != nil {
			fmt.Println(err)
			con.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		con.HTML(
			http.StatusOK,
			c.TemplateHtmlName,
			c.LayoutBuilder(metaData, template.HTML(htmlBytes)),
		)
	}
}
