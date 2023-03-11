package main

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math/rand"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func main() {
	g := &Gomoku{
		searchDeep:   7,
		deepDecrease: 0.8,
		countLimit:   10,
		threshold:    1.1,
		cache:        make(map[int64]*gomokuCache),
		aiStatus:     make(chan int),
	}
	for i, v := range [][]byte{humImgData, comImgData, humImgWinData, comImgWinData, background} {
		img, _, err := image.Decode(bytes.NewReader(v))
		if err != nil {
			log.Fatal(err)
		}
		g.img[i] = ebiten.NewImageFromImage(img)
	}

	for g.zobristCode == 0 {
		g.zobristCode = rand.Int63n(1000000000) // 初始化随机hash值
	}

	lineColor := color.RGBA{R: 0, G: 0, B: 0, A: 0xff}
	for i := 0; i < boardSize; i++ {
		for j := 0; j < boardSize; j++ {
			for g.zobrist[i][j][0] == 0 { // 玩家随机值
				g.zobrist[i][j][0] = rand.Int63n(1000000000)
			}
			for g.zobrist[i][j][1] == 0 { // 电脑随机值
				g.zobrist[i][j][1] = rand.Int63n(1000000000)
			}
		}

		var (
			lt  = strconv.Itoa(i + 1)
			ln  = 25 + 40*i // 通过调试得到计算数值
			lnf = float64(ln)
		)
		// 为背景图片添加横竖线条,以及每个线条对应数字
		ebitenutil.DrawLine(g.img[4], 0, lnf, 650, lnf, lineColor)
		ebitenutil.DebugPrintAt(g.img[4], lt, 600, ln)
		ebitenutil.DrawLine(g.img[4], lnf, 0, lnf, 650, lineColor)
		ebitenutil.DebugPrintAt(g.img[4], lt, ln, 610)
	}

	bgImg := ebiten.NewImage(screenWidth, screenHeight)
	bgImg.DrawImage(g.img[4], nil)
	ebitenutil.DebugPrintAt(bgImg, "press B to computer first,W to player first", 10, screenWidth)
	g.img[4] = bgImg // 背景宽度比图片大,在最底下显示一些信息

	go g.ai() // 启动协程运行ai

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Gomoku Golang")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

const (
	// 窗口宽一些,用于显示额外文字
	screenWidth = 630
	// 带颜色的图片长和宽
	screenHeight = screenWidth + 20
	// 棋盘横竖格子数
	boardSize = 15

	allNoneFlag = 0

	humImgFlag    = 1 // 玩家棋子
	comImgFlag    = 2 // 电脑棋子
	humWinImgFlag = 3 // 玩家赢了的棋子
	comWinImgFlag = 4 // 电脑赢了的棋子

	statusComputerRun = 1 // 电脑正在思考中
	statusWhiteWin    = 2 // 白棋赢了
	statusBlackWin    = 3 // 黑棋赢了
)

var (
	//go:embed White.png
	humImgData []byte
	//go:embed WhiteWin.png
	humImgWinData []byte
	//go:embed Black.png
	comImgData []byte
	//go:embed BlackWin.png
	comImgWinData []byte
	//go:embed background.jpg
	background []byte
)

//goland:noinspection SpellCheckingInspection
type (
	gomokuCache struct {
		deep            int
		score, maxScore float64
	}
	Gomoku struct {
		// 缓存图片对象
		img [5]*ebiten.Image
		// 五子棋棋盘数据,另一个是界面显示
		board, show [boardSize][boardSize]int
		// 保存当前状态
		status int
		// 暂存赢了的5个棋子位置
		win [5][2]int
		// zobrist 棋盘每个位置的随机值
		zobrist [boardSize][boardSize][2]int64
		// zobrist 棋盘hash code
		zobristCode int64
		// 缓存 zobristCode 棋盘状态时的 deep,score
		cache map[int64]*gomokuCache
		// ai 通过该通道更新坐标
		aiStatus chan int

		comScore [boardSize][boardSize]float64 // 电脑分数
		humScore [boardSize][boardSize]float64 // 玩家分数

		searchDeep   int     // 搜索深度
		deepDecrease float64 // 按搜索深度递减分数,为了让短路径的结果比深路径的分数高
		countLimit   int     // gen函数返回的节点数量上限,超过之后将会按照分数进行截断
		threshold    float64 // 阈值
	}
)

