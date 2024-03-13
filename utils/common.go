package utils

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func GenUUID() string {
	return uuid.New().String()[0:8]
}

// SlicesContains 检查 s 中是否包含所有的 sub
func SlicesContains(s, sub []string) bool {
	mapS := make(map[string]bool, len(s))
	for _, str := range s {
		mapS[str] = true
	}

	for _, substr := range sub {
		if !mapS[substr] {
			return false
		}
	}

	return true
}

// SlicesContainEqual 检查两个字符串切片 s1 和 s2 是否包含相同的元素，不考虑顺序
func SlicesContainEqual(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}

	mapS := make(map[string]int)

	for _, str := range s1 {
		mapS[str]++
	}

	for _, str := range s2 {
		if _, found := mapS[str]; !found || mapS[str] < 1 {
			return false
		}
		mapS[str]--
	}

	return true
}

func PrintMap[V fmt.Stringer](m map[string]V) string {
	ret := "{"
	for k, v := range m {
		ret += fmt.Sprintf(`"%s": %s,`, k, v)
	}
	ret = strings.TrimSuffix(ret, ",")
	ret += "}"

	return ret
}
