package main

import (
	"quickreleasego/core"
	"fmt"
	"time"
)
func main() {
	args:=core.Getargs()
	release:=core.Release{Args:args}
	t1 := time.Now()
	fmt.Println(release.Call())
	fmt.Println(time.Since(t1))
}
