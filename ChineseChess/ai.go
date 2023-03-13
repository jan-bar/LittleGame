package main

/*
博弈算法: https://blog.csdn.net/fsdev/category_1085675.html

象棋巫师: https://www.xqbase.com/index.htm
游戏源码: https://github.com/xqbase/xqwlight

教程: https://www.cnblogs.com/royhoo/p/6426394.html
*/

func (g *chessGame) ai() {
	defer g.isAI.Store(aiOn) // 设置状态,ai思考结束

	// todo ai
	g.aiPoint <- aiMove{x0: 2, y0: 1, x1: 9, y1: 1}
}