func (g *Gomoku) reset() {
	// 重新开始游戏,这里重置游戏棋盘数据
	for i := 0; i < boardSize; i++ {
		for j := 0; j < boardSize; j++ {
			g.board[i][j] = allNoneFlag
		}
	}
	g.status = allNoneFlag // 清除标记
}

func (g *Gomoku) isWin(i, j, img, imgWin int) bool {
	var x, y, cnt int
	g.win[0][0], g.win[0][1] = i, j

	const five = 5
	defer func() {
		if cnt == five {
			for x = 0; x < five; x++ {
				// 赢了,将5个棋子换成赢了时的状态
				g.board[g.win[x][0]][g.win[x][1]] = imgWin
			}
		}
	}()

	// 横向向左
	for cnt, x = 1, i-1; cnt != five && x >= 0 && g.board[x][j] == img; x-- {
		g.win[cnt][0] = x
		g.win[cnt][1] = j
		cnt++
	}
	// 横向向右
	for x = i + 1; cnt != five && x < boardSize && g.board[x][j] == img; x++ {
		g.win[cnt][0] = x
		g.win[cnt][1] = j
		cnt++
	}
	if cnt == five {
		return true // 横向满足5个
	}

	// 纵向向上
	for cnt, y = 1, j-1; cnt != five && y >= 0 && g.board[i][y] == img; y-- {
		g.win[cnt][0] = i
		g.win[cnt][1] = y
		cnt++
	}
	// 纵向向下
	for y = j + 1; cnt != five && y < boardSize && g.board[i][y] == img; y++ {
		g.win[cnt][0] = i
		g.win[cnt][1] = y
		cnt++
	}
	if cnt == five {
		return true // 纵向满足5个
	}

	// 从落点往左上
	for cnt, x, y = 1, i-1, j-1; cnt != five && x >= 0 && y >= 0 && g.board[x][y] == img; x, y = x-1, y-1 {
		g.win[cnt][0] = x
		g.win[cnt][1] = y
		cnt++
	}
	// 从落点往右下
	for x, y = i+1, j+1; cnt != five && x < boardSize && y < boardSize && g.board[x][y] == img; x, y = x+1, y+1 {
		g.win[cnt][0] = x
		g.win[cnt][1] = y
		cnt++
	}
	if cnt == five {
		return true // 左上右下满足5个
	}

	// 从落点往左下
	for cnt, x, y = 1, i-1, j+1; cnt != five && x >= 0 && y < boardSize && g.board[x][y] == img; x, y = x-1, y+1 {
		g.win[cnt][0] = x
		g.win[cnt][1] = y
		cnt++
	}
	// 从落点往右上
	for x, y = i+1, j-1; cnt != five && x < boardSize && y >= 0 && g.board[x][y] == img; x, y = x+1, y-1 {
		g.win[cnt][0] = x
		g.win[cnt][1] = y
		cnt++
	}
	return cnt == five // 左下右上满足5个
}

func (g *Gomoku) put(x, y, role int) {
	g.board[x][y] = role
	// 使用x,y位置role角色的随机值设置hash code
	g.zobristCode ^= g.zobrist[x][y][role-1]
	g.updateScore(x, y)
}

func (g *Gomoku) remove(x, y int) {
	// 使用x,y位置role角色的随机值清除hash code
	g.zobristCode ^= g.zobrist[x][y][g.board[x][y]-1]
	g.board[x][y] = allNoneFlag
	g.updateScore(x, y)
}

