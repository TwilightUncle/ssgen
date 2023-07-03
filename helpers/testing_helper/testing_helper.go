package testing_helper

import (
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type TestFileData struct {
	Path     string
	Contents []byte
}

// 指定長さの半角英数字文字列を生成
func MakeRandomStr(digit int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const maxIdx = len(letterBytes)

	rand.Seed(time.Now().UnixMicro())
	b := make([]byte, digit)
	for i := 0; i < digit; i++ {
		letterIds := rand.Intn(int(digit)) % maxIdx
		b[i] = letterBytes[letterIds]
	}
	return string(b)
}

// テスト用一時ファイルを作成する
// 終了処理の登録も含む
func MakeTestFiles(baseDir string, data []TestFileData, t *testing.T) {
	// 基本ディレクトリ作成
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		t.Errorf("create baseDir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(baseDir) })

	for _, fileData := range data {
		dirPath, _ := filepath.Split(fileData.Path)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			if err2 := os.MkdirAll(dirPath, 0755); err2 != nil {
				t.Errorf("create %s: %v", dirPath, err2)
			}
		}

		if err := os.WriteFile(fileData.Path, fileData.Contents, 0666); err != nil {
			t.Errorf("create %s: %v", fileData.Path, err)
		}
	}
}
