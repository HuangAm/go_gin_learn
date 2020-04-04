package test

import (
	"fmt"
	"testing"
)

func TestSliceLen(t *testing.T){
	s1 := []string{"张三", "李四", "王五"}
	fmt.Println(len(s1))
}
