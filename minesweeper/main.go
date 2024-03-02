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
	//goland:noinspection GoDeprecation
	rand.Seed(time.Now().Unix())

	// m := &mine{h: 9, w: 9, mineCnt: 10}
	m := &mine{h: 16, w: 16, mineCnt: 40}
	// m := &mine{h: 16, w: 30, mineCnt: 99}
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
		// 0: 正常,1: 赢,2: 输
		playing int
		// 雷区格子数据
		data [][]*grid
		// 总雷数
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
		// 笑脸X坐标
		faceX float64
		// 判断在笑脸位置
		isFace func(h, w int) bool
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

	m.playing = 0
	m.timeStart = time.Time{}
	m.timeCnt = 0
	m.gridW, m.gridH = m.w*gridHW+6, (m.h+2)*gridHW+6

	faceX := m.gridW/2 - 18
	m.faceX = float64(faceX)
	m.isFace = func(h, w int) bool {
		return h >= 4 && h <= 28 && w >= faceX && w < faceX+24
	}
	m.faceNum = 0

	m.timeY = float64(m.gridW - 18)
	ebiten.SetWindowSize(m.gridW, m.gridH)
}

func (m *mine) cursorPos() (h, w, state int) {
	w, h = ebiten.CursorPosition()
	if m.isFace(h, w) {
		state = 1
	} else {
		w, h = (w-3)/gridHW, h/gridHW-2
		if w >= 0 && w < m.w && h >= 0 && h < m.h {
			state = 2
		}
	}
	return
}

func (m *mine) reactionChain(h, w int) {
	d := m.data[h][w]
	if d.state != 0 {
		return // 已打开或插旗 或 游戏完成
	}
	d.state = -1

	switch d.data {
	case 10:
		if m.playing == 0 {
			m.playing = 2
			d.data = 12 // 游戏结束,标记第1个踩到的雷
		}
	case 0: // 递归点开所有空白区域
		m.around(h, w, func(h, w int) { m.reactionChain(h, w) })
	}
}

func (m *mine) Update() error {
	if m.playing != 0 {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			_, _, state := m.cursorPos()
			if state == 1 {
				m.faceNum = 4
			} else {
				switch m.playing {
				case 1:
					m.faceNum = 3
				case 2:
					m.faceNum = 2
				}
			}
		} else if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			_, _, state := m.cursorPos()
			if state == 1 {
				m.initData() // 左键小脸松开重新开始游戏
			}
		}
		return nil
	}

	if m.timeCnt < 999 && !m.timeStart.IsZero() {
		m.timeCnt = int(time.Since(m.timeStart) / time.Second)
	}

	var i, j int
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		for i = 0; i < m.h; i++ {
			for j = 0; j < m.w; j++ {
				if d := m.data[i][j]; d.state == 1 {
					d.state = 0
				}
			}
		}

		h, w, state := m.cursorPos()
		switch state {
		case 1:
			m.faceNum = 4
		case 2:
			switch d := m.data[h][w]; d.state {
			case 0:
				d.state = 1
			case -1:
				m.around(h, w, func(ah, aw int) {
					if ad := m.data[ah][aw]; ad.state == 0 {
						ad.state = 1
					}
				})
			}
			m.faceNum = 1
		default:
			m.faceNum = 1
		}
	} else if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		for i = 0; i < m.h; i++ {
			for j = 0; j < m.w; j++ {
				if d := m.data[i][j]; d.state == 1 {
					d.state = 0
				}
			}
		}
		m.faceNum = 0

		h, w, state := m.cursorPos()
		switch state {
		case 1:
			m.initData() // 笑脸位置松开左键,重新开局
			return nil
		case 2:
			if m.timeStart.IsZero() {
				m.timeStart = time.Now()
			}

			switch d := m.data[h][w]; d.state {
			case 0: // 判断单击
				m.reactionChain(h, w)
			case -1: // 判断双击
				if d.data >= 1 && d.data <= 8 {
					i = 0
					m.around(h, w, func(ah, aw int) {
						if ad := m.data[ah][aw]; ad.state == 2 {
							i++
						}
					})
					if d.data == i {
						m.around(h, w, func(ah, aw int) {
							m.reactionChain(ah, aw)
						})
					}
				}
			}

			if m.playing == 2 {
				m.faceNum = 2

				for i = 0; i < m.h; i++ {
					for j = 0; j < m.w; j++ {
						switch d := m.data[i][j]; d.state {
						case 0: // 将所有雷打开
							if d.data == 10 {
								d.state = -1
							}
						case 2: // 插旗位置不是雷,设置标雷错误
							if d.data != 10 {
								d.state = -1
								d.data = 11
							}
						}
					}
				}
				return nil
			}

			// 点开位置 + 总雷数 = 全部格子数, 此时赢
			h = 0
			for i = 0; i < m.h; i++ {
				for j = 0; j < m.w; j++ {
					if d := m.data[i][j]; d.state == -1 {
						h++
					}
				}
			}

			if h+m.mineCnt == m.h*m.w {
				m.faceNum = 3
				m.playing = 1

				for i = 0; i < m.h; i++ {
					for j = 0; j < m.w; j++ {
						if d := m.data[i][j]; d.state == 0 {
							d.state = 2 // 剩余全插旗
						}
					}
				}
				return nil
			}
		}
	} else if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonRight) {
		h, w, state := m.cursorPos()
		if state == 2 {
			switch d := m.data[h][w]; d.state {
			case 0:
				d.state = 2 // 插旗
			case 2:
				d.state = 0 // 取消
			}
		}
	}
	return nil
}

var backgroundColor = color.RGBA{R: 0xc0, G: 0xc0, B: 0xc0, A: 0xff}

func (m *mine) Draw(screen *ebiten.Image) {
	screen.Fill(backgroundColor)

	ct := m.mineCnt
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(3, 2*gridHW)
	for i := 0; i < m.h; i++ {
		for j := 0; j < m.w; j++ {
			switch d := m.data[i][j]; d.state {
			case 0: // 默认状态
				screen.DrawImage(m.img[15], op)
			case 1: // 按住左键不松开
				screen.DrawImage(m.img[0], op)
			case 2: // 标记旗子
				screen.DrawImage(m.img[14], op)
				ct--
			default: // 按照数据显示
				if d.data == 11 {
					ct-- // 错误插旗也算标雷
				}
				screen.DrawImage(m.img[d.data], op)
			}
			op.GeoM.Translate(gridHW, 0)
		}
		op.GeoM.Translate(0, gridHW)
		op.GeoM.SetElement(0, 2, 3)
	}

	op.GeoM.Reset() // 显示雷数
	op.GeoM.Translate(5, 5)
	var num []int
	if ct >= 0 {
		num = []int{(ct / 100) % 10, (ct / 10) % 10, ct % 10}
	} else {
		ct = -ct // 负数只显示2位
		num = []int{11, (ct / 10) % 10, ct % 10}
	}
	for _, v := range num {
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
	op.GeoM.Translate(m.faceX, 4)
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
