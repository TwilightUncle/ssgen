package md_parse

import (
	"reflect"
	"regexp"
	"testing"
)

const input1 = `---
title: "test"
---abc testtest`

const input2 = `abc testtest`

func TestGetMetaData(t *testing.T) {
	exp := regexp.MustCompile(mETADATA_MATCH_PATTERN)

	// メタデータ記述ありの場合
	metaYaml, err := getMetaData(input1, exp)
	want1 := MetaData{Title: "test"}
	if err != nil {
		t.Error(err)
	}
	if metaYaml != want1 {
		t.Errorf("Actual [%+v], want [%+v]", metaYaml, want1)
	}

	// メタデータの記述なし
	metaYaml, err = getMetaData(input2, exp)
	if err != nil {
		t.Error(err)
	}
	if !reflect.ValueOf(metaYaml).IsZero() {
		t.Errorf("Actual [%+v], want zeroValue", metaYaml)
	}
}

func TestGetMd(t *testing.T) {
	exp := regexp.MustCompile(mETADATA_MATCH_PATTERN)
	md1 := string(getMd(input1, exp))
	md2 := string(getMd(input2, exp))
	if md1 != md2 {
		t.Errorf("Want same md1 and md2")
	}
}