func (g *Gomoku) Update() error {
	// 这里的chan确保 g.status 是线程安全
	select {
	case s, ok := <-g.aiStatus:
		if ok {
			g.status = s
		}
	default:
		if g.status == statusComputerRun {
			return nil // ai 思考时不允许其他操作
		}
	}

	var sendAI bool
	if inpututil.IsKeyJustPressed(ebiten.KeyB) {
		g.reset() // 按下B键重新开始游戏,电脑先手,正中央下黑棋
		g.put(boardSize/2, boardSize/2, comImgFlag)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyW) {
		g.reset() // 按下W键重新开始游戏,玩家先手
	} else if g.status == allNoneFlag && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		x, y = (x-9)/40, (y-9)/40 // 计算鼠标点击位置,此位置没有落子时才响应
		if x >= 0 && y >= 0 && x < boardSize && y < boardSize && g.board[x][y] == allNoneFlag {
			g.put(x, y, humImgFlag)
			if g.isWin(x, y, humImgFlag, humWinImgFlag) {
				g.status = statusWhiteWin // 人赢了,设置状态
			} else {
				g.status = statusComputerRun // AI正在思考中

				sendAI = true
			}
		}
	}

	for i := 0; i < boardSize; i++ {
		for j := 0; j < boardSize; j++ {
			g.show[i][j] = g.board[i][j]
		}
	}

	// 在将 g.board 复制到 g.show 之后才该ai发消息,确保线程安全
	// 更新UI只用到了 g.status 和 g.show ,确保这两个变量线程安全就OK
	if sendAI {
		select {
		case g.aiStatus <- 0: // 发信号让ai
		default:
		}
	}
	return nil
}

func (g *Gomoku) Draw(screen *ebiten.Image) {
	screen.DrawImage(g.img[4], nil)
	for i := 0; i < boardSize; i++ {
		for j := 0; j < boardSize; j++ {
			if b := g.show[i][j] - 1; b >= allNoneFlag {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(9+40*i), float64(9+40*j))
				screen.DrawImage(g.img[b], op)
			}
		}
	}

	switch g.status {
	case statusComputerRun:
		ebitenutil.DebugPrintAt(screen,
			"AI is thinking, please wait!", 300, screenWidth)
	case statusWhiteWin:
		ebitenutil.DebugPrintAt(screen,
			"White has won, please restart the game!", 300, screenWidth)
	case statusBlackWin:
		ebitenutil.DebugPrintAt(screen,
			"Black has won, please restart the game!", 300, screenWidth)
	}
}

func (g *Gomoku) Layout(_, _ int) (int, int) {
	return screenWidth, screenHeight
}

// ai 计算落子 ----------------------------------------------------------------
// https://github.com/lihongxun945/gobang 详细讲解AI算法过程,password: Gomoku_js#xxx
const (
	scoreOne          = 10      // 活一
	scoreTwo          = 100     // 活二
	scoreThree        = 1000    // 活三
	scoreFour         = 100000  // 活四
	scoreFive         = 1000000 // 连五
	scoreBlockedOne   = 1       // 眠一
	scoreBlockedTwo   = 10      // 眠二
	scoreBlockedThree = 100     // 眠三
	scoreBlockedFour  = 10000   // 眠四

	scoreFiveNeg = float64(-scoreFive)
	scoreMin     = scoreFiveNeg * 10
)

func (g *Gomoku) ai() {
	var (
		x, y  int
		score float64
	)
	for {
		select {
		case <-g.aiStatus:
			for i := 2; i <= g.searchDeep; i += 2 {
				x, y, score = g.maxMin(i)
				if g.greatOrEqualThan(score, scoreFour) {
					break // 所得分数大于阈值计算则不用继续搜索
				}
			}

			g.put(x, y, comImgFlag) // 完成ai落子
			if g.isWin(x, y, comImgFlag, comWinImgFlag) {
				g.aiStatus <- statusBlackWin // AI赢了,设置状态
			} else {
				g.aiStatus <- allNoneFlag // AI还没赢,玩家可以继续落子
			}
		}
	}
}

