package main

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"image/png"
	"io"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

//go:embed resources.zip
var resources []byte

func (g *chessGame) loadResources() error {
	var (
		data     = bytes.NewReader(resources)
		audioCtx = audio.NewContext(48000)
	)

	zr, err := zip.NewReader(data, data.Size())
	if err != nil {
		return err
	}

	//goland:noinspection SpellCheckingInspection 文件名和资源对应关系
	resMap := map[string]uint8{
		"ChessBoard.png": imgChessBoard,
		"Select.png":     imgSelect,
		"RedShuai.png":   imgRedShuai,
		"RedShi.png":     imgRedShi,
		"RedXiang.png":   imgRedXiang,
		"RedMa.png":      imgRedMa,
		"RedJu.png":      imgRedJu,
		"RedPao.png":     imgRedPao,
		"RedBing.png":    imgRedBing,
		"BlackJiang.png": imgBlackJiang,
		"BlackShi.png":   imgBlackShi,
		"BlackXiang.png": imgBlackXiang,
		"BlackMa.png":    imgBlackMa,
		"BlackJu.png":    imgBlackJu,
		"BlackPao.png":   imgBlackPao,
		"BlackBing.png":  imgBlackBing,
		"Select.wav":     musicSelect,
		"Put.wav":        musicPut,
		"Eat.wav":        musicEat,
		"Jiang.wav":      musicJiang,
		"GameWin.wav":    musicGameWin,
		"GameLose.wav":   musicGameLose,
		"book.dat":       0,
	}
	for _, f := range zr.File {
		i, ok := resMap[f.Name]
		if !ok {
			continue
		}

		err = func() error {
			fr, err := f.Open()
			if err != nil {
				return err
			}
			//goland:noinspection GoUnhandledErrorResult
			defer fr.Close()

			switch filepath.Ext(f.Name) {
			case ".png":
				img, err := png.Decode(fr)
				if err != nil {
					return err
				}
				g.images[i] = ebiten.NewImageFromImage(img)
			case ".wav":
				wr, err := wav.DecodeWithSampleRate(audioCtx.SampleRate(), fr)
				if err != nil {
					return err
				}
				wd, err := io.ReadAll(wr)
				if err != nil {
					return err
				}
				g.audios[i] = audioCtx.NewPlayerFromBytes(wd)
			case ".dat":
				// todo 开局库解析
			}
			return nil
		}()
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *chessGame) loadFEN(fen string) {
	// fen介绍: https://www.xqbase.com/protocol/cchess_fen.htm
	// king,advisor,bishop,knight,rook,cannon,pawn
	// 红: 帅-K,仕-A,相-B,马-N,车-R,炮-C,兵-P
	// 黑: 将-k,仕-a,相-b,马-n,车-r,炮-c,兵-p
	// 数字代表空位数量,"w"代表红方走,"b"代表黑方走,两个"-"在中国象棋中没有意义
	// 表示双方没有吃子的走棋步数(半回合数),通常该值达到120就要判和(六十回合自然限着),一旦形成局面的上一步是吃子,这里就标记"0"
	// 最后一个数字表示回合数, 示例请看: boardStart
	// 本逻辑只确保解析不会报错,没有判断棋局是否正确

	var i, j, p uint8
	for i = 0; i < boardX; i++ {
		for j = 0; j < boardY; j++ {
			g.board[i][j] = 0 // 先清空棋盘
		}
	}

	i, j = 0, 0
	g.redPlayer = true // 默认红棋先行
	for is := 0; is < len(fen); is++ {
		p = 0

		switch fen[is] {
		case ' ':
			if is++; is < len(fen) && fen[is] == 'b' {
				g.redPlayer = false // 只有这种情况轮到黑棋
			}
			return
		case '/':
			i++
			j = 0 // 换行
		case '1', '2', '3', '4', '5', '6', '7', '8', '9':
			j += fen[is] - '0' // 跳过空行
		case 'K':
			p = imgRedShuai
		case 'k':
			p = imgBlackJiang
		case 'A':
			p = imgRedShi
		case 'a':
			p = imgBlackShi
		case 'B':
			p = imgRedXiang
		case 'b':
			p = imgBlackXiang
		case 'N':
			p = imgRedMa
		case 'n':
			p = imgBlackMa
		case 'R':
			p = imgRedJu
		case 'r':
			p = imgBlackJu
		case 'C':
			p = imgRedPao
		case 'c':
			p = imgBlackPao
		case 'P':
			p = imgRedBing
		case 'p':
			p = imgBlackBing
		}

		if p > 0 && i < boardX && j < boardY {
			g.board[i][j] = p
			j++ // 合法数据的情况
		}
	}
}

func (g *chessGame) storeFEN() string {
	var (
		i, j, cur int
		// 棋盘所有格子数,再加boardY个'/'
		s    = make([]byte, (boardX+1)*boardY)
		temp uint8
	)
	for i = 0; i < boardX; i++ {
		if i > 0 {
			s[cur] = '/'
			cur++
		}

		for j = 0; j < boardY; j++ {
			switch g.board[i][j] {
			case 0:
				temp = 0
			case imgRedShuai:
				temp = 'K'
			case imgBlackJiang:
				temp = 'k'
			case imgRedShi:
				temp = 'A'
			case imgBlackShi:
				temp = 'a'
			case imgRedXiang:
				temp = 'B'
			case imgBlackXiang:
				temp = 'b'
			case imgRedMa:
				temp = 'N'
			case imgBlackMa:
				temp = 'n'
			case imgRedJu:
				temp = 'R'
			case imgBlackJu:
				temp = 'r'
			case imgRedPao:
				temp = 'C'
			case imgBlackPao:
				temp = 'c'
			case imgRedBing:
				temp = 'P'
			case imgBlackBing:
				temp = 'p'
			default:
				temp = 0xff
			}

			if temp != 0xff {
				s[cur] = temp
				cur++
			}
		}
	}

	cur, temp = 0, 0
	for i = 0; i < len(s); i++ {
		if s[i] == 0 {
			temp++ // 统计连续空位数
		} else {
			if temp > 0 {
				s[cur] = '0' + temp
				temp = 0
				cur++
			}
			s[cur] = s[i]
			cur++
		}
	}

	if g.redPlayer {
		s = append(s[:cur], " w - - 0 1"...)
	} else {
		s = append(s[:cur], " b - - 0 1"...)
	}
	return string(s)
}
