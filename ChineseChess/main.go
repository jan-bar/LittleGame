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
	game := &chessGame{}
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
	moveXY struct {
		x0, y0, x1, y1 int // [x0,y0] -> [x1,y1]
	}
	chessGame struct {
		images [imgLength]*ebiten.Image   // 所需图片资源
		audios [musicLength]*audio.Player // 所需音频资源

		// 棋盘数据
		board, copy chessBord

		// ai 运行状态
		aiStatus atomic.Uint32

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

func (g *chessGame) Update() (err error) {
	switch g.aiStatus.Load() {
	case aiThink:
		return // ai 正在思考,忽略其他任何操作
	case aiPlay:
		if !g.gameOver { // 游戏没结束
			if err = g.clickSquare(g.selected[0], g.selected[1]); err != nil {
				return
			}
		}
		g.aiStatus.Store(aiOn)
		return // ai模拟黑棋落子,恢复状态
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		if g.selected[0] == -1 {
			// 在初始化时,按空格键切换 ai对战 / 人人对战
			if !g.aiStatus.CompareAndSwap(aiOff, aiOn) {
				g.aiStatus.Store(aiOff)
			}
		} // else {} 玩到中途按空格只会重新开始,不切换模式
		g.reset()
		return
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if g.gameOver {
			g.reset()
		} else {
			x, y := ebiten.CursorPosition()
			// 鼠标坐标转换为g.board[x][y],判断合法则进行走棋逻辑
			x, y = (y-topY)/squareSize, (x-topX)/squareSize
			if x >= 0 && x < boardX && y >= 0 && y < boardY {
				if err = g.clickSquare(x, y); err != nil {
					return
				}
			}
		}
	}
	return
}

func (g *chessGame) Draw(screen *ebiten.Image) {
	aiStatus, board := g.aiStatus.Load(), g.board
	if aiStatus == aiThink {
		board = g.copy // ai 思考时,画界面用 g.copy, g.board 会用于计算
	}

	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(g.images[imgChessBoard], op)

	var (
		xy [2]int // 用于和 g.selected,g.lastXY 进行比较

		geoMReset = func(i, j, off int) {
			op.GeoM.Reset()
			// 图像向右是X,向下是Y,但是数组向下是i,向右是j
			// [8,13]是棋盘左上角起始点,每个棋子长宽squareSize
			xp, yp := float64(j*squareSize+topX), float64(i*squareSize+topY+off)
			op.GeoM.Translate(xp, yp)
		}
	)
	for xy[0] = 0; xy[0] < boardX; xy[0]++ {
		for xy[1] = 0; xy[1] < boardY; xy[1]++ {
			if qz := board[xy[0]][xy[1]]; qz > 0 {
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

	var show0, show1 string
	if g.gameOver {
		if g.redPlayer {
			show0 = "Red Win"
		} else {
			show0 = "Black Win"
		}
		show1 = "Click Mouse To Restart"
	} else {
		switch aiStatus {
		case aiOff:
			show0 = "AI OFF"
		case aiOn:
			show0 = "AI ON"
		case aiThink:
			show0 = "AI THINK"
		}
		show1 = "Key Space To Switch And Restart"
	}
	const infoX, infoY = 245, 270 // 楚河汉界中间位置显示提示信息
	ebitenutil.DebugPrintAt(screen, show0, infoX, infoY)
	ebitenutil.DebugPrintAt(screen, show1, infoX-40, infoY+20)
}

func (g *chessGame) clickSquare(x, y int) (err error) {
	if qz := g.board[x][y]; qz > 0 {
		if isRed(qz) == g.redPlayer {
			if err = g.playAudio(musicSelect); err != nil {
				return
			}
			// 点击 g.redPlayer 方棋子,等于切换选中棋子
			g.selected[0], g.selected[1] = x, y
			g.lastXY = g.selected
		} else {
			// 点击 g.redPlayer 对方棋子,尝试吃掉该棋子
			if err = g.stepNext(x, y, musicEat); err != nil {
				return
			}
		}
	} else {
		// 点击空白位置,尝试走到该位置
		if err = g.stepNext(x, y, musicPut); err != nil {
			return
		}
	}
	return
}

func isRed(p uint8) bool { return p >= imgRedShuai && p <= imgRedBing }

/*
walk
  true:  表示黑棋走子
  false: 表示红棋走子
move:
  move != nil: 记录所有能走的步法(我方被将军时,只返回能解除将军的走法)

return
  true:  表示我方没棋
  false: 表示我方还有棋可走
*/
func (g *chessGame) canStep(walk bool, move *[]moveXY) bool {
	var (
		// 空位和敌方棋子预估值,定义该长度最多扩容1次
		enemy = make([][]int, 0, boardX*boardY/2)
		// 我方棋子最多16个
		our = make([][]int, 0, 16)
	)
	for i := 0; i < boardX; i++ {
		for j := 0; j < boardY; j++ {
			if g.board[i][j] == 0 || isRed(g.board[i][j]) == walk {
				enemy = append(enemy, []int{i, j}) // 空位和敌方棋子
			} else {
				our = append(our, []int{i, j}) // 我方棋子
			}
		}
	}

	if move != nil {
		*move = make([]moveXY, 0, len(our))
	}
	for _, v0 := range our {
		for _, v1 := range enemy {
			if g.canNext(v0[0], v0[1], v1[0], v1[1]) {
				st := moveXY{x0: v0[0], y0: v0[1], x1: v1[0], y1: v1[1]}
				qz1, qz0 := g.board[st.x1][st.y1], g.board[st.x0][st.y0]
				g.board[st.x1][st.y1] = qz0
				g.board[st.x0][st.y0] = 0
				jiang := g.isJiang(walk)
				g.board[st.x1][st.y1], g.board[st.x0][st.y0] = qz1, qz0
				if !jiang { // 我方走这步,敌方未将军
					if move != nil {
						*move = append(*move, st)
					} else {
						return false
					}
				}
			}
		}
	}
	if move != nil {
		return len(*move) == 0 // 没得选也是没棋
	}
	return true // 我方没棋
}

// attack
//   true:  红棋攻击黑将
//   false: 黑棋攻击红帅
func (g *chessGame) isJiang(attack bool) bool {
	var i, j, jx, jy, sx, sy int
	for j = 3; j <= 5; j++ {
		for i = 0; jy == 0 && i <= 2; i++ {
			if g.board[i][j] == imgBlackJiang {
				jx, jy = i, j // 找到黑将
				break
			}
		}

		for i = 7; sy == 0 && i <= 9; i++ {
			if g.board[i][j] == imgRedShuai {
				sx, sy = i, j // 找到红帅
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

	if attack {
		sx, sy = jx, jy // 红棋->将
	} // else {} 黑棋->帅
	for i = 0; i < boardX; i++ {
		for j = 0; j < boardY; j++ {
			if isRed(g.board[i][j]) == attack {
				if g.canNext(i, j, sx, sy) {
					return true // 玩家棋子下一步可以吃将帅,则当前为将军
				}
			}
		}
	}
	return false
}
func (g *chessGame) playAudio(music int) (err error) {
	if music >= musicSelect && music <= musicGameLose {
		p := g.audios[music]
		if err = p.Rewind(); err != nil {
			return
		}
		p.Play()
	}
	return
}

func (g *chessGame) reset() {
	g.loadFEN(boardStart)
	g.gameOver = false
	g.lastXY[0], g.lastXY[1] = -1, -1
	g.selected[0], g.selected[1] = -1, -1
}

func (g *chessGame) stepNext(x, y, music int) (err error) {
	if g.selected[0] < 0 {
		return // 初始未选中
	}

	if lx, ly := g.lastXY[0], g.lastXY[1]; g.canNext(lx, ly, x, y) {
		qz0, qz1 := g.board[x][y], g.board[lx][ly]
		g.board[x][y] = qz1 // 尝试走这一步
		g.board[lx][ly] = 0

		if g.isJiang(!g.redPlayer) {
			g.board[x][y], g.board[lx][ly] = qz0, qz1
			return // 走这一步己方被将军,不能走,恢复局势
		}

		g.selected[0], g.selected[1] = x, y // 成功走出这一步,记录当前选择
		log.Println(g.storeFEN())

		if err = g.playAudio(music); err != nil {
			return
		}

		if g.isJiang(g.redPlayer) {
			// 当前将军,敌方没有任何棋子阻止将军,则胜利
			if g.canStep(g.redPlayer, nil) {
				if !g.redPlayer && g.aiStatus.Load() > aiOff {
					err = g.playAudio(musicGameLose) // ai模式黑棋赢了
				} else {
					err = g.playAudio(musicGameWin)
				}
				g.gameOver = true
				return // 赢了直接返回
			}
			// 没有赢,因此只播放一下将军
			if err = g.playAudio(musicJiang); err != nil {
				return
			}
		}

		if g.redPlayer && g.aiStatus.Load() == aiOn {
			for x = 0; x < boardX; x++ {
				for y = 0; y < boardY; y++ {
					// ai 思考时,界面用 g.copy 渲染
					g.copy[x][y] = g.board[x][y]
				}
			}
			g.aiStatus.Store(aiThink)
			go g.ai() // 设置状态,ai思考中,并启动 ai 协程
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