// count 棋子数量,block 两边阻挡数量只能为[0,1,2],empty 棋子中间空位数量
func (g *Gomoku) mType(count, block, empty int) int {
	// 连续棋子中间没有空位
	if empty <= 0 {
		if count >= 5 {
			// _11111_
			return scoreFive // 连续5个棋子,直接连五
		}
		if block == 0 {
			// 两边都没有另一种棋子挡住,因此根据连续X棋子数返回活X
			switch count {
			case 1:
				// _1_
				return scoreOne
			case 2:
				// _11_
				return scoreTwo
			case 3:
				// _111_
				return scoreThree
			case 4:
				// _1111_
				return scoreFour // 下一步就赢了
			}
		} else if block == 1 {
			// 连续棋子两边只有一边被挡住,连续X棋子返回眠X
			switch count {
			case 1:
				// 01_,_10
				return scoreBlockedOne
			case 2:
				// 011_,_110
				return scoreBlockedTwo
			case 3:
				// 0111_,_1110
				return scoreBlockedThree
			case 4:
				// 01111_,_11110
				return scoreBlockedFour // 敌人可以堵住
			}
		}
	} else if empty == 1 || empty == count-1 {
		// 存在1个空位
		if count >= 6 {
			// _11111_1_,_111_111_ 还有很多情况,但下一步直接赢了
			return scoreFive // 6个棋子中间空1位,直接下在空位也是连五
		}
		if block == 0 {
			// 以下情况棋子两边没有被阻挡
			switch count {
			// case 1: 存在一个空位时count不可能为1
			case 2:
				// _1_1_,连续2个棋子中间空1位,比眠二强比活二差
				return scoreTwo / 2
			case 3:
				// _11_1_,_1_11_,连续3个棋子中间空1位,活三
				return scoreThree
			case 4:
				// _111_1_,连续4个棋子中间空1位,眠四
				return scoreBlockedFour
			case 5:
				// _1111_1_,_11_111_,连续5个棋子中间空1位,活四,下一步就赢了
				return scoreFour
			}
		} else if block == 1 {
			// 两边只存在1个棋子阻挡
			switch count {
			// case 1: 存在一个空位时count不可能为1
			case 2:
				// 01_1_,_1_10,中间空位,只有一边被挡,眠二
				return scoreBlockedTwo
			case 3:
				// 011_1_,_1_110,中间空位,只有一边被挡,眠三
				return scoreBlockedThree
			case 4:
				// 011_11_,_11_110,中间空位,只有一边被挡,眠四
				return scoreBlockedFour
			case 5:
				// 0111_11_,_111_110,中间空位,只有一边被挡,可被敌人破解,降级为眠四
				return scoreBlockedFour
			}
		}
	} else if empty == 2 || empty == count-2 {
		// 中间存在2个空位
		if count >= 7 {
			// _111_1111_ 等多种情况,下一步直接落在中间空位就赢了
			return scoreFive
		}
		if block == 0 {
			switch count {
			// case 1, 2: 情况不存在
			case 3:
				// _1_1_1_,无阻挡3个中间2个空位,只有一种情况,活三
				return scoreThree
			case 4, 5:
				// _1_1_11_,_1_11_1_,_11_1_1_,无阻挡4个中间2个空位,眠四
				// _1_1_111_,_1_111_1_ 等,无阻挡5个中间2个空位,眠四
				return scoreBlockedFour
			case 6:
				// _111_1_11_ 等,无阻挡6个中间2个空位,活四
				return scoreFour
			}
		} else if block == 1 {
			switch count {
			case 3:
				// 01_1_1_,_1_1_10,3个中间2个空位1边阻挡,眠三
				return scoreBlockedThree
			case 4:
				// 011_1_1_,_11_1_10,4个中间2个空位1边阻挡,眠四
				return scoreBlockedFour
			case 5:
				// 011_11_1_,_11_11_10,5个中间2个空位1边阻挡,眠五
				return scoreBlockedFour
			case 6:
				// 011_111_1_,_11_111_10,6个中间2个空位1边阻挡,降级为活四,下一步赢了
				return scoreFour
			}
		} else if block == 2 {
			switch count {
			case 4, 5, 6:
				// 011_1_10,01_11_10,01_1_110,4个中间2个空位2边阻挡,眠四
				// 011_11_10,011_11_10,01_1_1110,5个中间2个空位2边阻挡,眠四
				// 011_111_10,011_111_10,01_11_1110,6个中间2个空位2边阻挡,眠四
				return scoreBlockedFour
			}
		}
	} else if empty == 3 || empty == count-3 {
		// 中间3个空位
		if count >= 8 {
			// _11_1_1_1111_ 等,这种情况下一步也直接赢了
			return scoreFive
		}
		if block == 0 {
			switch count {
			case 4, 5:
				// _1_1_1_1_,4个中间3个空位无阻挡只有一种情况,降级为活三
				// _11_1_1_1_ 等,5个中间3个空位无阻挡,降级为活三
				return scoreThree
			case 6:
				// _11_1_1_11_ 等,6个中间3个空位无阻挡,降级为眠四
				return scoreBlockedFour
			case 7:
				// _11_11_1_11_ 等,7个中间3个空位无阻挡,降级为活四,下两步就赢了
				return scoreFour
			}
		} else if block == 1 {
			switch count {
			case 4, 5, 6:
				// 01_1_1_1_,_1_1_1_10,4个中间3个空位1边阻挡,眠四
				// 01_11_1_1_ 等,5个中间3个空位1边阻挡,眠四
				// 01_11_11_1_ 等,6个中间3个空位1边阻挡,眠四
				return scoreBlockedFour
			case 7:
				// 01_11_11_11_ 等,7个中间3个空位1边阻挡,活四,下两步赢了
				return scoreFour
			}
		} else if block == 2 {
			switch count {
			case 4, 5, 6, 7:
				// 01_1_1_10,01_1_1_10,4个中间3个空位2边阻挡,眠四
				// 01_11_1_10 等,5个中间3个空位2边阻挡,眠四
				// 01_11_11_10 等,6个中间3个空位2边阻挡,眠四
				// 01_11_111_10 等,7个中间3个空位2边阻挡,眠四
				return scoreBlockedFour
			}
		}
	} else if empty == 4 || empty == count-4 {
		// 4个空位
		if count >= 9 {
			// _11_11_11_1_11_ 等,再下2次就赢了
			return scoreFive
		}
		if block == 0 {
			switch count {
			case 5, 6, 7, 8:
				// _1_1_1_1_1_,5个棋子4个空位无阻挡就一种情况,活四
				// _1_11_1_1_1_ 等,6个棋子4个空位无阻挡,活四
				// _1_11_11_1_1_ 等,7个棋子4个空位无阻挡,活四
				// _1_11_11_11_1_ 等,8个棋子4个空位无阻挡,活四
				return scoreFour
			}
		} else if block == 1 {
			switch count {
			case 4, 5, 6, 7:
				return scoreBlockedFour
			case 8:
				return scoreFour
			}
		} else if block == 2 {
			switch count {
			case 5, 6, 7, 8:
				return scoreBlockedFour
			}
		}
	} else if empty == 5 || empty == count-5 {
		// _1_1_1_1_1_1_ 等,中间5个空位
		return scoreFive
	}
	return 0
}

