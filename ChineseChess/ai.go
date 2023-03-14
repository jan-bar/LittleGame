package main

import (
	"log"
)

/*
博弈算法: https://blog.csdn.net/fsdev/category_1085675.html

象棋巫师: https://www.xqbase.com/index.htm
游戏源码: https://github.com/xqbase/xqwlight

教程: https://www.cnblogs.com/royhoo/p/6426394.html
*/

func (g *chessGame) ai() {
	defer g.aiStatus.Store(aiPlay) // 设置状态,ai落子

	// var move []moveXY
	// if g.canStep(true, &move) {
	// 	g.gameOver = true // ai没棋了
	// } else {
	// 	m := move[rand.Intn(len(move))] // 随机挑选合理走法
	// 	g.lastXY[0], g.lastXY[1] = m.x0, m.y0
	// 	g.selected[0], g.selected[1] = m.x1, m.y1
	// }
	g.maxMinSearch(false)
}

const (
	minMaxDepth = 3
	mateValue   = 1000
)

func (g *chessGame) maxMinSearch(red bool) {
	if red {
		g.maxSearch(red, minMaxDepth)
	} else {
		g.minSearch(red, minMaxDepth)
	}
}

func (g *chessGame) evaluate() int {
	return g.vlRed - g.vlBlack
}
func (g *chessGame) addPiece(x, y int, p uint8, del bool) {
	pv := pieceValue[p]
	if del {
		g.board[x][y] = 0
		if isRed(p) {
			g.vlRed -= int(pv[x][y])
		} else {
			g.vlBlack -= int(pv[x][y])
		}
	} else {
		g.board[x][y] = p
		if isRed(p) {
			g.vlRed += int(pv[x][y])
		} else {
			g.vlBlack += int(pv[x][y])
		}
	}
}
func (g *chessGame) movePiece(m moveXY, sp, dp uint8) {
	if dp > 0 {
		g.addPiece(m.x1, m.y1, dp, true)
	}
	g.addPiece(m.x0, m.y0, sp, true)
	g.addPiece(m.x1, m.y1, sp, false)
}
func (g *chessGame) makeMove(m moveXY, sp, dp uint8) {
	g.movePiece(m, sp, dp)
	g.redPlayer = !g.redPlayer
}
func (g *chessGame) undoMovePiece(m moveXY, sp, dp uint8) {
	g.addPiece(m.x1, m.y1, sp, true)
	g.addPiece(m.x0, m.y0, sp, false)

	if dp > 0 {
		g.addPiece(m.x1, m.y1, dp, false)
	}
}
func (g *chessGame) undoMakeMove(m moveXY, sp, dp uint8) {
	g.redPlayer = !g.redPlayer
	g.undoMovePiece(m, sp, dp)
}
func (g *chessGame) maxSearch(red bool, depth int) int {
	if depth == 0 {
		return g.evaluate()
	}

	vlBest := -mateValue
	var move []moveXY
	g.canStep(red, &move)
	var value int
	for _, v := range move {
		sp, dp := g.board[v.x0][v.y0], g.board[v.x1][v.y1]
		s1 := g.storeFEN()
		g.makeMove(v, sp, dp)
		value = g.minSearch(!red, depth-1)
		g.undoMakeMove(v, sp, dp)
		s2 := g.storeFEN()
		if s1 != s2 {
			log.Fatalln(s1, s2)
		}

		if value > vlBest {
			// 找到了当前的最佳值
			vlBest = value

			// 如果回到了根节点，需要记录根节点的最佳走法
			if depth == minMaxDepth {
				g.aiMove = v
			}
		}
	}

	return vlBest
}
func (g *chessGame) minSearch(red bool, depth int) int {
	if depth == 0 {
		return g.evaluate()
	}

	vlBest := mateValue
	var move []moveXY
	g.canStep(!red, &move)
	var value int
	for _, v := range move {
		sp, dp := g.board[v.x0][v.y0], g.board[v.x1][v.y1]
		s1 := g.storeFEN()
		g.makeMove(v, sp, dp)
		value = g.maxSearch(red, depth-1)
		g.undoMakeMove(v, sp, dp)
		s2 := g.storeFEN()
		if s1 != s2 {
			log.Fatalln(s1, s2)
		}

		if value < vlBest { // 这里与极大点搜索不同
			vlBest = value
			if depth == minMaxDepth {
				g.aiMove = v
			}
		}
	}

	return vlBest
}
