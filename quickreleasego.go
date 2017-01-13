package main

import (
	"quickreleasego/core"
)
func main() {
	args:=core.Getargs()
	release:=core.Release{Args:args}
	release.Call()
}
