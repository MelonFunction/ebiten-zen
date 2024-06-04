// Package main 👍
package main

import (
	"bytes"
	"embed"
	"errors"
	"image/color"
	"image/png"
	"log"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	zen "github.com/melonfunction/ebiten-zen"
)

//go:embed tiles.png
var embedded embed.FS

// vars
var (
	WindowWidth  = 640 * 2
	WindowHeight = 480 * 2

	SpriteSheet *zen.SpriteSheet

	alreadyDrew bool
	surface     *ebiten.Image

	ErrNormalExit = errors.New("Normal exit")
)

// Game implements ebiten.Game interface.
type Game struct{}

// Update proceeds the game state.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ErrNormalExit
	}

	return nil
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	w, h := ebiten.WindowSize()

	// draw some tiles to the surface
	if alreadyDrew {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(4, 4)
		op.GeoM.Translate(-float64(w), -float64(h))
		op.GeoM.Rotate(math.Pi / 4)
		screen.DrawImage(surface, op)

		fill := ebiten.NewImageFromImage(SpriteSheet.PaddedImage)
		fill.Fill(color.RGBA{0, 0, 0, 255})
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Scale(8, 8)
		screen.DrawImage(fill, op)

		op = &ebiten.DrawImageOptions{}
		op.GeoM.Scale(8, 8)
		screen.DrawImage(SpriteSheet.PaddedImage, op)
		return
	}
	alreadyDrew = true

	for x := 0; x < w/int(float64(SpriteSheet.SpriteWidth)); x++ {
		for y := 0; y < h/int(float64(SpriteSheet.SpriteHeight)); y++ {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(
				float64(SpriteSheet.SpriteWidth)*float64(x)*1.2, // add a lil space between tiles
				float64(SpriteSheet.SpriteHeight)*float64(y)*1.2)
			surface.DrawImage(SpriteSheet.GetSprite(1, 1+rand.Int()%3), op)
		}
	}
}

// Layout sets window size
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	WindowWidth = outsideWidth
	WindowHeight = outsideHeight
	return outsideWidth, outsideHeight
}

func main() {
	game := &Game{}
	ebiten.SetWindowSize(WindowWidth, WindowHeight)
	ebiten.SetWindowTitle("Tiles example")
	ebiten.SetWindowResizable(true)

	surface = ebiten.NewImage(ebiten.WindowSize())

	if b, err := embedded.ReadFile("tiles.png"); err == nil {
		if s, err := png.Decode(bytes.NewReader(b)); err == nil {
			sprites := ebiten.NewImageFromImage(s)
			SpriteSheet = zen.NewSpriteSheet(sprites, 8, 8, zen.SpriteSheetOptions{
				Scale: 2,
			})
		}
	} else {
		log.Fatal(err)
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
