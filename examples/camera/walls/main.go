// Package main 👍
package main

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	zen "github.com/melonfunction/ebiten-zen"
)

//go:embed tiles.png
var embedded embed.FS

// vars
var (
	camera *zen.Camera
	floor  *zen.Floor
	wall   *zen.Wall

	WindowWidth  = 640 * 2
	WindowHeight = 480 * 2
	SpriteSheet  *zen.SpriteSheet

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
	if ebiten.IsKeyPressed(ebiten.KeyG) {
		camera.RotateWorld(-0.025)
	}
	if ebiten.IsKeyPressed(ebiten.KeyR) {
		camera.RotateWorld(0.025)
	}

	// camera movement isn't adjusted by camera rotation here
	if ebiten.IsKeyPressed(ebiten.KeyH) {
		camera.MovePosition(-10, 0)
	}
	if ebiten.IsKeyPressed(ebiten.KeyN) {
		camera.MovePosition(10, 0)
	}
	if ebiten.IsKeyPressed(ebiten.KeyC) {
		camera.MovePosition(0, -10)
	}
	if ebiten.IsKeyPressed(ebiten.KeyT) {
		camera.MovePosition(0, 10)
	}

	return nil
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	camera.Surface.Clear()
	// floor.Rotation += 0.01
	// wall.Rotation += 0.01
	floor.Draw(camera)
	wall.Draw(camera)

	// camera.Surface.DrawImage(SpriteSheet.GetSprite(3, 0), camera.GetTranslation(&ebiten.DrawImageOptions{}, 0, 0))
	mx, my := ebiten.CursorPosition()
	wx, wy := camera.GetWorldCoords(float64(mx), float64(my))
	ebitenutil.DebugPrintAt(camera.Surface, fmt.Sprintf("%d, %d", int(wx), int(wy)), mx, my-16)
	camera.Blit(screen)
}

// Layout sets window size
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	WindowWidth = outsideWidth
	WindowHeight = outsideHeight
	return outsideWidth, outsideHeight
}

func main() {
	if b, err := embedded.ReadFile("tiles.png"); err == nil {
		if s, err := png.Decode(bytes.NewReader(b)); err == nil {
			sprites := ebiten.NewImageFromImage(s)
			SpriteSheet = zen.NewSpriteSheet(sprites, 16, 16, zen.SpriteSheetOptions{
				Scale: 2,
			})
		}
	} else {
		log.Fatal(err)
	}

	game := &Game{}
	ebiten.SetWindowSize(WindowWidth, WindowHeight)
	ebiten.SetWindowTitle("Wall and floor rendering example")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	camera = zen.NewCamera(WindowWidth, WindowHeight, 0, 0, 0, 1)
	floor = &zen.Floor{
		Sprite:        SpriteSheet.GetSprite(4, 0),
		Rotation:      0,
		Position:      zen.NewVector(-float64(SpriteSheet.SpriteWidth)*2, 0),
		RotationPoint: zen.NewVector(float64(SpriteSheet.SpriteWidth)/2, float64(SpriteSheet.SpriteWidth)/2),
		RotatedPos:    zen.NewVector(0, 0),
	}

	wall = &zen.Wall{
		Height:    float64(SpriteSheet.SpriteWidth),
		TopSprite: SpriteSheet.GetSprite(0, 0),
		WallSprites: []*ebiten.Image{
			SpriteSheet.GetSprite(2, 0),
			SpriteSheet.GetSprite(1, 0),
			SpriteSheet.GetSprite(1, 0),
			SpriteSheet.GetSprite(1, 0),
		},
		Rotation:      0,
		Position:      zen.NewVector(-float64(SpriteSheet.SpriteWidth), 0),
		RotationPoint: zen.NewVector(float64(SpriteSheet.SpriteWidth)/2, float64(SpriteSheet.SpriteWidth)/2),
		RotatedPos:    zen.NewVector(0, 0),
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
