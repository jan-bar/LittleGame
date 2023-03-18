package main

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"strconv"
	"sync/atomic"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/jan-bar/LittleGame/runAI"
)

func main() {
	g := &Gomoku{}
	for i, v := range [][]byte{humImgData, comImgData, humImgWinData, comImgWinData, background} {
		img, _, err := image.Decode(bytes.NewReader(v))
		if err != nil {
			log.Fatal(err)
		}
		g.img[i] = ebiten.NewImageFromImage(img)
	}

	lineColor := color.RGBA{R: 0, G: 0, B: 0, A: 0xff}
	for i := 0; i < boardSize; i++ {
		var (
			lt  = strconv.Itoa(i + 1)
			ln  = 25 + 40*i // 通过调试得到计算数值
			lnf = float64(ln)
		)
		// 为背景图片添加横竖线条,以及每个线条对应数字
		ebitenutil.DrawLine(g.img[4], 0, lnf, 650, lnf, lineColor)
		ebitenutil.DebugPrintAt(g.img[4], lt, 600, ln)
		ebitenutil.DrawLine(g.img[4], lnf, 0, lnf, 650, lineColor)
		ebitenutil.DebugPrintAt(g.img[4], lt, ln, 610)
	}

	bgImg := ebiten.NewImage(screenWidth, screenHeight)
	bgImg.DrawImage(g.img[4], nil)
	ebitenutil.DebugPrintAt(bgImg, "press B to computer first,W to player first", 10, screenWidth)
	g.img[4] = bgImg // 背景宽度比图片大,在最底下显示一些信息

	var err error
	g.aiPlay, err = runAI.NewAI("./gomokuAI.exe")
	if err != nil {
		log.Fatal(err)
	}

	wait := "OK" // 告诉ai开始15*15的五子棋
	g.aiPlay.Send("START 15", runAI.MatchLineContains(&wait))
	wait = "INFO max_memory 83886080\nINFO timeout_match 180000\nINFO timeout_turn 15000\nINFO game_type 0\nINFO rule 0"
	g.aiPlay.Send(wait, nil)

	ebiten.SetWindowClosingHandled(true)
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Gomoku Golang")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

const (
	// 窗口宽一些,用于显示额外文字
	screenWidth = 630
	// 带颜色的图片长和宽
	screenHeight = screenWidth + 20
	// 棋盘横竖格子数
	boardSize = 15

	allNoneFlag = 0

	humImgFlag    = 1 // 玩家棋子
	comImgFlag    = 2 // 电脑棋子
	humWinImgFlag = 3 // 玩家赢了的棋子
	comWinImgFlag = 4 // 电脑赢了的棋子

	statusComputerRun = 1 // 电脑正在思考中
	statusWhiteWin    = 3 // 白棋赢了
	statusBlackWin    = 4 // 黑棋赢了
)

var (
	//go:embed White.png
	humImgData []byte
	//go:embed WhiteWin.png
	humImgWinData []byte
	//go:embed Black.png
	comImgData []byte
	//go:embed BlackWin.png
	comImgWinData []byte
	//go:embed background.jpg
	background []byte
)

//goland:noinspection SpellCheckingInspection
type (
	Gomoku struct {
		// 缓存图片对象
		img [5]*ebiten.Image
		// 五子棋棋盘数据,另一个是界面显示
		board [boardSize][boardSize]int
		// 暂存赢了的5个棋子位置
		win [5][2]int

		// ai 通过该通道更新坐标
		status atomic.Uint32
		// ai 引擎
		aiPlay *runAI.AI
	}
)

func (g *Gomoku) reset() {
	// 重新开始游戏,这里重置游戏棋盘数据
	for i := 0; i < boardSize; i++ {
		for j := 0; j < boardSize; j++ {
			g.board[i][j] = allNoneFlag
		}
	}
	g.status.Store(allNoneFlag) // 清除标记
}

func (g *Gomoku) isWin(i, j, img, imgWin int) bool {
	var x, y, cnt int
	g.win[0][0], g.win[0][1] = i, j

	const five = 5
	defer func() {
		if cnt == five {
			for x = 0; x < five; x++ {
				// 赢了,将5个棋子换成赢了时的状态
				g.board[g.win[x][0]][g.win[x][1]] = imgWin
			}
		}
	}()

	// 横向向左
	for cnt, x = 1, i-1; cnt != five && x >= 0 && g.board[x][j] == img; x-- {
		g.win[cnt][0] = x
		g.win[cnt][1] = j
		cnt++
	}
	// 横向向右
	for x = i + 1; cnt != five && x < boardSize && g.board[x][j] == img; x++ {
		g.win[cnt][0] = x
		g.win[cnt][1] = j
		cnt++
	}
	if cnt == five {
		return true // 横向满足5个
	}

	// 纵向向上
	for cnt, y = 1, j-1; cnt != five && y >= 0 && g.board[i][y] == img; y-- {
		g.win[cnt][0] = i
		g.win[cnt][1] = y
		cnt++
	}
	// 纵向向下
	for y = j + 1; cnt != five && y < boardSize && g.board[i][y] == img; y++ {
		g.win[cnt][0] = i
		g.win[cnt][1] = y
		cnt++
	}
	if cnt == five {
		return true // 纵向满足5个
	}

	// 从落点往左上
	for cnt, x, y = 1, i-1, j-1; cnt != five && x >= 0 && y >= 0 && g.board[x][y] == img; x, y = x-1, y-1 {
		g.win[cnt][0] = x
		g.win[cnt][1] = y
		cnt++
	}
	// 从落点往右下
	for x, y = i+1, j+1; cnt != five && x < boardSize && y < boardSize && g.board[x][y] == img; x, y = x+1, y+1 {
		g.win[cnt][0] = x
		g.win[cnt][1] = y
		cnt++
	}
	if cnt == five {
		return true // 左上右下满足5个
	}

	// 从落点往左下
	for cnt, x, y = 1, i-1, j+1; cnt != five && x >= 0 && y < boardSize && g.board[x][y] == img; x, y = x-1, y+1 {
		g.win[cnt][0] = x
		g.win[cnt][1] = y
		cnt++
	}
	// 从落点往右上
	for x, y = i+1, j-1; cnt != five && x < boardSize && y >= 0 && g.board[x][y] == img; x, y = x+1, y-1 {
		g.win[cnt][0] = x
		g.win[cnt][1] = y
		cnt++
	}
	return cnt == five // 左下右上满足5个
}

func (g *Gomoku) put(x, y, role int) {
	g.board[x][y] = role
}

func (g *Gomoku) remove(x, y int) {
	g.board[x][y] = allNoneFlag
}

func (g *Gomoku) Update() error {
	if ebiten.IsWindowBeingClosed() {
		err := g.aiPlay.Close() // 关闭ai引擎
		if err == nil {
			err = errors.New("close window")
		}
		return err
	}

	if g.status.Load() == statusComputerRun {
		return nil // ai 正在思考,忽略其他任何操作
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyB) {
		g.reset() // 按下B键重新开始游戏,电脑先手,正中央下黑棋
		g.put(boardSize/2, boardSize/2, comImgFlag)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyW) {
		g.reset() // 按下W键重新开始游戏,玩家先手
	} else if g.status.Load() == allNoneFlag && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		x, y = (x-9)/40, (y-9)/40 // 计算鼠标点击位置,此位置没有落子时才响应
		if x >= 0 && y >= 0 && x < boardSize && y < boardSize && g.board[x][y] == allNoneFlag {
			g.put(x, y, humImgFlag)
			if g.isWin(x, y, humImgFlag, humWinImgFlag) {
				g.status.Store(statusWhiteWin) // 人赢了,设置状态
			} else {
				g.status.Store(statusComputerRun) // AI正在思考中

				go g.ai(x, y) // 启动ai协程
			}
		}
	}
	return nil
}

