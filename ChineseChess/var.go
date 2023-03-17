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
	// 棋盘的长,和 ChessBoard.png 图片长相同
	boardWidth = boardEdge + squareSize*9 + boardEdge
	// 棋盘的宽,在 ChessBoard.png 图片宽度增加底部信息显示
	boardHeight = boardEdge + squareSize*10 + boardEdge + 25
)

//goland:noinspection SpellCheckingInspection
const (
	boardX, boardY = 10, 9 // 棋盘的x,y格子数
	topX, topY     = 8, 13 // 棋盘左上角起始x,y

	// 开局棋谱
	boardStart = "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"
)

const (
	aiOff   = iota // 禁用ai
	aiOn           // 启用ai
	aiPlay         // ai落子
	aiThink        // ai正在思考
)
