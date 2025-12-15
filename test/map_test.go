package test

import (
	"sort"
	"testing"
)

func TestMap(t *testing.T) {
	a := []int{1, 3, 6, 10, 15, 21, 28, 36, 45, 55}
	x := 6

	r := sort.Search(len(a), func(i int) bool {
		t.Log("check index:", i, "value:", a[i])
		return a[i] >= x
	})
	// 检查结果是否正确
	t.Log("r:", r)
}