func (g *Gomoku) Draw(screen *ebiten.Image) {
	screen.DrawImage(g.img[4], nil)
	for i := 0; i < boardSize; i++ {
		for j := 0; j < boardSize; j++ {
			if b := g.board[i][j] - 1; b >= allNoneFlag {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(9+40*i), float64(9+40*j))
				screen.DrawImage(g.img[b], op)
			}
		}
	}

	switch g.status.Load() {
	case statusComputerRun:
		ebitenutil.DebugPrintAt(screen,
			"AI is thinking, please wait!", 300, screenWidth)
	case statusWhiteWin:
		ebitenutil.DebugPrintAt(screen,
			"White has won, please restart the game!", 300, screenWidth)
	case statusBlackWin:
		ebitenutil.DebugPrintAt(screen,
			"Black has won, please restart the game!", 300, screenWidth)
	}
}

func (g *Gomoku) Layout(_, _ int) (int, int) {
	return screenWidth, screenHeight
}

/* ai 计算落子 ----------------------------------------------------------------
https://github.com/lihongxun945/gobang 详细讲解AI算法过程,password: Gomoku_js#xxx
https://github.com/Hexik/Embryo_engine 五子棋ai引擎

START 15 启动 15*15 的棋盘, 等待返回OK

INFO max_memory 83886080
INFO timeout_match 600000
INFO timeout_turn 60000
INFO game_type 0
INFO rule 0
INFO time_left 60000
INFO folder C:\xxx

BEGIN   新游戏用这个
RESTART 重新开始用这个,等待OK

每步下棋都发送下面2条命令
INFO time_left 599886
TURN 10,9
等待返回坐标,如下这种
7,10

END 停止思考

继续时,会发送如下指令,用于恢复局面
BOARD
9,10,2
7,10,1
9,12,2
DONE

*/
func (g *Gomoku) ai(x, y int) {
	send := fmt.Sprintf("INFO time_left 179921\nTURN %d,%d", x, y)
	g.aiPlay.Send(send, func(cmp []byte) bool {
		n, _ := fmt.Sscanf(string(cmp), "%d,%d", &x, &y)
		return n == 2 // 发送ai走法,从结果中读取2个数字
	})

	g.put(x, y, comImgFlag) // 完成ai落子
	if g.isWin(x, y, comImgFlag, comWinImgFlag) {
		g.status.Store(statusBlackWin) // AI赢了,设置状态
	} else {
		g.status.Store(allNoneFlag) // AI还没赢,玩家可以继续落子
	}
}
