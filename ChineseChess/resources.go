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
	// king,advisor,bishop,knight,rook,cannon,pawn
	// 红: 帅-K,仕-A,相-B,马-N,车-R,炮-C,兵-P
	// 黑: 将-k,仕-a,相-b,马-n,车-r,炮-c,兵-p
	// 数字代表空位数量,"w"代表红方走,"b"代表黑方走,两个"-"在中国象棋中没有意义
	// 表示双方没有吃子的走棋步数(半回合数)，通常该值达到120就要判和(六十回合自然限着),一旦形成局面的上一步是吃子,这里就标记"0"
	// 最后一个数字表示回合数, 示例请看: boardStart

	var i, j int // fen数据必须合法,否则可能出panic(下标越界)
	for is := 0; is < len(fen); is++ {
		switch fen[is] {
		case ' ':
			g.redPlayer = fen[is+1] == 'w'
			return // 赋值哪方走子,并结束解析
		case '/':
			i++
			j = 0 // 换行
		case '1', '2', '3', '4', '5', '6', '7', '8', '9':
			for t := fen[is] - '0'; t > 0; t-- {
				g.board[i][j] = 0
				j++ // 填充空格
			}
		case 'K':
			g.board[i][j] = imgRedShuai
			j++
		case 'k':
			g.board[i][j] = imgBlackJiang
			j++
		case 'A':
			g.board[i][j] = imgRedShi
			j++
		case 'a':
			g.board[i][j] = imgBlackShi
			j++
		case 'B':
			g.board[i][j] = imgRedXiang
			j++
		case 'b':
			g.board[i][j] = imgBlackXiang
			j++
		case 'N':
			g.board[i][j] = imgRedMa
			j++
		case 'n':
			g.board[i][j] = imgBlackMa
			j++
		case 'R':
			g.board[i][j] = imgRedJu
			j++
		case 'r':
			g.board[i][j] = imgBlackJu
			j++
		case 'C':
			g.board[i][j] = imgRedPao
			j++
		case 'c':
			g.board[i][j] = imgBlackPao
			j++
		case 'P':
			g.board[i][j] = imgRedBing
			j++
		case 'p':
			g.board[i][j] = imgBlackBing
			j++
		}
	}
}
