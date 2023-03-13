package main

/*
博弈算法: https://blog.csdn.net/fsdev/category_1085675.html

象棋巫师: https://www.xqbase.com/index.htm
游戏源码: https://github.com/xqbase/xqwlight

教程: https://www.cnblogs.com/royhoo/p/6426394.html
*/

func (g *chessGame) ai() {
	defer g.aiStatus.Store(aiPlay) // 设置状态,ai落子

	// todo ai
	g.lastXY[0], g.lastXY[1] = 2, 1
	g.selected[0], g.selected[1] = 9, 1
}
