package md_parse

import (
	"regexp"

	"gopkg.in/yaml.v3"
)

const mETADATA_MATCH_PATTERN = `(?s)^---(.*?)---`

type MetaData struct {
	Title    string `yaml:"title"`
	PageName string
}

// ファイル内のうち、メタデータ部分を取得
func getMetaData(fileStr string, metaDataMatcher *regexp.Regexp) (MetaData, error) {
	// メタデータ読み取り
	var metaData MetaData
	if metaDataMatcher.MatchString(fileStr) {
		metaStr := metaDataMatcher.FindStringSubmatch(fileStr)[1]
		yamlParseErr := yaml.Unmarshal([]byte(metaStr), &metaData)
		if yamlParseErr != nil {
			return metaData, yamlParseErr
		}
	}
	return metaData, nil
}

// ファイル内のうち、マークダウン部分を取得
func getMd(fileStr string, metaDataMatcher *regexp.Regexp) []byte {
	mdStr := metaDataMatcher.ReplaceAllString(fileStr, "")
	return []byte(mdStr)
}

// ファイル内容を読み取り、メタデータとマークダウンを取得する
func ParseFileBytes(fileBytes []byte) (MetaData, []byte, error) {
	fileStr := string(fileBytes)
	exp := regexp.MustCompile(mETADATA_MATCH_PATTERN)
	metaData, ok := getMetaData(fileStr, exp)
	mdBytes := getMd(fileStr, exp)
	return metaData, mdBytes, ok
}
