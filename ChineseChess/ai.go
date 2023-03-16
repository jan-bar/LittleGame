package main

import (
	"sort"
	"time"
)

/*
博弈算法: https://blog.csdn.net/fsdev/category_1085675.html

象棋巫师: https://www.xqbase.com/index.htm
游戏源码: https://github.com/xqbase/xqwlight

教程: https://www.cnblogs.com/royhoo/p/6426394.html
*/

func (g *chessGame) ai() {
	defer g.aiStatus.Store(aiPlay) // 设置状态,ai落子

	g.distance = 0
	g.aiPlayer = true
	var (
		ts       = time.Now()
		i, value int
	)

	for j := uint32(0); j <= hashMask; j++ {
		ht, ok := g.hashTable[j]
		if ok {
			ht.depth = 0 // 重置置换表
		} else {
			g.hashTable[j] = &hashTable{depth: 0}
		}
	}

	for i = 0; i < limitMaxDepth; i++ {
		kt, ok := g.killerTable[i]
		if ok {
			kt[0].x0 = -1 // 重置杀手走法表
		} else {
			g.killerTable[i] = new([2]moveXY)
		}
	}

	for i = range g.historyTable {
		g.historyTable[i] = 0 // 重置历史表
	}

	// 限定最大搜索深度,迭代加深会用历史表提高效率
	for i = 1; i <= limitMaxDepth; i++ {
		value = g.searchFull(-mateValue, mateValue, i, false)
		if time.Since(ts) > time.Second {
			break // 时间用完了,不再搜索
		}
		if value > winValue || value < -winValue {
			break // 胜负已分,不用继续搜索
		}
	}
}

const (
	mateValue      = 10000           // 最高分值
	banValue       = mateValue - 100 // 长将判负的分值
	winValue       = mateValue - 200 // 赢棋分值(高于此分值都是赢棋)
	drawValue      = 20              // 和棋时返回的分数(取负值)
	nullSafeMargin = 400             // 空步裁剪有效的最小优势
	nullOKeyMargin = 200             // 可以进行空步裁剪的最小优势
	advancedValue  = 3               // 先行权分值
	limitMaxDepth  = 64              // 搜索最大深度
	nullDepth      = 2               // 空步搜索多减去的搜索值
)

/*
walk:
  true:  搜索黑棋走法
  false: 搜索红棋走法
*/
func (g *chessGame) searchFull(vlAlpha, vlBeta, depth int, noNull bool) int {
	mvHash := moveXY{x0: -1}
	if g.distance > 0 {
		// 1. 到达水平线,则调用静态搜索(注意: 由于空步裁剪,深度可能小于零)
		if depth <= 0 {
			return g.searchQuiesce(vlAlpha, vlBeta)
		}

		vlRep := g.repStatus(1)
		if vlRep > 0 {
			return g.repValue(vlRep)
		}

		// 尝试置换表
		vlRep = g.probeHash(vlAlpha, vlBeta, depth, &mvHash)
		if vlRep > -mateValue {
			return vlRep
		}

		// 1-2. 到达极限深度就返回局面评价
		if g.distance == limitMaxDepth {
			return g.evaluate()
		}

		// 1-3. 尝试空步裁剪(根节点的Beta值是"MATE_VALUE"，所以不可能发生空步裁剪)
		if !noNull && !g.inCheck() && g.nullOkay() {
			g.nullMove()
			vlRep = -g.searchFull(-vlBeta, 1-vlBeta, depth-nullDepth-1, true)
			g.undoNullMove()
			if vlRep >= vlBeta && (g.nullSafe() ||
				g.searchFull(vlAlpha, vlBeta, depth-nullDepth, true) >= vlBeta) {
				return vlRep
			}
		}
	}

	// 2. 初始化最佳值和最佳走法
	var (
		hashFlag = hashAlpha // 节点类型
		vlBest   = -mateValue
		mvBest   = moveXY{x0: -1}
		v        moveXY
		vl       int
	)

	if g.moveSort(mvHash) {
		return g.mateValue() // 没棋了
	}
	for {
		if v = g.next(); v.x0 < 0 {
			if v.x0 == -2 {
				return g.mateValue() // 没棋了
			}
			break
		}

		sp, dp := g.board[v.x0][v.y0], g.board[v.x1][v.y1]
		g.board[v.x1][v.y1] = sp
		g.board[v.x0][v.y0] = 0
		g.makeMove(v, sp, dp) // 尝试走法,更新分数

		newDepth := depth
		if !g.inCheck() {
			newDepth--
		}
		// 递归调用自身,切换红黑棋,Alpha和Beta调换位置,返回负分
		vl = -g.searchFull(-vlBeta, -vlAlpha, newDepth, false)

		g.board[v.x0][v.y0], g.board[v.x1][v.y1] = sp, dp
		g.undoMakeMove(v, sp, dp) // 恢复走法,恢复分数

		// 5. 进行Alpha-Beta大小判断和截断
		if vl > vlBest {
			vlBest = vl
			if vl >= vlBeta {
				hashFlag = hashBeta
				mvBest = v
				break
			}

			if vl > vlAlpha {
				vlAlpha = vl
				hashFlag = hashPv
				mvBest = v

				if g.distance == 0 {
					g.chessMove = v
				}
			}
		}
	}

	if vlBest == -mateValue {
		return g.mateValue()
	}

	// 记录到置换表
	g.recordHash(hashFlag, vlBest, depth, mvBest)
	if mvBest.x0 >= 0 {
		// 找到好的走法,更新历史表
		g.setBestMove(mvBest, depth)
	}

	return vlBest
}