func (g *Gomoku) scorePoint(px, py, role int) float64 {
	var (
		i, t, x, y, result  int
		empty, count, block int
	)
	reset := func() {
		count = 1  // role类型棋子数量
		block = 0  // 被另一种棋子堵住数量,只能是[0,1,2]
		empty = -1 // 空位数量,初始-1用于判断
	}
	reset()

	// 从px,py向右计算
	for i = py + 1; true; i++ {
		if i >= boardSize {
			block++ // 遇到右边界,算一个阻挡
			break
		}

		if t = g.board[px][i]; t == allNoneFlag {
			if empty == -1 && i < boardSize-1 && g.board[px][i+1] == role {
				// 首次遇到空位,且空位右边是role棋子能连上
				// 此时empty为空位前面连续棋子数,例如 111_11 此时empty=3
				// 会造成后续统计棋子数量时少了空位,这时候count-empty就是空位数量
				empty = count
				continue
			} else {
				break // 连续遇到2次空位,不用在此方向继续了
			}
		} else if t == role {
			count++ // 该位置同role类型
		} else {
			block++ // 该方向遇到阻挡
			break
		}
	}

	// 从px,py向左计算
	for i = py - 1; true; i-- {
		if i < 0 {
			block++ // 遇到左边界,算一个阻挡
			break
		}

		if t = g.board[px][i]; t == allNoneFlag {
			if empty == -1 && i > 0 && g.board[px][i-1] == role {
				// 向左首次遇到空位,且空位左边是role棋子,并且上面向右没有遇到空位
				// 此时empty = 0,下一个循环开始记录棋子数量,最终count-empty就是空位数量
				empty = 0
				continue
			} else {
				break // 上面向右遇到空位,这次向左又遇到空位则直接退出
			}
		}
		if t == role {
			count++ // 另一个方向计数
			if empty != -1 {
				// 此时empty改为统计棋子数量
				// 会造成后续统计棋子数量时少了空位,这时候count-empty就是空位数量
				empty++
			}
		} else {
			block++ // 被阻挡
			break
		}
	}

	// 将得分累加
	result += g.mType(count, block, empty)

	reset() // 从px,py下计算,过程同上
	for i = px + 1; true; i++ {
		if i >= boardSize {
			block++
			break
		}

		if t = g.board[i][py]; t == allNoneFlag {
			if empty == -1 && i < boardSize-1 && g.board[i+1][py] == role {
				empty = count
				continue
			} else {
				break
			}
		}
		if t == role {
			count++
		} else {
			block++
			break
		}
	}

	// 从px,py上计算,过程同上
	for i = px - 1; true; i-- {
		if i < 0 {
			block++
			break
		}

		if t = g.board[i][py]; t == allNoneFlag {
			if empty == -1 && i > 0 && g.board[i-1][py] == role {
				empty = 0
				continue
			} else {
				break
			}
		}
		if t == role {
			count++
			if empty != -1 {
				empty++
			}
		} else {
			block++
			break
		}
	}

	result += g.mType(count, block, empty)

	reset() // 从px,py向右下计算
	for i = 1; true; i++ {
		x, y = px+i, py+i
		if x >= boardSize || y >= boardSize {
			block++
			break
		}

		if t = g.board[x][y]; t == allNoneFlag {
			if empty == -1 && (x < boardSize-1 && y < boardSize-1) && g.board[x+1][y+1] == role {
				empty = count
				continue
			} else {
				break
			}
		}
		if t == role {
			count++
		} else {
			block++
			break
		}
	}

	// 从px,py向左上计算
	for i = 1; true; i++ {
		x, y = px-i, py-i
		if x < 0 || y < 0 {
			block++
			break
		}

		if t = g.board[x][y]; t == allNoneFlag {
			if empty == -1 && (x > 0 && y > 0) && g.board[x-1][y-1] == role {
				empty = 0
				continue
			} else {
				break
			}
		}
		if t == role {
			count++
			if empty != -1 {
				empty++
			}
		} else {
			block++
			break
		}
	}

	result += g.mType(count, block, empty)

	reset() // 从px,py向左下计算
	for i = 1; true; i++ {
		x, y = px+i, py-i
		if x >= boardSize || y < 0 {
			block++
			break
		}

		if t = g.board[x][y]; t == allNoneFlag {
			if empty == -1 && (x < boardSize-1 && y > 0) && g.board[x+1][y-1] == role {
				empty = count
				continue
			} else {
				break
			}
		}
		if t == role {
			count++
		} else {
			block++
			break
		}
	}

	// 从px,py向右上计算
	for i = 1; true; i++ {
		x, y = px-i, py+i
		if x < 0 || y >= boardSize {
			block++
			break
		}

		if t = g.board[x][y]; t == allNoneFlag {
			if empty == -1 && (x > 0 && y < boardSize-1) && g.board[x-1][y+1] == role {
				empty = 0
				continue
			} else {
				break
			}
		}
		if t == role {
			count++
			if empty != -1 {
				empty++
			}
		} else {
			block++
			break
		}
	}

	// 当前位置在所有方向上分数全部累加,综合分数高最优先落子
	result += g.mType(count, block, empty)

	// 只做一件事,就是修复冲四
	if result < scoreFour && result >= scoreBlockedFour {
		if result >= scoreBlockedFour && result < (scoreBlockedFour+scoreThree) {
			return scoreThree // 单独冲四,意义不大,则将分数降至活三
		} else if result >= scoreBlockedFour+scoreThree && result < scoreBlockedFour*2 {
			return scoreFour // 冲四+活三,比双三分高,相当于自己形成活四
		} else {
			return scoreFour * 2 // 双冲四,比活四分数也高
		}
	}
	return float64(result)
}

