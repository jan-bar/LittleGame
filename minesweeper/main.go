package main

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func main() {
	//goland:noinspection GoDeprecation
	rand.Seed(time.Now().Unix())

	m := &mine{h: 16, w: 30, mineCnt: 99}
	err := m.loadResources()
	if err != nil {
		log.Fatal(err)
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
		// 计时器最右侧数字坐标
		timeX float64
		// 界面宽高
		gridW, gridH int
		// 显示哪个笑脸
		faceNum int
		// 笑脸坐标
		faceX float64
		// 判断在笑脸位置
		isFace func(h, w int) bool
		// 界面图片,数字图片,笑脸图片
		img, num, face []*ebiten.Image
		// 背景图片
		background *ebiten.Image
		// 显示输入数据
		text string
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
	var i, j, k int

	if len(m.data) < m.h {
		m.data = make([][]*grid, m.h)
	}
	for i = 0; i < m.h; i++ {
		if len(m.data[i]) < m.w {
			m.data[i] = make([]*grid, m.w)
		}
	}

	for i = 0; i < m.h; i++ {
		for j = 0; j < m.w; j++ {
			d := m.data[i][j]
			if d == nil {
				d = new(grid)
				m.data[i][j] = d
			}
			d.state = 0

			if k < m.mineCnt {
				d.data = 10
				k++ // 布雷
			} else {
				d.data = 0
			}
		}
	}

	rand.Shuffle(m.h*m.w, func(i, j int) {
		mi := m.data[i/m.w][i%m.w]
		mj := m.data[j/m.w][j%m.w] // 洗牌算法打乱雷区
		mi.data, mj.data = mj.data, mi.data
	})

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
	m.gridW, m.gridH = m.w*gridHW+6, (m.h+3)*gridHW+6

	faceX := m.gridW/2 - 18
	m.faceX = float64(faceX)
	m.isFace = func(h, w int) bool {
		return h >= 4 && h <= 28 && w >= faceX && w < faceX+24
	}
	m.faceNum = 0

	m.timeX = float64(m.gridW - 18)
	ebiten.SetWindowSize(m.gridW, m.gridH)

	m.background = ebiten.NewImage(m.gridW, m.gridH-gridHW)
	m.background.Fill(backgroundColor) // 创建背景图片
	m.text = fmt.Sprintf("H:%d,W:%d,M:%d >", m.h, m.w, m.mineCnt)
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
	var state int
	if m.playing != 0 {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			_, _, state = m.cursorPos()
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
			_, _, state = m.cursorPos()
			if state == 1 {
				m.initData() // 左键小脸松开重新开始游戏
			}
		}
		return nil
	}

	if m.timeCnt < 999 && !m.timeStart.IsZero() {
		m.timeCnt = int(time.Since(m.timeStart) / time.Second)
	}

	var d *grid
	var i, j int
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		for i = 0; i < m.h; i++ {
			for j = 0; j < m.w; j++ {
				if d = m.data[i][j]; d.state == 1 {
					d.state = 0
				}
			}
		}

		i, j, state = m.cursorPos()
		switch state {
		case 1:
			m.faceNum = 4
		case 2:
			switch d = m.data[i][j]; d.state {
			case 0:
				d.state = 1
			case -1:
				m.around(i, j, func(ah, aw int) {
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
				if d = m.data[i][j]; d.state == 1 {
					d.state = 0
				}
			}
		}
		m.faceNum = 0

		i, j, state = m.cursorPos()
		switch state {
		case 1:
			m.initData() // 笑脸位置松开左键,重新开局
			return nil
		case 2:
			if m.timeStart.IsZero() {
				m.timeStart = time.Now()
			}

			switch d = m.data[i][j]; d.state {
			case 0: // 判断单击
				m.reactionChain(i, j)
			case -1: // 判断双击
				if d.data >= 1 && d.data <= 8 {
					state = 0
					m.around(i, j, func(ah, aw int) {
						if ad := m.data[ah][aw]; ad.state == 2 {
							state++
						}
					})
					if d.data == state {
						m.around(i, j, func(ah, aw int) {
							m.reactionChain(ah, aw)
						})
					}
				}
			}

			if m.playing == 2 {
				m.faceNum = 2 // 游戏结束,输了

				for i = 0; i < m.h; i++ {
					for j = 0; j < m.w; j++ {
						switch d = m.data[i][j]; d.state {
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

			state = 0
			for i = 0; i < m.h; i++ {
				for j = 0; j < m.w; j++ {
					if d = m.data[i][j]; d.state == -1 {
						state++
					}
				}
			}
			// 点开位置 + 总雷数 = 全部格子数, 此时赢
			if state+m.mineCnt == m.h*m.w {
				m.faceNum = 3 // 游戏结束,赢了
				m.playing = 1

				for i = 0; i < m.h; i++ {
					for j = 0; j < m.w; j++ {
						if d = m.data[i][j]; d.state == 0 {
							d.state = 2 // 剩余全插旗
						}
					}
				}
				return nil
			}
		}
	} else if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonRight) {
		i, j, state = m.cursorPos()
		if state == 2 {
			switch d = m.data[i][j]; d.state {
			case 0:
				d.state = 2 // 插旗
			case 2:
				d.state = 0 // 取消
			}
		}
	}

	for k, v := range eKey {
		if inpututil.IsKeyJustReleased(k) {
			switch v {
			case "d":
				if i = len(m.text) - 1; m.text[i] != '>' {
					m.text = m.text[:i]
				}
			case "e":
				i = strings.IndexByte(m.text, '>') + 1

				var ok bool
				n, _ := fmt.Sscanf(m.text[i:], "%d %d %d", &i, &j, &state)
				switch n {
				case 3: // 读取 h/w/mine 这3个数据
					if i >= 9 && i <= 45 && j >= 9 && j <= 45 &&
						state >= 10 && state <= (i-1)*(j-1) {
						m.h, m.w, m.mineCnt = i, j, state
						ok = true
					}
				case 1: // 输入单个数字切换难度模式
					switch i {
					case 1:
						m.h, m.w, m.mineCnt = 9, 9, 10 // 初级
						ok = true
					case 2:
						m.h, m.w, m.mineCnt = 16, 16, 40 // 中级
						ok = true
					case 3:
						m.h, m.w, m.mineCnt = 16, 30, 99 // 高级
						ok = true
					case 4:
						m.h, m.w, m.mineCnt = 24, 30, 99 // 最大
						ok = true
					}
				}

				if ok {
					m.initData()
					return nil
				}
			default:
				m.text += v
			}
		}
	}
	return nil
}

var (
	eKey = map[ebiten.Key]string{
		ebiten.KeyDigit0: "0",
		ebiten.KeyDigit1: "1",
		ebiten.KeyDigit2: "2",
		ebiten.KeyDigit3: "3",
		ebiten.KeyDigit4: "4",
		ebiten.KeyDigit5: "5",
		ebiten.KeyDigit6: "6",
		ebiten.KeyDigit7: "7",
		ebiten.KeyDigit8: "8",
		ebiten.KeyDigit9: "9",
		ebiten.KeySpace:  " ",

		ebiten.KeyBackspace:   "d", // 删除
		ebiten.KeyEnter:       "e", // 回车
		ebiten.KeyNumpadEnter: "e", // 回车
	}

	backgroundColor = color.RGBA{R: 0xc0, G: 0xc0, B: 0xc0, A: 0xff}
)

func (m *mine) Draw(screen *ebiten.Image) {
	screen.DrawImage(m.background, nil)

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
	op.GeoM.Translate(m.timeX, 5)
	for _, v := range []int{m.timeCnt % 10, (m.timeCnt / 10) % 10, (m.timeCnt / 100) % 10} {
		screen.DrawImage(m.num[v], op)
		op.GeoM.Translate(-13, 0)
	}

	op.GeoM.Reset() // 显示笑脸
	op.GeoM.Translate(m.faceX, 4)
	screen.DrawImage(m.face[m.faceNum], op)

	// 打印输入的[高 宽 雷]数据,以及输入数据显示
	ebitenutil.DebugPrintAt(screen, m.text, 10, m.gridH-gridHW)
}

func (m *mine) Layout(_, _ int) (int, int) {
	return m.gridW, m.gridH
}

func (m *mine) loadResources() error {
	//goland:noinspection SpellCheckingInspection
	const sd = "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"
	bd, err := base64.StdEncoding.DecodeString(sd)
	if err != nil {
		return err
	}

	data := bytes.NewReader(bd)
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