type sortMoveXY struct {
	mvs []moveXY
	vls []int
}

func (m *sortMoveXY) Len() int {
	return len(m.mvs)
}
func (m *sortMoveXY) Less(i, j int) bool {
	return m.vls[i] > m.vls[j] // 此时用vls进行排序
}
func (m *sortMoveXY) Swap(i, j int) {
	m.mvs[i], m.mvs[j] = m.mvs[j], m.mvs[i]
	m.vls[i], m.vls[j] = m.vls[j], m.vls[i]
}

const (
	phaseHash     = 0
	phaseKiller1  = 1
	phaseKiller2  = 2
	phaseGenMoves = 3
	phaseRest     = 4
)

func (g *chessGame) moveSort(mvHash moveXY) bool {
	g.mvs = g.mvs[:0]
	g.vls = g.vls[:0]
	g.mvHash.x0 = -1
	g.mvKiller1.x0 = -1
	g.mvKiller2.x0 = -1
	g.phase = phaseHash
	g.index = 0
	g.singleReply = false

	if g.inCheck() {
		g.phase = phaseRest

		if g.canStep(g.aiPlayer, &g.mvs, nil) {
			return true // 没棋了
		}
		for _, mv := range g.mvs {
			// 要使用置换表启发,把置换表中的走法排在最前面
			if mv == mvHash {
				g.vls = append(g.vls, 0x7fffffff)
			} else {
				g.vls = append(g.vls, g.historyTable[historyIndex(mv)])
			}
		}
		sort.Sort(&sortMoveXY{mvs: g.mvs, vls: g.vls})
		g.singleReply = len(g.mvs) == 1 // 只有1个回棋
	} else {
		g.mvHash = mvHash
		g.mvKiller1 = g.killerTable[g.distance][0]
		g.mvKiller2 = g.killerTable[g.distance][1]
	}
	return false
}

func (g *chessGame) next() moveXY {
	switch g.phase {
	case phaseHash:
		g.phase = phaseKiller1
		if g.mvHash.x0 >= 0 {
			return g.mvHash
		}
		fallthrough
	case phaseKiller1:
		g.phase = phaseKiller2
		if g.mvKiller1 != g.mvHash && g.mvKiller1.x0 >= 0 &&
			g.canNext(g.mvKiller1.x0, g.mvKiller1.y0, g.mvKiller1.x1, g.mvKiller1.y1) {
			return g.mvKiller1
		}
		fallthrough
	case phaseKiller2:
		g.phase = phaseGenMoves
		if g.mvKiller2 != g.mvHash && g.mvKiller2.x0 >= 0 &&
			g.canNext(g.mvKiller2.x0, g.mvKiller2.y0, g.mvKiller2.x1, g.mvKiller2.y1) {
			return g.mvKiller2
		}
		fallthrough
	case phaseGenMoves:
		g.phase = phaseRest

		g.mvs = g.mvs[:0]
		if g.canStep(g.aiPlayer, &g.mvs, nil) {
			return moveXY{x0: -2}
		}
		g.vls = g.vls[:0]
		for _, mv := range g.mvs {
			g.vls = append(g.vls, g.historyTable[historyIndex(mv)])
		}
		sort.Sort(&sortMoveXY{mvs: g.mvs, vls: g.vls})
		g.index = 0
		fallthrough
	default:
		for g.index < len(g.mvs) {
			mv := g.mvs[g.index]
			g.index++
			if mv != g.mvHash && mv != g.mvKiller1 && mv != g.mvKiller2 {
				return mv
			}
		}
	}
	return moveXY{x0: -1}
}

func historyIndex(m moveXY) int {
	// 根据走法,得到一个索引值,最大值为 0x99aa
	return m.y0<<12 | m.y1<<8 | m.x0<<4 | m.x1
}

