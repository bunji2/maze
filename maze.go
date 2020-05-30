package main

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"
)

// MazeData 迷路データ
type MazeData struct {
	Width  int     `json:"width"`  // 迷路の width。width>0
	Height int     `json:"height"` // 迷路の height。height>0
	Areas  []int   `json:"areas"`  // cid->aid, 各セルのエリアを格納する配列。0<=cid<w*h && 0<=aid<w*h
	Walls  []int   `json:"walls"`  // wid->0:壁あり or 1:壁なし, 各壁の有無を格納する配列。0<=wid<w*h*2-(w+h)
	Paths  [][]int `json:"paths"`
}

// NewMaze は新しい迷路データを作成する関数
func NewMaze(w, h int) (r MazeData) {
	if w < 1 || h < 1 {
		panic("w or h is too small")
	}
	areas := make([]int, w*h)
	for i := 0; i < w*h; i++ {
		areas[i] = i
	}
	r = MazeData{
		Width:  w,
		Height: h,
		Areas:  areas,
		Walls:  make([]int, w*h*2-(w+h)),
		Paths:  make([][]int, w*h),
	}
	return
}

// Build 迷路を作成する関数
func (m MazeData) Build() {
	numWalls := len(m.Walls)

	// 疑似乱数の初期化
	rand.Seed(time.Now().UnixNano())

	// ループ回数の上限
	counter := m.Width * m.Height * 10

	for !m.isAllSameArea() { // 一つのエリアに統合されるまで繰り返し
		// 壁をランダムに選択して、
		wid := rand.Intn(numWalls)
		// 壁を削除していく
		m.removeWall(wid)

		// 以下のコメントアウトをはずすと途中経過を確認できる
		//fmt.Println("wid =", wid)
		//m.Print(os.Stdout)

		counter--
		if counter < 0 {
			panic("too large loop times")
		}
	}
}

// removeWall は壁を削除する関数。
func (m MazeData) removeWall(wid int) {
	if m.Walls[wid] == 1 { // 1:壁なし
		// すでにその壁がないときは何もしない
		return
	}

	// 壁が仕切っているセルの組を特定
	cid1, cid2 := m.getSeparatedCells(wid)
	//fmt.Println(cid1, cid2)

	// すでに同じエリアのときは何もしない
	if m.Areas[cid1] == m.Areas[cid2] {
		return
	}

	// 二つのセルのエリアを統合
	m.updateAreas(cid1, cid2)
	m.Walls[wid] = 1 // 1:壁なし
	m.addPath(cid1, cid2)
}

// isAllSameArea はセルがすべて同じエリアであるかチェックする関数
func (m MazeData) isAllSameArea() bool {
	for i := 0; i < m.Width*m.Height; i++ {
		if m.Areas[i] != 0 { // 一つでも 0 以外のエリアがあるときは false を返す
			return false
		}
	}
	return true
}

// updateAreas は二つのセルのエリアを統合する関数。エリアIDの小さい方にそろえる。
func (m MazeData) updateAreas(cid1, cid2 int) {
	oldAid := m.Areas[cid1]
	newAid := m.Areas[cid2]

	// エリアID の小さい方のエリアにそろえる
	if newAid > oldAid {
		oldAid, newAid = newAid, oldAid
	}

	m.margeArea(oldAid, newAid)
}

// addPath は cid1 から cid2 へのパスを追加する関数
func (m MazeData) addPath(cid1, cid2 int) {
	if cid2 < cid1 {
		cid1, cid2 = cid2, cid1
	}
	m.Paths[cid1] = append(m.Paths[cid1], cid2)
}

// MakeDot は迷路の経路を dot 形式でファイルに保存する関数
func (m MazeData) MakeDot(filePath string) (err error) {
	var w *os.File
	w, err = os.Create(filePath)
	if err != nil {
		return
	}
	defer w.Close()

	m.DumpPaths(w)
	return
}

