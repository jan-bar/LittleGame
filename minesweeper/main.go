package main

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"image"
	"image/color"
	"image/png"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func main() {
	rand.Seed(time.Now().Unix())

	m := &mine{h: 16, w: 30, mineCnt: 99}
	err := m.loadResources()
	if err != nil {
		log.Fatal(err)
	}

	m.data = make([][]*grid, m.h)
	for i := 0; i < m.h; i++ {
		m.data[i] = make([]*grid, m.w)
	}
	m.initData() // 开局初始数据

	ebiten.SetWindowTitle("Mine Sweeping")
	if err = ebiten.RunGame(m); err != nil {
		log.Fatal(err)
	}
}

const (
	gridHW = 16 // 格子宽高
)

type (
	mine struct {
		// 雷区宽高
		h, w int
		// 雷区格子数据
		data [][]*grid
		// 开局状态
		start bool
		// 剩余雷数
		mineCnt int
		// 开始时间
		timeStart time.Time
		// 计时器
		timeCnt int
		// 计时器最右侧数字Y坐标
		timeY float64
		// 界面宽高
		gridW, gridH int
		// 显示哪个笑脸
		faceNum int
		// 笑脸左上角坐标(用于显示或计算点击笑脸)
		faceX, faceY float64
		// 界面图片,数字图片,笑脸图片
		img, num, face []*ebiten.Image
	}

	grid struct {
		data  int // 格子数据
		state int // 状态
	}
)

var aroundPos = [][]int{
	{-1, 1, 0, 0, -1, -1, 1, 1},
	{0, 0, -1, 1, -1, 1, -1, 1},
}

func (m *mine) around(h, w int, f func(h, w int)) {
	var i, nh, nw int
	for ; i < 8; i++ {
		nh, nw = h+aroundPos[0][i], w+aroundPos[1][i]
		if nh >= 0 && nh < m.h && nw >= 0 && nw < m.w {
			f(nh, nw) // [h,w]周围合法8个位置
		}
	}
}

func (m *mine) initData() {
	if m.start {
		return
	}

	var i, j int
	for ; i < m.h; i++ {
		for j = 0; j < m.w; j++ {
			if d := m.data[i][j]; d == nil {
				m.data[i][j] = new(grid)
			} else {
				d.data, d.state = 0, 0
			}
		}
	}

	cnt := 0
	for cnt < m.mineCnt {
		i, j = rand.Intn(m.h), rand.Intn(m.w)
		if d := m.data[i][j]; d.data != 10 {
			cnt++
			d.data = 10
		}
	}

	for i = 0; i < m.h; i++ {
		for j = 0; j < m.w; j++ {
			if d := m.data[i][j]; d.data != 10 {
				m.around(i, j, func(h, w int) {
					if m.data[h][w].data == 10 {
						d.data++
					}
				})
			}
		}
	}

	m.start = true
	m.timeStart = time.Now()
	m.gridW, m.gridH = m.w*gridHW+6, (m.h+2)*gridHW+6
	m.faceX, m.faceY = float64(m.gridW/2)-18, 4
	m.timeY = float64(m.gridW - 18)
	ebiten.SetWindowSize(m.gridW, m.gridH)
}

func (m *mine) Update() error {
	if m.timeCnt < 999 {
		m.timeCnt = int(time.Since(m.timeStart).Seconds())
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		m.faceNum = 1
		x, y := ebiten.CursorPosition()
		w, h := (x-3)/gridHW, y/gridHW-2
		m.data[h][w].state = 1
	} else if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		m.faceNum = 0
	}

	return nil
}

var backgroundColor = color.RGBA{R: 0xc0, G: 0xc0, B: 0xc0, A: 0xff}

func (m *mine) Draw(screen *ebiten.Image) {
	screen.Fill(backgroundColor)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(3, 2*gridHW)
	for i := 0; i < m.h; i++ {
		for j := 0; j < m.w; j++ {
			switch d := m.data[i][j]; d.state {
			case 0:
				screen.DrawImage(m.img[15], op)
			case 1:
				screen.DrawImage(m.img[0], op)
			default:
				screen.DrawImage(m.img[d.data], op)
			}
			op.GeoM.Translate(gridHW, 0)
		}
		op.GeoM.Translate(0, gridHW)
		op.GeoM.SetElement(0, 2, 3)
	}

	op.GeoM.Reset() // 显示雷数
	op.GeoM.Translate(5, 5)
	for _, v := range []int{(m.mineCnt / 100) % 10, (m.mineCnt / 10) % 10, m.mineCnt % 10} {
		screen.DrawImage(m.num[v], op)
		op.GeoM.Translate(13, 0)
	}

	op.GeoM.Reset() // 显示时间
	op.GeoM.Translate(m.timeY, 5)
	for _, v := range []int{m.timeCnt % 10, (m.timeCnt / 10) % 10, (m.timeCnt / 100) % 10} {
		screen.DrawImage(m.num[v], op)
		op.GeoM.Translate(-13, 0)
	}

	op.GeoM.Reset() // 显示笑脸
	op.GeoM.Translate(m.faceX, m.faceY)
	screen.DrawImage(m.face[m.faceNum], op)
}

func (m *mine) Layout(_, _ int) (int, int) {
	return m.gridW, m.gridH
}

//go:embed mine.zip
var mineZip []byte

func (m *mine) loadResources() error {
	data := bytes.NewReader(mineZip)
	zr, err := zip.NewReader(data, data.Size())
	if err != nil {
		return err
	}

	subImg := func(img image.Image, x, y, num int) []*ebiten.Image {
		var (
			ei = ebiten.NewImageFromImage(img)
			ri = make([]*ebiten.Image, num)
		)
		for i := 0; i < num; i++ {
			var rc image.Rectangle
			rc.Min.Y, rc.Max.X = i*y, x
			rc.Max.Y = rc.Min.Y + y
			ri[num-i-1] = ei.SubImage(rc).(*ebiten.Image)
		}
		return ri
	}

	for _, fv := range zr.File {
		fr, err := fv.Open()
		if err != nil {
			return err
		}

		img, err := png.Decode(fr)
		_ = fr.Close()
		if err != nil {
			return err
		}

		switch fv.Name {
		case "ico.png":
			ebiten.SetWindowIcon([]image.Image{img})
		case "mine.png":
			m.img = subImg(img, 16, 16, 16)
		case "num.png":
			m.num = subImg(img, 13, 23, 12)
		case "face.png":
			m.face = subImg(img, 24, 24, 5)
		}
	}
	return nil
}