// 静态(Quiescence)搜索
func (g *chessGame) searchQuiesce(vlAlpha, vlBeta int) int {
	vl := g.mateValue()
	if vl >= vlBeta {
		return vl
	}

	vlRep := g.repStatus(1)
	if vlRep > 0 {
		return g.repValue(vlRep)
	}

	if g.distance == limitMaxDepth {
		return g.evaluate()
	}

	var (
		vlBest = -mateValue
		mvs    []moveXY
		vls    []int
	)
	if g.inCheck() {
		// 5. 如果被将军，则生成全部走法
		if g.canStep(g.aiPlayer, &mvs, nil) {
			return g.mateValue() // 没棋了
		}
		for _, mv := range mvs {
			vls = append(vls, g.historyTable[historyIndex(mv)])
		}
		// 根据vls排序,且vls也要进行排序
		sort.Sort(&sortMoveXY{vls: vls, mvs: mvs})
	} else {
		// 6. 如果不被将军，先做局面评价
		vl = g.evaluate()
		if vl > vlBest {
			if vl >= vlBeta {
				return vl
			}
			vlBest = vl
			if vl > vlAlpha {
				vlAlpha = vl
			}
		}

		// 7. 如果局面评价没有截断，再生成吃子走法
		if g.canStep(g.aiPlayer, &mvs, &vls) {
			return g.mateValue() // 没棋了
		}
		// 根据vls排序,且vls也要进行排序
		sort.Sort(&sortMoveXY{vls: vls, mvs: mvs})
		for i, mv := range mvs {
			if vls[i] < 10 || (vls[i] < 20 && g.homeHalf(mv)) {
				mvs = mvs[:i] // 棋子过少的话不搜索了
				break
			}
		}
	}

	for _, v := range mvs {
		sp, dp := g.board[v.x0][v.y0], g.board[v.x1][v.y1]
		g.board[v.x1][v.y1] = sp
		g.board[v.x0][v.y0] = 0
		g.makeMove(v, sp, dp) // 尝试走法,更新分数

		// 递归调用自身,切换红黑棋,Alpha和Beta调换位置,返回负分
		vl = -g.searchQuiesce(-vlBeta, -vlAlpha)

		g.board[v.x0][v.y0], g.board[v.x1][v.y1] = sp, dp
		g.undoMakeMove(v, sp, dp) // 恢复走法,恢复分数

		// 9. 进行Alpha-Beta大小判断和截断
		if vl > vlBest { // 找到最佳值
			if vl >= vlBeta { // 找到一个Beta走法
				return vl // Beta截断
			}
			vlBest = vl // "vlBest"就是目前要返回的最佳值，可能超出Alpha-Beta边界
			if vl > vlAlpha {
				vlAlpha = vl // 缩小Alpha-Beta边界
			}
		}
	}

	if vlBest == -mateValue {
		return g.mateValue()
	}
	return vlBest
}

// 求MVV/LVA值
func mvvLva(sp, dp uint8) int { return mvvValue[dp][0] - mvvValue[sp][1] }

func (g *chessGame) homeHalf(m moveXY) bool {
	if g.aiPlayer {
		return m.x1 <= 4 // 黑棋没过河返回true
	}
	return m.x1 >= 5 // 红棋没过河返回true
}

