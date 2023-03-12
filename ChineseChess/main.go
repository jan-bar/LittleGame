package main

import (
	"log"
	"sync/atomic"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

/*
用到的资源文件,和相关思路来自下面项目
https://github.com/Capricornwqh/ChineseChess
*/

func main() {
	game := &chessGame{aiPoint: make(chan []int, 1)}
	err := game.loadResources()
	if err != nil {
		log.Fatal(err)
	}
	game.reset() // 开局

	ebiten.SetWindowSize(boardWidth, boardHeight)
	ebiten.SetWindowTitle("中国象棋")
	if err = ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

//goland:noinspection SpellCheckingInspection
type (
	chessGame struct {
		images [imgLength]*ebiten.Image   // 所需图片资源
		audios [musicLength]*audio.Player // 所需音频资源

		// 棋盘数据
		board [boardX][boardY]uint8

		// 参照 aiOff, aiOn, aiThink
		isAI atomic.Uint32
		// ai 走黑棋时某棋子从([0],[1])移动到([2],[3])
		aiPoint chan []int
		// 是否游戏结束
		gameOver bool

		// 选中的格子,上一步棋位置
		selected, lastXY [2]int
		// treu:红方,false:黑方
		redPlayer bool
	}
)

func (g *chessGame) Layout(_, _ int) (int, int) {
	return boardWidth, boardHeight
}

func (g *chessGame) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		if !g.isAI.CompareAndSwap(aiOff, aiOn) {
			g.isAI.Store(aiOff) // 按空格键切换 ai对战 / 人人对战
		}
		g.reset()
	}

	switch g.isAI.Load() {
	case aiThink:
		return nil // ai 等待黑棋结果,停止其他任何操作
	case aiOn:
		if !g.redPlayer {
			select {
			case p, ok := <-g.aiPoint:
				if ok { // 模拟黑棋移动某个棋子
					err := g.clickSquare(p[0], p[1])
					if err == nil {
						err = g.clickSquare(p[2], p[3])
					}
					return err
				}
			default:
			}
			return nil // ai 正在玩黑棋,停止其他任何操作
		}
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if g.gameOver {
			g.reset()
		} else {
			x, y := ebiten.CursorPosition()
			// 鼠标坐标转换为g.board[x][y],判断合法则进行走棋逻辑
			x, y = (y-topY)/squareSize, (x-topX)/squareSize
			if x >= 0 && x < boardX && y >= 0 && y < boardY {
				if err := g.clickSquare(x, y); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (g *chessGame) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(g.images[imgChessBoard], op)

	var (
		xy [2]int // 用于和 g.selected,g.lastXY 进行比较

		geoMReset = func(i, j, off int) (xp, yp float64) {
			op.GeoM.Reset()
			// 图像向右是X,向下是Y,但是数组向下是i,向右是j
			// [8,13]是棋盘左上角起始点,每个棋子长宽squareSize
			xp, yp = float64(j*squareSize+topX), float64(i*squareSize+topY+off)
			op.GeoM.Translate(xp, yp)
			return
		}
	)
	for xy[0] = 0; xy[0] < boardX; xy[0]++ {
		for xy[1] = 0; xy[1] < boardY; xy[1]++ {
			if qz := g.board[xy[0]][xy[1]]; qz > 0 {
				geoMReset(xy[0], xy[1], 0)
				screen.DrawImage(g.images[qz], op)

				if g.selected == xy {
					// 棋子被选中,在相对偏移-5位置画圆圈
					op.GeoM.Translate(0, -5)
					screen.DrawImage(g.images[imgSelect], op)
				}
			} else if g.lastXY == xy {
				// 该棋子上次所在位置,圈起来,提示该棋子从哪里走
				geoMReset(xy[0], xy[1], -5)
				screen.DrawImage(g.images[imgSelect], op)
			}
		}
	}

	const infoX, infoY = 245, 270 // 楚河汉界中间位置显示提示信息
	if g.gameOver {
		winMsg := "Black Win"
		if g.redPlayer {
			winMsg = "Red Win"
		}
		ebitenutil.DebugPrintAt(screen, winMsg, infoX, infoY)
		ebitenutil.DebugPrintAt(screen, "Click Mouse to restart", infoX-40, infoY+20)
	} else {
		switch g.isAI.Load() {
		case aiOff:
			ebitenutil.DebugPrintAt(screen, "AI OFF", infoX, infoY)
		case aiOn:
			ebitenutil.DebugPrintAt(screen, "AI ON", infoX, infoY)
		case aiThink:
			ebitenutil.DebugPrintAt(screen, "AI THINK", infoX, infoY)
		}
		ebitenutil.DebugPrintAt(screen, "Key Space to switch", infoX-40, infoY+20)
	}
}

func (g *chessGame) clickSquare(x, y int) (err error) {
	if qz := g.board[x][y]; qz > 0 {
		if isRed(qz) == g.redPlayer {
			err = g.playAudio(musicSelect)
			if err != nil {
				return
			}
			// 点击 g.redPlayer 方棋子,等于切换选中棋子
			g.selected[0], g.selected[1] = x, y
			g.lastXY = g.selected
		} else {
			// 点击 g.redPlayer 对方棋子,尝试吃掉该棋子
			err = g.stepNext(x, y, musicEat)
			if err != nil {
				return
			}
		}
	} else {
		// 点击空白位置,尝试走到该位置
		err = g.stepNext(x, y, musicPut)
		if err != nil {
			return
		}
	}
	return
}

func isRed(p uint8) bool { return p >= imgRedShuai && p <= imgRedBing }

func (g *chessGame) isWin(red bool) bool {
	var qz [][]int
	for i := 0; i < boardX; i++ {
		for j := 0; j < boardY; j++ {
			if isRed(g.board[i][j]) == !red {
				qz = append(qz, []int{i, j}) // 得到敌方所有棋子
			}
		}
	}

	for i := 0; i < boardX; i++ {
		for j := 0; j < boardY; j++ {
			// 该点是空位或我方棋子
			if g.board[i][j] == 0 || isRed(g.board[i][j]) == red {
				for _, v := range qz {
					// 敌方棋子可以走到我方棋子时,进行判将
					if g.canNext(v[0], v[1], i, j) {
						qz0, qz1 := g.board[i][j], g.board[v[0]][v[1]]
						g.board[i][j] = qz1
						g.board[v[0]][v[1]] = 0
						ok := g.isJiang(red) // 假设落子,解除将军,说明还有棋
						g.board[i][j], g.board[v[0]][v[1]] = qz0, qz1
						if !ok {
							return false
						}
					}
				}
			}
		}
	}
	return true // 所有方案均不能挽救将军,胜利
}
func (g *chessGame) isJiang(red bool) bool {
	var i, j, jx, jy, sx, sy int
	for j = 3; j <= 5; j++ {
		for i = 0; i <= 2; i++ {
			if g.board[i][j] == imgBlackJiang {
				jx, jy = i, j
				break
			}
		}

		for i = 7; i <= 9; i++ {
			if g.board[i][j] == imgRedShuai {
				sx, sy = i, j
				break
			}
		}
	}

	if jy == sy {
		ok := true
		for i = jx + 1; i < sx; i++ {
			if g.board[i][jy] != 0 {
				ok = false
				break
			}
		}
		if ok {
			return true // 将和帅之间没有棋子,也算将军
		}
	}

	if red {
		sx, sy = jx, jy
	}
	for i = 0; i < boardX; i++ {
		for j = 0; j < boardY; j++ {
			if isRed(g.board[i][j]) == red {
				if g.canNext(i, j, sx, sy) {
					return true // 玩家棋子下一步可以吃将帅,则当前为将军
				}
			}
		}
	}
	return false
}
func (g *chessGame) playAudio(music int) (err error) {
	if music >= 0 && music < musicLength {
		p := g.audios[music]
		if err = p.Rewind(); err != nil {
			return
		}
		p.Play()
	}
	return
}

func (g *chessGame) reset() {
	for i := 0; i < boardX; i++ {
		for j := 0; j < boardY; j++ {
			// 将棋局恢复成开始状态
			g.board[i][j] = boardStart[i][j]
		}
	}
	g.gameOver = false
	g.lastXY[0], g.lastXY[1] = -1, -1
	g.selected[0], g.selected[1] = -1, -1
	g.redPlayer = true
}

func (g *chessGame) stepNext(x, y, music int) (err error) {
	if g.selected[0] < 0 {
		return // 初始未选中
	}

	if g.canNext(g.lastXY[0], g.lastXY[1], x, y) {
		qz0, qz1 := g.board[x][y], g.board[g.lastXY[0]][g.lastXY[1]]
		g.board[x][y] = qz1 // 尝试走这一步
		g.board[g.lastXY[0]][g.lastXY[1]] = 0

		if g.isJiang(!g.redPlayer) {
			g.board[x][y], g.board[g.lastXY[0]][g.lastXY[1]] = qz0, qz1
			return // 走这一步己方被将军,不能走,恢复局势
		}

		g.selected[0], g.selected[1] = x, y // 成功走出这一步,记录当前选择

		err = g.playAudio(music)
		if err != nil {
			return
		}

		if g.isJiang(g.redPlayer) {
			// 当前将军,敌方没有任何棋子阻止将军,则胜利
			if g.isWin(g.redPlayer) {
				winMusic := musicGameWin
				if !g.redPlayer && g.isAI.Load() == aiOn {
					// 黑棋赢了,启用ai则播放玩家失败音乐
					winMusic = musicGameLose
				}
				err = g.playAudio(winMusic)
				g.gameOver = true
				return // 赢了直接返回
			}
			// 没有赢,因此只播放一下将军
			err = g.playAudio(musicJiang)
			if err != nil {
				return
			}
		}

		if g.redPlayer && g.isAI.Load() == aiOn {
			go g.ai() // 红方走子后,启动协程执行ai走黑棋
		}
		g.redPlayer = !g.redPlayer // 切换角色
	}
	return
}

func (g *chessGame) canNext(x0, y0, x1, y1 int) bool {
	if x0 == x1 && y0 == y1 {
		return false // 起止点不能是同一个
	}

	qz0 := g.board[x0][y0]
	if qz0 <= 0 {
		return false // 第一个位置必须是棋子
	}

	qz1 := g.board[x1][y1]
	if qz1 > 0 && isRed(qz0) == isRed(qz1) {
		return false // 两个都是同类型棋子,不允许
	}

	switch qz0 {
	case imgRedShuai:
		if x1 < 7 || y1 < 3 || y1 > 5 || (x0 != x1 && y0 != y1) ||
			abs(x0, x1) > 1 || abs(y0, y1) > 1 {
			return false // 帅一步只能走一格,只能在己方9宫格走
		}
		return true
	case imgRedShi:
		if x0 == 8 && y0 == 4 {
			if (x1 == 7 && y1 == 3) || (x1 == 9 && y1 == 3) ||
				(x1 == 7 && y1 == 5) || (x1 == 9 && y1 == 5) {
				return true // 当前位置在中心,则只能走四个角
			}
		} else if x1 == 8 && y1 == 4 {
			return true // 否则4个角只能走中心
		}
		return false
	case imgRedXiang:
		if x1 >= 5 && abs(x0, x1) == 2 && abs(y0, y1) == 2 &&
			g.board[(x0+x1)/2][(y0+y1)/2] == 0 {
			return true // 不能过河,只能走田字,不能被填相心
		}
		return false
	case imgRedMa, imgBlackMa: // 红黑马一样
		if (abs(x0, x1) == 2 && abs(y0, y1) == 1 && g.board[(x0+x1)/2][y0] == 0) ||
			(abs(x0, x1) == 1 && abs(y0, y1) == 2 && g.board[x0][(y0+y1)/2] == 0) {
			return true // 只能走日,不能撇脚
		}
		return false
	case imgRedJu, imgBlackJu: // 红黑车一样
		if x0 == x1 {
			min, max := y0+1, y1
			if min > max {
				min, max = y1+1, y0
			}
			for min < max {
				if g.board[x0][min] != 0 {
					return false // 中间有子直接返回
				}
				min++
			}
			return true
		} else if y0 == y1 {
			min, max := x0+1, x1
			if min > max {
				min, max = x1+1, x0
			}
			for min < max {
				if g.board[min][y0] != 0 {
					return false // 中间有子直接返回
				}
				min++
			}
			return true
		}
		return false
	case imgRedPao, imgBlackPao: // 红黑炮规则一样
		if x0 == x1 {
			min, max, cnt := y0+1, y1, 0
			if min > max {
				min, max = y1+1, y0
			}
			for min < max {
				if g.board[x0][min] != 0 {
					if cnt++; cnt > 1 {
						return false // 中间有2子直接返回
					}
				}
				min++
			}
			if (cnt == 0 && qz1 == 0) || (cnt == 1 && qz1 > 0) {
				return true // 中间无棋子,落点为空位 或 中间有1子,落点敌方子
			}
		} else if y0 == y1 {
			min, max, cnt := x0+1, x1, 0
			if min > max {
				min, max = x1+1, x0
			}
			for min < max {
				if g.board[min][y0] != 0 {
					if cnt++; cnt > 1 {
						return false // 中间有2子直接返回
					}
				}
				min++
			}
			if (cnt == 0 && qz1 == 0) || (cnt == 1 && qz1 > 0) {
				return true // 中间无棋子,落点为空位 或 中间有1子,落点敌方子
			}
		}
		return false
	case imgRedBing:
		if x0 < x1 || (x0 != x1 && y0 != y1) || ((x0 == 5 || x0 == 6) && y0 != y1) ||
			abs(x0, x1) > 1 || abs(y0, y1) > 1 {
			return false // 兵不能后退,没过河不能左右走,每次只能向前或左右移动一格
		}
		return true
	// 上面是红棋,下面是黑棋
	case imgBlackJiang:
		if x1 > 2 || y1 < 3 || y1 > 5 || (x0 != x1 && y0 != y1) ||
			abs(x0, x1) > 1 || abs(y0, y1) > 1 {
			return false // 将一步只能走一格,只能在己方9宫格走
		}
		return true
	case imgBlackShi:
		if x0 == 1 && y0 == 4 {
			if (x1 == 0 && y1 == 3) || (x1 == 2 && y1 == 3) ||
				(x1 == 0 && y1 == 5) || (x1 == 2 && y1 == 5) {
				return true // 当前位置在中心,则只能走四个角
			}
		} else if x1 == 1 && y1 == 4 {
			return true // 否则4个角只能走中心
		}
		return false
	case imgBlackXiang:
		if x1 <= 4 && abs(x0, x1) == 2 && abs(y0, y1) == 2 &&
			g.board[(x0+x1)/2][(y0+y1)/2] == 0 {
			return true // 不能过河,只能走田字,不能被填相心
		}
		return false
	case imgBlackBing:
		if x0 > x1 || (x0 != x1 && y0 != y1) || ((x0 == 3 || x0 == 4) && y0 != y1) ||
			abs(x0, x1) > 1 || abs(y0, y1) > 1 {
			return false // 兵不能后退,没过河不能左右走,每次只能向前或左右移动一格
		}
		return true
	default:
		return false
	}
}

func abs(a, b int) int {
	if a -= b; a >= 0 {
		return a
	}
	return -a
}

// -----------------------------------------------------------------------------
func (g *chessGame) ai() {
	g.isAI.Store(aiThink)    // 设置状态,ai思考中
	defer g.isAI.Store(aiOn) // 设置状态,ai思考结束

	// todo ai
	g.aiPoint <- []int{2, 1, 9, 1}
}
