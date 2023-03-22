package main

import (
	"fmt"

	"github.com/jan-bar/LittleGame/runAI"
)

/*
博弈算法: https://blog.csdn.net/fsdev/category_1085675.html

象棋巫师: https://www.xqbase.com/index.htm
游戏源码: https://github.com/xqbase/xqwlight

教程: https://www.cnblogs.com/royhoo/p/6426394.html

本AI通过UCCI协议,调用ELEEYE.EXE可执行程序执行AI逻辑

永久开源免费,且号称棋力很强
皮卡鱼引擎: https://github.com/official-pikafish/Pikafish ,注意这个引擎用 uci 而不是 ucci

ucci 引擎启动后第一条指令,只有收到这个指令才能进行后续

isready, 返回: readyok, 看注释这个命令没啥用

setoption batch [on | true]  批处理模式,不停止中断
setoption debug [on | true]  调试模式
setoption ponder [on | true] 增加后台思考时间
setoption usehash [off | false] 使用置换表裁剪
setoption usebook [off | false] 使用开局库
setoption useegtb [off | false] 源码中没发现用法
setoption bookfiles [path] 开局库文件
setoption egtbpaths [path] 源码中没发现用法
setoption hashsize [int(0,1024)] 置换表大小
setoption threads [int(0,32)] 源码中没发现用法
setoption promotion [on | true] 可升变时,没受威胁的仕(士)和相(象)重新计算分值
setoption idle [none(bIdle=false,nCountMask=0xfff) | small(bIdle=true,nCountMask=0x3ff) | medium(bIdle=true,nCountMask=0xff) | large(bIdle=true,nCountMask=0x3f)] default(none)
  搜索若干结点后调用中断,bIdle=true时中断后会Sleep(1)
setoption pruning [none(否) | small(是) | medium(是) | large(是)] default(large) 是否空着裁剪
setoption knowledge [none(否) | small(是) | medium(是) | large(是)] default(large) 使用局面评价知识
setoption randomness [none(0) | tiny(1) | small(3) | medium(7) | large(15) | huge(31)] default(none)
  随机性屏蔽位,适当增大随机性,可以避免相同局面走出一样的着法
setoption style [solid | normal | risky] default(normal) 源码中没发现用法
setoption newgame 没啥意义


position {<special_position> | fen <fen_string>} [moves <move_list>] 空闲时设置走法
probe {<special_position> | fen <fen_string>} [moves <move_list>]    探测走法
指定fen字符串并带上移动位置
position fen rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w moves b2c2
使用默认开局带上移动位置
position startpos moves b2c2
probe startpos moves b2c2 b9a7 探测走法

只传移动位置
"banmoves <move_list> "指令,处理起来和"position ... moves ..."

执行搜索,根据各种条件限定搜索深度
go [ponder(后台思考) | draw(提和)] <mode[depth(0,32){默认值} | nodes(0, 2000000000) | time(0, 2000000000)]>
go time 1000 [movestogo(1, 999){时段制} | increment(0, 999999){加时制}]

stop 当长时间思考时可以停止并返回
quit 退出指令

上面是空闲时会响应的命令,下面是思考时才会响应的命令
isready
ponderhit draw 指令启动计时功能,并设置提和标志
ponderhit 指令启动计时功能,如果"SearchMain()"例程认为已经搜索了足够的时间,那么发出中止信号
stop
quit
probe 探测走法

(1) 支持的UCCI命令有：
  ucci
  setoption ...
  position {fen <fen_str> | startpos} [moves <move_list>]
  banmoves <move_list>
  go [ponder | draw] ...
  ponderhit [draw] | stop
  probe {fen <fen_str> | startpos} [moves <move_list>]
  quit
(2) 可以返回的UCCI信息有：
  id {name <engine_name> | version <version_name> | copyright <copyright_info> | author <author_name> | user <user_name>}
  option ... 显示配置
  ucciok     表示显示ucci命令所有信息
  info ...   提示信息

  由go开始搜索,有最好的则返回,过程中使用stop返回当时最好走法
  {nobestmove | bestmove <best_move(最好走法)> [ponder <ponder_move(思考走法)>] [draw(提和) | resign(认输)]}

  由probe指令得到结果
    pophash [bestmove <best_move(最好走法)>] [lowerbound <value> depth <depth>] [upperbound <value> depth <depth>]

  bye 退出提示语


开始新棋局发送下面2个命令
setoption newgame
setoption ponder false
[position startpos, go nodes xxx] ai先手时多发送这两个指令


每走一步棋就发送下面2条指令,moves 后面根全部历史走法
position startpos moves xxx...
go nodes xxx
收到下面数据时,表示查到最好走法
bestmove g6g5 ponder a0a1 , nobestmove(没有找到走法)

*/

func (g *chessGame) ai() {
	defer g.aiStatus.Store(aiPlay) // 设置状态,ai落子

	g.goNext(&g.move, nil)
}

func (g *chessGame) goNext(mv *moveXY, ms *string) {
	g.eleeye.Send(g.position.String(), nil) // 发送当前局面

	move := "bestmove" // 启动ai引擎思考,接收结果
	g.eleeye.Send(g.goCommand, runAI.MatchLineContains(&move))

	var a, b, c, d byte
	fmt.Sscanf(move, "bestmove %c%c%c%c ", &a, &b, &c, &d)

	if mv != nil {
		mv.y0 = int(a - 'a')
		mv.x0 = int('9' - b)
		mv.y1 = int(c - 'a')
		mv.x1 = int('9' - d)
	}

	if ms != nil {
		*ms = string([]byte{a, b, c, d})
	}
}
