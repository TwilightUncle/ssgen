package access_md

import (
	"github.com/TwilightUncle/ssgen/helpers/testing_helper"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/exp/slices"
)

func TestNewMdPaths(t *testing.T) {
	// データ生成
	baseDir := filepath.Join(os.TempDir(), "github.com/TwilightUncle/ssgen-test-"+testing_helper.MakeRandomStr(32))
	targetFileDatas := []testing_helper.TestFileData{
		{Path: filepath.Join(baseDir, "md_a.md"), Contents: []byte(filepath.Join(baseDir, "md_a.md"))},
		{Path: filepath.Join(baseDir, "dir_a", "md_b.txt"), Contents: []byte(filepath.Join(baseDir, "dir_a", "md_b.txt"))},
		{Path: filepath.Join(baseDir, "dir_b", "md_c.md"), Contents: []byte(filepath.Join(baseDir, "dir_b", "md_c.md"))},
	}
	nonTargetFileDatas := []testing_helper.TestFileData{
		// 拡張子が違うため対象外であること
		{Path: filepath.Join(baseDir, "md_a.csv"), Contents: []byte(filepath.Join(baseDir, "md_a.csv"))},
		{Path: filepath.Join(baseDir, "dir_a", "md_b.csv"), Contents: []byte(filepath.Join(baseDir, "dir_a", "md_b.csv"))},
		// skipするディレクトリであるため対象外であること
		{Path: filepath.Join(baseDir, "dir_c", "md_d.md"), Contents: []byte(filepath.Join(baseDir, "dir_c", "md_d.md"))},
		{Path: filepath.Join(baseDir, "dir_d", "md_d.md"), Contents: []byte(filepath.Join(baseDir, "dir_d", "md_d.md"))},
	}
	makeDatas := make([]testing_helper.TestFileData, len(targetFileDatas))
	copy(makeDatas, targetFileDatas)
	makeDatas = append(makeDatas, nonTargetFileDatas...)
	testing_helper.MakeTestFiles(baseDir, makeDatas, t)

	// 各メソッドのテスト
	testMdPaths, err := NewMdPaths(
		baseDir,
		[]string{filepath.Join(baseDir, "dir_c"), filepath.Join(baseDir, "dir_d")},
		[]string{".md", ".txt"},
	)
	if err != nil {
		t.Error(err)
	}
	// GetAll()
	if len(testMdPaths.GetAll()) != len(targetFileDatas) {
		t.Errorf("number of files: actual %d, want %d", len(testMdPaths.GetAll()), len(targetFileDatas))
	}
	for _, wantData := range targetFileDatas {
		if !slices.Contains(testMdPaths.GetAll(), wantData.Path) {
			t.Errorf("want exists: %s", wantData.Path)
		}
	}
	// Map()
	// MapFileContent()
}
