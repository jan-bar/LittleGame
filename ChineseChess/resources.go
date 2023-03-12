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