// 判断是否重复局面
func (g *chessGame) repStatus(recur int) (res int) {
	var (
		selfSide     = false
		perpCheck    = true
		oppPerpCheck = true
		index        = len(g.mvList) - 1
	)
	for g.mvList[index].x0 >= 0 && g.pcList[index] == 0 {
		if selfSide {
			perpCheck = perpCheck && g.chkList[index]

			if g.keyList[index] == g.zobristKey {
				if recur--; recur == 0 {
					res = 1
					if perpCheck {
						res += 2
					}
					if oppPerpCheck {
						res += 4
					}
					return
				}
			}
		} else {
			oppPerpCheck = oppPerpCheck && g.chkList[index]
		}
		selfSide = !selfSide
		index--
	}
	return
}
func (g *chessGame) repValue(rep int) int {
	var vl int
	if (rep & 2) != 0 {
		vl = g.banValue()
	}
	if (rep & 4) != 0 {
		vl -= g.banValue()
	}
	if vl == 0 {
		return g.drawValue()
	}
	return vl
}
func (g *chessGame) getHashItem() *hashTable {
	return g.hashTable[g.zobristKey&hashMask]
}
func (g *chessGame) probeHash(vlAlpha, vlBeta, depth int, mvHash *moveXY) int {
	hash := g.getHashItem()
	if hash.zobristLock != g.zobristLock {
		mvHash.x0 = -1
		return -mateValue
	}

	*mvHash = hash.mv
	mate := false
	if hash.vl > winValue {
		hash.vl -= g.distance
		mate = true
	} else if hash.vl < -winValue {
		hash.vl += g.distance
		mate = true
	}

	if hash.depth < depth && !mate {
		return -mateValue
	}

	if hash.flag == hashBeta {
		if hash.vl >= vlBeta {
			return hash.vl
		}
		return -mateValue
	}

	if hash.flag == hashAlpha {
		if hash.vl <= vlAlpha {
			return hash.vl
		}
		return -mateValue
	}

	return hash.vl
}
func (g *chessGame) recordHash(flag, vl, depth int, mv moveXY) {
	hash := g.getHashItem()
	// 深度优先覆盖原则
	if hash.depth > depth {
		return
	}

	hash.flag = flag
	hash.depth = depth
	if vl > winValue {
		hash.vl = vl + g.distance
	} else if vl < -winValue {
		hash.vl = vl - g.distance
	} else if vl == g.drawValue() && mv.x0 == -1 {
		return
	} else {
		hash.vl = vl
	}
	hash.mv = mv
	hash.zobristLock = g.zobristLock
}
func (g *chessGame) setBestMove(m moveXY, depth int) {
	g.historyTable[historyIndex(m)] += depth * depth
}
func (g *chessGame) drawValue() int {
	if (g.distance & 1) == 0 {
		return -drawValue
	}
	return drawValue
}
func (g *chessGame) banValue() int {
	return g.distance - banValue
}
func (g *chessGame) mateValue() int {
	return g.distance - mateValue
}
func (g *chessGame) evaluate() int {
	if g.aiPlayer { // 计算分数, advancedValue 表示先手优势
		return g.vlBlack - g.vlRed + advancedValue
	}
	return g.vlRed - g.vlBlack + advancedValue
}
func (g *chessGame) changeSide() {
	g.aiPlayer = !g.aiPlayer
	g.zobristKey ^= PreGenZobristKeyPlayer
	g.zobristLock ^= PreGenZobristLockPlayer
}
func (g *chessGame) addPiece(x, y int, p uint8, del ...bool) {
	pv := int(pieceValue[p][x][y])
	if len(del) > 0 && del[0] {
		pv = -pv
	}
	// 仅更新分数,移动棋子交给调用方处理
	if isRed(p) {
		g.vlRed += pv
	} else {
		g.vlBlack += pv
	}
	g.zobristKey ^= PreGenZobristKeyTable[p][x][y]
	g.zobristLock ^= PreGenZobristLockTable[p][x][y]
}

// 某步走过的棋是否被将军
func (g *chessGame) inCheck() bool {
	return g.chkList[len(g.chkList)-1]
}

// 当前局面的优势是否足以进行空步搜索
func (g *chessGame) nullOkay() bool {
	if g.aiPlayer {
		return g.vlBlack > nullOKeyMargin
	}
	return g.vlRed > nullOKeyMargin
}

// 空步搜索得到的分值是否有效
func (g *chessGame) nullSafe() bool {
	if g.aiPlayer {
		return g.vlBlack > nullSafeMargin
	}
	return g.vlRed > nullSafeMargin
}
func (g *chessGame) nullMove() {
	g.mvList = append(g.mvList, moveXY{x0: -1})
	g.pcList = append(g.pcList, 0)
	g.keyList = append(g.keyList, g.zobristKey)
	g.changeSide()
	g.chkList = append(g.chkList, false)
	g.distance++
}
func (g *chessGame) undoNullMove() {
	g.distance--
	g.chkList = g.chkList[:len(g.chkList)-1]
	g.changeSide()
	g.keyList = g.keyList[:len(g.keyList)-1]
	g.pcList = g.pcList[:len(g.pcList)-1]
	g.mvList = g.mvList[:len(g.mvList)-1]
}
func (g *chessGame) makeMove(m moveXY, sp, dp uint8) {
	tk := g.zobristKey // 缓存局面信息

	g.pcList = append(g.pcList, dp)
	if dp > 0 {
		g.addPiece(m.x1, m.y1, dp, true)
	}
	g.addPiece(m.x0, m.y0, sp, true)
	g.addPiece(m.x1, m.y1, sp)
	g.mvList = append(g.mvList, m)
	g.keyList = append(g.keyList, tk)
	g.changeSide()
	g.chkList = append(g.chkList, g.isJiang(g.aiPlayer))
	g.distance++ // 增加搜索深度
}
func (g *chessGame) undoMakeMove(m moveXY, sp, dp uint8) {
	g.chkList = g.chkList[:len(g.chkList)-1]
	g.changeSide()
	g.keyList = g.keyList[:len(g.keyList)-1]

	g.mvList = g.mvList[:len(g.mvList)-1]
	g.addPiece(m.x1, m.y1, sp, true)
	g.addPiece(m.x0, m.y0, sp)

	g.pcList = g.pcList[:len(g.pcList)-1]
	if dp > 0 {
		g.addPiece(m.x1, m.y1, dp)
	}
	g.distance-- // 减少搜索深度
}