// DumpPaths は経路を dot 形式で出力する関数
func (m MazeData) DumpPaths(out io.Writer) {
	fmt.Fprintf(out, "graph maze%dx%d {\n", m.Width, m.Height)
	for cid := 0; cid < m.Width*m.Height; cid++ {
		switch cid {
		case 0:
			fmt.Fprintf(out, "\tn%d [ label = \"%d\", shape = \"doublecircle\", style=\"filled\", color = \"skyblue\", fillcolor = \"skyblue\" ];\n", cid, cid)
		case m.Width*m.Height - 1:
			fmt.Fprintf(out, "\tn%d [ label = \"%d\", shape = \"doublecircle\", style=\"filled\", color = \"green\", fillcolor = \"green\" ];\n", cid, cid)
		default:
			fmt.Fprintf(out, "\tn%d [ label = \"%d\" ];\n", cid, cid)
		}
	}
	for cid1, cid2s := range m.Paths {
		for _, cid2 := range cid2s {
			fmt.Fprintf(out, "\tn%d -- n%d;\n", cid1, cid2)
		}
	}
	fmt.Fprintf(out, "}\n")
}

// margeArea は旧エリアのすべてのセルを新しいエリアにマージする関数
func (m MazeData) margeArea(oldAid, newAid int) {
	size := m.Width * m.Height
	if oldAid < 0 || oldAid >= size || newAid < 0 || newAid >= size {
		panic("oldAid or newAid is abnormal")
	}
	for i := 0; i < size; i++ {
		if m.Areas[i] == oldAid {
			m.Areas[i] = newAid
		}
	}
}

// wid2cids は壁で仕切られたセルの組を返す関数
func (m MazeData) getSeparatedCells(wid int) (cid1, cid2 int) {
	w := m.Width
	h := m.Height
	if wid < 0 {
		panic("wid is minus!")
	}
	t1 := (w - 1) * h
	t2 := w*h*2 - (w + h)
	if wid < t1 { // 縦壁
		col := wid % (w - 1)
		row := wid / (w - 1)
		cid1 = col + row*w
		cid2 = cid1 + 1
	} else if wid < t2 { // 横壁
		cid1 = wid - t1
		cid2 = cid1 + w
	} else {
		panic("wid is too large!")
	}
	return
}

// Print は迷路をコンソールに表示する関数
func (m MazeData) Print(out io.Writer) {
	w := m.Width
	h := m.Height
	t1 := (w - 1) * h
	fmt.Fprint(out, "+")
	for i := 0; i < w; i++ {
		fmt.Fprint(out, "-+")
	}
	fmt.Fprintln(out, "")
	for i := 0; i < h; i++ {
		fmt.Fprint(out, "|")
		for j := 0; j < w-1; j++ {
			fmt.Fprint(out, " ")
			fmt.Fprint(out, m.wallStr(i*(w-1)+j))
		}
		fmt.Fprintln(out, " |")
		if i < h-1 {
			fmt.Fprint(out, "+")
			for j := 0; j < w; j++ {
				wid := t1 + j + i*w
				fmt.Fprintf(out, "%s+", m.wallStr(wid))
			}
			fmt.Fprintln(out, "")
		}
	}
	fmt.Fprint(out, "+")
	for i := 0; i < w; i++ {
		fmt.Fprint(out, "-+")
	}
	fmt.Fprintln(out, "")
}

func (m MazeData) wallStr(wid int) (r string) {
	if wid < 0 {
		panic("wid is minus!")
	}
	if m.Walls[wid] == 1 {
		r = " "
		return
	}
	w := m.Width
	h := m.Height
	t1 := (w - 1) * h
	t2 := w*h*2 - (w + h)
	if wid < t1 { // 縦壁
		r = "|"
	} else if wid < t2 { // 横壁
		r = "-"
	} else {
		panic("wid is too large!")
	}
	return

}
