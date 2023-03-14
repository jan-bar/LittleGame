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

type chessBord [boardX][boardY]uint8

func flipPiece(p chessBord) (res chessBord) {
	for i := 0; i < boardX; i++ {
		for j := 0; j < boardY; j++ {
			res[boardX-1-i][j] = p[i][j] // 翻转棋盘分数
		}
	}
	return
}

//goland:noinspection SpellCheckingInspection
var (
	shuaiBing = chessBord{
		{9, 9, 9, 11, 13, 11, 9, 9, 9},
		{19, 24, 34, 42, 44, 42, 34, 24, 19},
		{19, 24, 32, 37, 37, 37, 32, 24, 19},
		{19, 23, 27, 29, 30, 29, 27, 23, 19},
		{14, 18, 20, 27, 29, 27, 20, 18, 14},
		{7, 0, 13, 0, 16, 0, 13, 0, 7},
		{7, 0, 7, 0, 15, 0, 7, 0, 7},
		{0, 0, 0, 1, 1, 1, 0, 0, 0},
		{0, 0, 0, 2, 2, 2, 0, 0, 0},
		{0, 0, 0, 11, 15, 11, 0, 0, 0},
	}
	jiangBing = flipPiece(shuaiBing)

	redShi = chessBord{
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 20, 0, 0, 0, 20, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{18, 0, 0, 20, 23, 20, 0, 0, 18},
		{0, 0, 0, 0, 23, 0, 0, 0, 0},
		{0, 0, 20, 20, 0, 20, 20, 0, 0},
	}
	redXiang = chessBord{
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 20, 0, 0, 0, 20, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{18, 0, 0, 20, 23, 20, 0, 0, 18},
		{0, 0, 0, 0, 23, 0, 0, 0, 0},
		{0, 0, 20, 20, 0, 20, 20, 0, 0},
	}
	redMa = chessBord{
		{90, 90, 90, 96, 90, 96, 90, 90, 90},
		{90, 96, 103, 97, 94, 97, 103, 96, 90},
		{92, 98, 99, 103, 99, 103, 99, 98, 92},
		{93, 108, 100, 107, 100, 107, 100, 108, 93},
		{90, 100, 99, 103, 104, 103, 99, 100, 90},
		{90, 98, 101, 102, 103, 102, 101, 98, 90},
		{92, 94, 98, 95, 98, 95, 98, 94, 92},
		{93, 92, 94, 95, 92, 95, 94, 92, 93},
		{85, 90, 92, 93, 78, 93, 92, 90, 85},
		{88, 85, 90, 88, 90, 88, 90, 85, 88},
	}
	redJu = chessBord{
		{206, 208, 207, 213, 214, 213, 207, 208, 206},
		{206, 212, 209, 216, 233, 216, 209, 212, 206},
		{206, 208, 207, 214, 216, 214, 207, 208, 206},
		{206, 213, 213, 216, 216, 216, 213, 213, 206},
		{208, 211, 211, 214, 215, 214, 211, 211, 208},
		{208, 212, 212, 214, 215, 214, 212, 212, 208},
		{204, 209, 204, 212, 214, 212, 204, 209, 204},
		{198, 208, 204, 212, 212, 212, 204, 208, 198},
		{200, 208, 206, 212, 200, 212, 206, 208, 200},
		{194, 206, 204, 212, 200, 212, 204, 206, 194},
	}
	redPao = chessBord{
		{100, 100, 96, 91, 90, 91, 96, 100, 100},
		{98, 98, 96, 92, 89, 92, 96, 98, 98},
		{97, 97, 96, 91, 92, 91, 96, 97, 97},
		{96, 99, 99, 98, 100, 98, 99, 99, 96},
		{96, 96, 96, 96, 100, 96, 96, 96, 96},
		{95, 96, 99, 96, 100, 96, 99, 96, 95},
		{96, 96, 96, 96, 96, 96, 96, 96, 96},
		{97, 96, 100, 99, 101, 99, 100, 96, 97},
		{96, 97, 98, 98, 98, 98, 98, 97, 96},
		{96, 96, 97, 99, 99, 99, 97, 96, 96},
	}

	// 每个棋子都有对应的局面分数,红黑相同棋子分数上下翻转
	pieceValue = map[uint8]chessBord{
		imgRedShuai:   shuaiBing,
		imgRedBing:    shuaiBing,
		imgBlackJiang: jiangBing,
		imgBlackBing:  jiangBing, // 兵和将帅落点不冲突,共用
		imgRedShi:     redShi,
		imgBlackShi:   flipPiece(redShi),
		imgRedXiang:   redXiang,
		imgBlackXiang: flipPiece(redXiang),
		imgRedMa:      redMa,
		imgBlackMa:    flipPiece(redMa),
		imgRedJu:      redJu,
		imgBlackJu:    flipPiece(redJu),
		imgRedPao:     redPao,
		imgBlackPao:   flipPiece(redPao),
	}
)
