package main

//goland:noinspection SpellCheckingInspection
const (
	imgChessBoard = iota // 棋盘
	imgSelect            // 选中
	imgRedShuai          // 红帅
	imgRedShi            // 红士
	imgRedXiang          // 红相
	imgRedMa             // 红马
	imgRedJu             // 红车
	imgRedPao            // 红炮
	imgRedBing           // 红兵
	imgBlackJiang        // 黑将
	imgBlackShi          // 黑士
	imgBlackXiang        // 黑相
	imgBlackMa           // 黑马
	imgBlackJu           // 黑车
	imgBlackPao          // 黑炮
	imgBlackBing         // 黑兵
	imgLength            // 图片总长度
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

	chessMask = 0x3f // 棋子对应值掩码
	chessSel  = 0x80 // 该bit表示选中状态
	chessRed  = 0x40 // 该bit表示该棋子是红方
)

// 初始棋局
var boardStart = [boardX][boardY]uint8{
	{imgBlackJu, imgBlackMa, imgBlackXiang, imgBlackShi, imgBlackJiang, imgBlackShi, imgBlackXiang, imgBlackMa, imgBlackJu},
	{0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, imgBlackPao, 0, 0, 0, 0, 0, imgBlackPao, 0},
	{imgBlackBing, 0, imgBlackBing, 0, imgBlackBing, 0, imgBlackBing, 0, imgBlackBing},
	{0, 0, 0, 0, 0, 0, 0, 0, 0}, // 楚河
	{0, 0, 0, 0, 0, 0, 0, 0, 0}, // 汉界
	{imgRedBing | chessRed, 0, imgRedBing | chessRed, 0, imgRedBing | chessRed, 0, imgRedBing | chessRed, 0, imgRedBing | chessRed},
	{0, imgRedPao | chessRed, 0, 0, 0, 0, 0, imgRedPao | chessRed, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0},
	{imgRedJu | chessRed, imgRedMa | chessRed, imgRedXiang | chessRed, imgRedShi | chessRed, imgRedShuai | chessRed, imgRedShi | chessRed, imgRedXiang | chessRed, imgRedMa | chessRed, imgRedJu | chessRed},
}
