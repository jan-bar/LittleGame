package main

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"image/png"
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
		audioCtx = audio.NewContext(4800)
	)

	zr, err := zip.NewReader(data, data.Size())
	if err != nil {
		return err
	}

	//goland:noinspection SpellCheckingInspection 文件名和资源对应关系
	resMap := map[string]int{
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

		fr, err := f.Open()
		if err != nil {
			return err
		}

		switch filepath.Ext(f.Name) {
		case ".png":
			img, e := png.Decode(fr)
			if e == nil {
				g.images[i] = ebiten.NewImageFromImage(img)
			}
			err = e // 避免阴影变量
		case ".wav":
			ado, e := wav.DecodeWithSampleRate(audioCtx.SampleRate(), fr)
			if e == nil {
				g.audios[i], e = audioCtx.NewPlayer(ado)
			}
			err = e // 避免阴影变量
		case ".dat":
			// todo 开局库解析
		}
		if _ = fr.Close(); err != nil {
			return err // fr.Close() 必须执行
		}
	}
	return nil
}