// 启发函数
func (g *Gomoku) gen() [][]int {
	var (
		fives        [][]int
		fours        [][]int
		blockedFours [][]int
		twoThrees    [][]int
		threes       [][]int
		twos         [][]int
		neighbors    [][]int
	)

	for i := 0; i < boardSize; i++ {
		for j := 0; j < boardSize; j++ {
			// 当前位置没有落子,且该位置有邻居,对于那些周围没有棋子的空位判断时没意义
			if g.hasNeighbor(i, j) {
				switch scoreHum, scoreCom := g.humScore[i][j], g.comScore[i][j]; {
				case scoreCom >= scoreFive:
					return [][]int{{i, j}} // 先看电脑能不能连成5
				case scoreHum >= scoreFive:
					// 再看玩家能不能连成5,别急着返回,因为遍历还没完成,说不定电脑自己能成五
					fives = append(fives, []int{i, j})
				case scoreCom >= scoreFour:
					fours = append([][]int{{i, j}}, fours...) // 对电脑有利放前面
				case scoreHum >= scoreFour:
					fours = append(fours, []int{i, j}) // 对玩家有利放后面,下面同理
				case scoreCom >= scoreBlockedFour:
					blockedFours = append([][]int{{i, j}}, blockedFours...)
				case scoreHum >= scoreBlockedFour:
					blockedFours = append(blockedFours, []int{i, j})
				case scoreCom >= 2*scoreThree: // 能成双三也行
					twoThrees = append([][]int{{i, j}}, twoThrees...)
				case scoreHum >= 2*scoreThree:
					twoThrees = append(twoThrees, []int{i, j})
				case scoreCom >= scoreThree:
					threes = append([][]int{{i, j}}, threes...)
				case scoreHum >= scoreThree:
					threes = append(threes, []int{i, j})
				case scoreCom >= scoreTwo:
					twos = append([][]int{{i, j}}, twos...)
				case scoreHum >= scoreTwo:
					twos = append(twos, []int{i, j})
				default:
					neighbors = append(neighbors, []int{i, j})
				}
			}
		}
	}

	if len(fives) > 0 {
		return fives // 如果成五,是必杀棋,直接返回
	}

	// 注意一个活三可以有两个位置形成活四,但是不能只考虑其中一个,要从多个中考虑更好的选择
	// 所以不能碰到活四就返回第一个,应该需要考虑多个
	if len(fours) > 0 {
		return fours
	}

	// 冲四+活三
	if len(blockedFours) > 0 {
		return [][]int{blockedFours[0]}
	}

	// 双三很特殊,因为能形成双三的不一定比一个活三强
	if len(twoThrees) > 0 {
		return append(twoThrees, threes...)
	}

	// 剩余情况连接起来进行返回
	result := append(threes, twos...)
	result = append(result, neighbors...)
	if len(result) > g.countLimit {
		return result[:g.countLimit]
	}
	return result
}

