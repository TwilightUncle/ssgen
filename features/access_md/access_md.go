package access_md

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/exp/slices"
)

// NewMdPaths関数により生成すること
type MdPaths struct {
	baseDir    string
	skipDirs   []string
	targetExts []string
	paths      []string
}

func NewMdPaths(baseDir string, skipDirs []string, targetExts []string) (MdPaths, error) {
	paths, err := getMdPaths(baseDir, skipDirs, targetExts)
	if err != nil {
		return MdPaths{}, err
	}

	return MdPaths{
		baseDir:    baseDir,
		skipDirs:   skipDirs,
		targetExts: targetExts,
		paths:      paths,
	}, nil
}

func (p *MdPaths) GetBaseDirPath() string {
	return p.baseDir
}

// 保持している全てのファイルパスを取得
func (p *MdPaths) GetAll() []string {
	return p.paths
}

// pathからふぁいる拡張子と、baseDir魔での文字列を除去する
func (p *MdPaths) GetPageName(path string) string {
	relPath, _ := filepath.Rel(p.GetBaseDirPath(), path)
	return strings.Split(relPath, ".")[0]
}

// ファイルを配置するディレクトリ名を生成
func (p *MdPaths) GetPageDir(path string) string {
	relPath, _ := filepath.Rel(p.GetBaseDirPath(), path)
	splited := strings.Split(relPath, "/")
	splited[len(splited)-1] = ""
	return strings.Join(splited, "/")
}

// 全要素に関数を適用した結果スライスを返却
func (p *MdPaths) Map(fn func(fPath string) interface{}) []interface{} {
	result := []interface{}{}
	for _, path := range p.paths {
		result = append(result, fn(path))
	}
	return result
}

// 全要素のファイル読み込み結果に対して処理を適用し、結果スライスを返却
func (p *MdPaths) MapFileContent(fn func(fPath string, fData []byte) interface{}) ([]interface{}, error) {
	result := []interface{}{}
	for _, path := range p.paths {
		bytes, err := os.ReadFile(path)
		if err != nil {
			return result, err
		}
		result = append(result, fn(path, bytes))
	}
	return result, nil
}

// baseDir以下の階層に存在するマークダウンファイルパスを全て取得
func getMdPaths(baseDir string, skipDirs []string, targetExts []string) ([]string, error) {
	// 探索対象外(再帰で入ったとき想定)
	if slices.Contains(skipDirs, baseDir) {
		return []string{}, nil
	}

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, fmt.Errorf("read dir: %w", err)
	}

	paths := []string{}
	for _, file := range entries {
		name := filepath.Join(baseDir, file.Name())

		// ディレクトリの場合再帰
		if file.IsDir() {
			recur_paths, err := getMdPaths(name, skipDirs, targetExts)
			if err != nil {
				return nil, err
			}
			paths = append(paths, recur_paths...)
			continue
		}

		// 対象拡張子ファイル以外除外
		if !slices.Contains(targetExts, filepath.Ext(name)) {
			continue
		}
		paths = append(paths, name)
	}
	return paths, nil
}
