package main

//goland:noinspection SpellCheckingInspection
const (
	imgChessBoard uint8 = iota // 棋盘
	imgSelect                  // 选中
	imgRedShuai                // 红帅
	imgRedShi                  // 红士
	imgRedXiang                // 红相
	imgRedMa                   // 红马
	imgRedJu                   // 红车
	imgRedPao                  // 红炮
	imgRedBing                 // 红兵
	imgBlackJiang              // 黑将
	imgBlackShi                // 黑士
	imgBlackXiang              // 黑相
	imgBlackMa                 // 黑马
	imgBlackJu                 // 黑车
	imgBlackPao                // 黑炮
	imgBlackBing               // 黑兵
	imgLength                  // 图片总长度
)

const (
	musicSelect   = iota // 选子
	musicPut             // 落子
	musicEat             // 吃子
	musicJiang           // 将军
	musicGameWin         // 胜利
	musicGameLose        // 失败
	musicLength          // 音乐总长度
)

const (
	squareSize = 56 // 棋子长宽
	boardEdge  = 8  // 边界大小
	// 棋盘的长和宽,和 ChessBoard.png 图片长宽相同
	boardWidth  = boardEdge + squareSize*9 + boardEdge
	boardHeight = boardEdge + squareSize*10 + boardEdge
)

const (
	boardX, boardY = 10, 9 // 棋盘的x,y格子数
	topX, topY     = 8, 13 // 棋盘左上角起始x,y
)

// 初始棋局
var boardStart = [boardX][boardY]uint8{
	{imgBlackJu, imgBlackMa, imgBlackXiang, imgBlackShi, imgBlackJiang, imgBlackShi, imgBlackXiang, imgBlackMa, imgBlackJu},
	{0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, imgBlackPao, 0, 0, 0, 0, 0, imgBlackPao, 0},
	{imgBlackBing, 0, imgBlackBing, 0, imgBlackBing, 0, imgBlackBing, 0, imgBlackBing},
	{0, 0, 0, 0, 0, 0, 0, 0, 0}, // 楚河
	{0, 0, 0, 0, 0, 0, 0, 0, 0}, // 汉界
	{imgRedBing, 0, imgRedBing, 0, imgRedBing, 0, imgRedBing, 0, imgRedBing},
	{0, imgRedPao, 0, 0, 0, 0, 0, imgRedPao, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0},
	{imgRedJu, imgRedMa, imgRedXiang, imgRedShi, imgRedShuai, imgRedShi, imgRedXiang, imgRedMa, imgRedJu},
}