// 棋面估分
func (g *Gomoku) evaluate(role int) float64 {
	comMaxScore, humMaxScore := scoreFiveNeg, scoreFiveNeg

	// 遍历出最高分,开销不大
	for i := 0; i < boardSize; i++ {
		for j := 0; j < boardSize; j++ {
			if g.board[i][j] == allNoneFlag {
				if g.comScore[i][j] > comMaxScore {
					comMaxScore = g.comScore[i][j]
				}
				if g.humScore[i][j] > humMaxScore {
					humMaxScore = g.humScore[i][j]
				}
			}
		}
	}

	if role == comImgFlag {
		return comMaxScore - humMaxScore
	}
	return humMaxScore - comMaxScore
}

func (g *Gomoku) max(deep, role int, alpha, beta float64) float64 {
	var v float64
	c, ok := g.cache[g.zobristCode]
	if ok {
		if c.deep >= deep {
			return c.score // 得到更深层次相同局面,直接返回分数
		}
		v = c.maxScore // 得到该状态下最大值,无需重复计算
	} else {
		v = g.evaluate(role)
		c = &gomokuCache{maxScore: v}
	}

	if deep <= 0 || g.greatOrEqualThan(v, scoreFive) {
		return v // 到达深度或棋面估分大于连五,直接返回
	}

	switchRole := func(r int) int {
		if r == comImgFlag {
			return humImgFlag
		}
		return comImgFlag
	}
	maxFloat64 := func(a, b float64) float64 {
		if a > b {
			return a
		}
		return b
	}

	var (
		best   = scoreMin
		points = g.gen()
	)
	for _, p := range points {
		g.put(p[0], p[1], role) // 假设role棋子落下

		v = -g.max(deep-1, switchRole(role), -beta, -maxFloat64(alpha, best)) * g.deepDecrease

		g.remove(p[0], p[1]) // 取消role落子

		if g.greatOrEqualThan(v, beta) { // AB 剪枝
			c.deep, c.score = deep, v
			g.cache[g.zobristCode] = c
			return v
		}
		if g.greatThan(v, best) {
			best = v
		}
	}
	// todo 可以计算特殊算杀逻辑

	c.deep, c.score = deep, best
	g.cache[g.zobristCode] = c
	return best
}

