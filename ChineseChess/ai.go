package main

func (g *chessGame) ai() {
	g.isAI.Store(aiThink)    // 设置状态,ai思考中
	defer g.isAI.Store(aiOn) // 设置状态,ai思考结束

	// todo ai
	g.aiPoint <- []int{2, 1, 9, 1}
}
