package main

import (
	"math/rand"
)

/*
博弈算法: https://blog.csdn.net/fsdev/category_1085675.html

象棋巫师: https://www.xqbase.com/index.htm
游戏源码: https://github.com/xqbase/xqwlight

教程: https://www.cnblogs.com/royhoo/p/6426394.html
*/

func (g *chessGame) ai() {
	defer g.aiStatus.Store(aiPlay) // 设置状态,ai落子

	var move []moveXY
	if g.canStep(true, &move) {
		g.gameOver = true // ai没棋了
	} else {
		m := move[rand.Intn(len(move))] // 随机挑选合理走法
		g.lastXY[0], g.lastXY[1] = m.x0, m.y0
		g.selected[0], g.selected[1] = m.x1, m.y1
	}
}
