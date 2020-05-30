// 迷路を作るプログラム

package main

import (
	"fmt"
	"os"
	"strconv"
)

const (
	usageFmt = "Usage: %s width height\n"
)

func main() {
	os.Exit(run())
}

func run() int {
	// 引数の数をチェック
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, usageFmt, os.Args[0])
		return 1
	}

	var w, h int
	var err error
	// 迷路の width/height を引数から取得
	w, err = strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "width is not number (%s)\n", os.Args[1])
		return 2
	}
	h, err = strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "height is not number (%s)\n", os.Args[2])
		return 2
	}

	// 迷路データの作成
	m := NewMaze(w, h)

	m.Build()

	// 迷路を表示
	m.Print(os.Stdout)

	m.MakeDot("maze.dot")

	return 0
}