func (g *Gomoku) maxMin(deep int) (x, y int, score float64) {
	var (
		best       = scoreMin
		points     = g.gen()
		bestPoints [][]int
	)
	for _, p := range points {
		g.put(p[0], p[1], comImgFlag) // 电脑先手

		// 进入极大值递归算法,这次玩家后手
		v := -g.max(deep-1, humImgFlag, scoreMin, -best)

		g.remove(p[0], p[1]) // 去掉电脑落子

		// 边缘棋子的话,要把分数打折,避免电脑总喜欢往边上走
		if p[0] < 3 || p[0] > 11 || p[1] < 3 || p[1] > 11 {
			v = 0.5 * v
		}

		if g.greatThan(v, best) {
			best = v // 找到一个更好的分,清除之前结果
			bestPoints = [][]int{p}
		} else if g.equal(v, best) {
			// 找到一样好的分,则将分数添加
			bestPoints = append(bestPoints, p)
		}
	}
	// 按照正常情况 len(bestPoints) > 0,随机选择位置,避免被发现规律
	p := bestPoints[rand.Intn(len(bestPoints))]
	return p[0], p[1], best
}

// 只更新一个点附近的分数
func (g *Gomoku) updateScore(x, y int) {
	update := func(x, y int) {
		g.comScore[x][y] = g.scorePoint(x, y, comImgFlag)
		g.humScore[x][y] = g.scorePoint(x, y, humImgFlag)
	}

	var i, tx, ty int
	const radius = 8 // 计算x,y半径范围内数据

	// - 左右计算
	for i = -radius; i < radius; i++ {
		if tx, ty = x, y+i; ty >= 0 {
			if ty >= boardSize {
				break
			}
			if g.board[tx][ty] == allNoneFlag {
				update(tx, ty)
			}
		}
	}

	// | 上下计算
	for i = -radius; i < radius; i++ {
		if tx, ty = x+i, y; tx >= 0 {
			if tx >= boardSize {
				break
			}
			if g.board[tx][ty] == allNoneFlag {
				update(tx, ty)
			}
		}
	}

	// \ 左上右下计算
	for i = -radius; i < radius; i++ {
		if tx, ty = x+i, y+i; tx >= 0 && ty >= 0 {
			if tx >= boardSize || ty >= boardSize {
				break
			}
			if g.board[tx][ty] == allNoneFlag {
				update(tx, ty)
			}
		}
	}

	// / 左下右上计算
	for i = -radius; i < radius; i++ {
		if tx, ty = x+i, y-i; tx >= 0 && ty >= 0 {
			if tx >= boardSize || ty >= boardSize {
				break // todo 验证有没有问题
			}
			if g.board[tx][ty] == allNoneFlag {
				update(tx, ty)
			}
		}
	}
}

func (g *Gomoku) hasNeighbor(x, y int) bool {
	if g.board[x][y] == allNoneFlag {
		// x,y周围distance距离存在不含自身的cnt个已经下的棋子则返回true
		sx, ex := x-2, x+2
		sy, ey := y-2, y+2
		for i := sx; i <= ex; i++ {
			if i >= 0 && i < boardSize {
				for j := sy; j <= ey; j++ {
					if j >= 0 && j < boardSize && (i != x || j != y) &&
						g.board[i][j] != allNoneFlag {
						return true // 当前x,y位置是空位,周围2格至少存在1个棋子
					}
				}
			}
		}
	}
	return false
}

// 下面是根据阈值计算相关大小的方法
func (g *Gomoku) equal(a, b float64) bool {
	return (a*g.threshold >= b) && (a <= b*g.threshold)
}
func (g *Gomoku) littleThan(a, b float64) bool {
	return a*g.threshold <= b
}
func (g *Gomoku) littleOrEqualThan(a, b float64) bool {
	return a <= b*g.threshold
}
func (g *Gomoku) greatThan(a, b float64) bool {
	return a >= b*g.threshold
}
func (g *Gomoku) greatOrEqualThan(a, b float64) bool {
	return a*g.threshold >= b
}
