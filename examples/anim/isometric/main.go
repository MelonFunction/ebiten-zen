// Package main
package main

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"image/color"
	"image/png"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	zen "github.com/melonfunction/ebiten-zen"
)

//go:embed car.png
//go:embed tiles.png
var embedded embed.FS

// vars
var (
	camera *zen.Camera

	floor       *zen.Floor
	wall        *zen.Wall
	spriteStack *zen.SpriteStack
	billboard   *zen.Billboard

	WindowWidth      = 640 * 2
	WindowHeight     = 480 * 2
	SpriteSheetTiles *zen.SpriteSheet
	SpriteSheetCar   *zen.SpriteSheet

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
		spriteStack.Rotation -= 0.025
	}
	if ebiten.IsKeyPressed(ebiten.KeyR) {
		camera.RotateWorld(0.025)
		spriteStack.Rotation += 0.025
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
	spriteStack.Draw(camera)
	billboard.Draw(camera)

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
	log.SetFlags(log.Lshortfile)

	if b, err := embedded.ReadFile("tiles.png"); err == nil {
		if s, err := png.Decode(bytes.NewReader(b)); err == nil {
			sprites := ebiten.NewImageFromImage(s)
			SpriteSheetTiles = zen.NewSpriteSheet(sprites, 16, 16, zen.SpriteSheetOptions{
				Scale: 2,
				// OutlineThickness: 2, // Spritestack creates a new spritesheet and reimplements this value
				// OutlineColor:     color.RGBA{255, 0, 0, 255},
			})
		}
	} else {
		log.Fatal(err)
	}

	if b, err := embedded.ReadFile("car.png"); err == nil {
		if s, err := png.Decode(bytes.NewReader(b)); err == nil {
			sprites := ebiten.NewImageFromImage(s)
			SpriteSheetCar = zen.NewSpriteSheet(sprites, 15, 32, zen.SpriteSheetOptions{
				Scale: 2,
				// OutlineThickness: 2, // Spritestack creates a new spritesheet and reimplements this value
				// OutlineColor:     color.RGBA{0, 255, 0, 255},
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

	// outlines on IsometricDrawables are dynamic!
	// you can't use a spritesheet with an outline as it
	// breaks how the stack is rendered! (but this is )
	floor = zen.NewFloor(
		SpriteSheetTiles.GetSprite(4, 0),
		0,
		zen.NewVector(-float64(SpriteSheetTiles.SpriteWidth)*4, 0),
		zen.NewVector(float64(SpriteSheetTiles.SpriteWidth)/2, float64(SpriteSheetTiles.SpriteWidth)/2))
	floor.OutlineColor = color.RGBA{0, 255, 0, 255}
	floor.OutlineThickness = 2

	wall = zen.NewWall(
		SpriteSheetTiles.GetSprite(0, 0),
		[]*ebiten.Image{
			SpriteSheetTiles.GetSprite(2, 0),
			SpriteSheetTiles.GetSprite(1, 0),
			SpriteSheetTiles.GetSprite(1, 0),
			SpriteSheetTiles.GetSprite(1, 0),
		},
		float64(SpriteSheetTiles.SpriteWidth),
		0,
		zen.NewVector(-float64(SpriteSheetTiles.SpriteWidth)*3, 0),
		zen.NewVector(float64(SpriteSheetTiles.SpriteWidth)/2, float64(SpriteSheetTiles.SpriteWidth)/2),
	)
	wall.OutlineColor = color.RGBA{255, 0, 0, 255}
	wall.OutlineThickness = 2

	spriteStack = zen.NewSpriteStack(
		SpriteSheetCar,
		math.Pi/4,
		zen.NewVector(float64(SpriteSheetTiles.SpriteWidth)*5, 0),
		zen.NewVector(float64(SpriteSheetCar.SpriteWidth)/2, float64(SpriteSheetCar.SpriteWidth)/2))
	spriteStack.OutlineColor = color.RGBA{0, 255, 0, 255}
	spriteStack.OutlineThickness = 2

	billboard = zen.NewBillboard(
		SpriteSheetTiles.GetSprite(5, 0),
		zen.NewVector(float64(SpriteSheetTiles.SpriteWidth)*2, 0))
	billboard.OutlineColor = color.RGBA{255, 0, 255, 255}
	billboard.OutlineThickness = 2

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
