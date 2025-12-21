package handler

import (
	"regexp"
)

var censorRegex = regexp.MustCompile(`!!(.*?)!!`)

// CensorReplacement: 伏せ字の置換後フォーマット
const CensorReplacement = "!!■■■!!"

// CensorContent : 文字列内の !!text!! を !!■■■!! に置換
func CensorContent(input string) string {
	return censorRegex.ReplaceAllString(input, CensorReplacement)
}

func ApplyCensorIfNeed(role string, input string) string {
    if role == "manager" {
        return input
    }

    return CensorContent(input)
}