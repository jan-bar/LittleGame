package main

import (
	"log"

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

type (
	chessGame struct {
		images [imgLength]*ebiten.Image   // 所需图片资源
		audios [musicLength]*audio.Player // 所需音频资源

		board [boardX][boardY]uint8 // 棋盘,board[x][y]表示棋子类型

		gameOver bool // 是否游戏结束

		lastX, lastY int
	}
)

func (g *chessGame) Layout(_, _ int) (int, int) {
	return boardWidth, boardHeight
}

func (g *chessGame) Update() error {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if g.gameOver {
			g.reset()
		} else {
			x, y := ebiten.CursorPosition() // 鼠标坐标转换为g.board[x][y]
			x, y = (y-topY)/squareSize, (x-topX)/squareSize
			if x >= 0 && x < boardX && y >= 0 && y < boardY {
				qz, red, sel := getPieces(g.board[x][y])
				if qz > 0 {
					if red {
						switchPieces(&g.board[x][y]) // 切换当前选中红棋状态
						if sel {
							// 选中的红棋再次点击则取消选中状态
							g.lastX, g.lastY = -1, -1
						} else {
							if g.lastX >= 0 {
								// 上次选中某个红棋,由于本次选中别的棋,取消上次选中
								switchPieces(&g.board[g.lastX][g.lastY])
							}
							g.lastX, g.lastY = x, y // 标记本次选中红棋
						}
					} else {
						// 点击黑棋,吃掉对方或者无效走法
					}
				} else {
					// 点击空白位置,判断上次选中红棋走到该位置
				}
			}
		}
	}
	return nil
}

func (g *chessGame) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(g.images[imgChessBoard], op)

	for i := 0; i < boardX; i++ {
		for j := 0; j < boardY; j++ {
			if qz, _, sel := getPieces(g.board[i][j]); qz > 0 {
				op.GeoM.Reset()
				// 图像向右是X,向下是Y,但是数组向下是i,向右是j
				// [8,13]是棋盘左上角起始点,每个棋子长宽squareSize
				xp, yp := float64(j*squareSize+topX), float64(i*squareSize+topY)
				op.GeoM.Translate(xp, yp)
				screen.DrawImage(g.images[qz], op)

				if sel {
					op.GeoM.Translate(0, -5) // 棋子被选中
					screen.DrawImage(g.images[imgSelect], op)
				}
			}
		}
	}

	if g.gameOver {
		ebitenutil.DebugPrintAt(screen, "You Win", 220, 270)
		ebitenutil.DebugPrintAt(screen, "Click Mouse to restart", 180, 290)
	}
}

func getPieces(p uint8) (uint8, bool, bool) {
	// 返回棋子原本的值,以及是否被选中,将最高位作为选中状态
	return p & chessMask, p&chessRed != 0, p&chessSel != 0
}
func switchPieces(p *uint8) {
	if *p&chessMask > 0 {
		// 棋子有值时,切换棋子选中状态
		if *p&chessSel != 0 {
			*p &= ^uint8(chessSel)
		} else {
			*p |= chessSel
		}
	}
}

func (g *chessGame) reset() {
	for i := 0; i < boardX; i++ {
		for j := 0; j < boardY; j++ {
			// 将棋局恢复成开始状态
			g.board[i][j] = boardStart[i][j]
		}
	}
	g.gameOver = false
	g.lastX, g.lastY = -1, -1
}
