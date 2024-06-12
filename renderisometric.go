// Package zen is the root for all ebiten-zen files
package zen

import (
	_ "embed"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed outlineshader.go
var outline_go []byte

var outlineShader *ebiten.Shader

// loadOutlineShader tries to load the outline shader, panics on error
func loadOutlineShader() {
	if outlineShader == nil {
		shader, err := ebiten.NewShader(outline_go)
		if err != nil {
			panic(err)
		}
		outlineShader = shader
	}
}

// IsometricDrawable is an interface that Wall and Floor must satisfy
type IsometricDrawable interface {
	Draw(camera *Camera)
}

// Billboard represents a sprite that always faces the camera
type Billboard struct {
	Position *Vector

	RotatedPos *Vector

	outlineShader    *ebiten.Shader
	OutlineThickness int // stolen from spritestack, added in later as shader
	OutlineColor     color.RGBA
	internalImage    *ebiten.Image

	Sprite *ebiten.Image
}

// SpriteStack represents a stack of sprites (or basically floor tiles)
type SpriteStack struct {
	Position *Vector
	Height   float64 // height of the top face of the tile

	Rotation      float64 // individual rotation
	RotationPoint *Vector
	RotatedPos    *Vector

	outlineShader    *ebiten.Shader
	OutlineThickness int // stolen from spritestack, added in later as shader
	OutlineColor     color.RGBA
	internalImage    *ebiten.Image
	SpriteSheet      *SpriteSheet // used internally, but public just in case
	// Sprites          []*ebiten.Image // if len() == 1, same will be used for all walls
}

// Floor represents a floor tile in the world
type Floor struct {
	Position *Vector

	Rotation      float64 // individual rotation
	RotationPoint *Vector
	RotatedPos    *Vector

	outlineShader    *ebiten.Shader
	OutlineThickness int // stolen from spritestack, added in later as shader
	OutlineColor     color.RGBA
	internalImage    *ebiten.Image

	Sprite *ebiten.Image
}

// Wall represents a wall tile in the world
type Wall struct {
	Position *Vector
	Height   float64 // height of the top face of the tile

	Rotation      float64 // individual rotation
	RotationPoint *Vector
	RotatedPos    *Vector

	outlineShader    *ebiten.Shader
	OutlineThickness int // stolen from spritestack, added in later as shader
	OutlineColor     color.RGBA
	internalImage    *ebiten.Image

	TopSprite   *ebiten.Image
	WallSprites []*ebiten.Image // if len() == 1, same will be used for all walls
}

// NewBillboard returns a *Billboard
func NewBillboard(sprite *ebiten.Image, position *Vector) *Billboard {
	size := math.Max(float64(sprite.Bounds().Dx()), float64(sprite.Bounds().Dy()))

	loadOutlineShader()

	s := &Billboard{
		Sprite:     sprite,
		Position:   position,
		RotatedPos: NewVector(0, 0),

		internalImage:    ebiten.NewImage(int(size)*2, int(size)*2),
		outlineShader:    outlineShader,
		OutlineThickness: 0,
		OutlineColor:     color.RGBA{0, 0, 0, 0},
	}

	return s
}

// Draw draws a rotated texture
func (s *Billboard) Draw(camera *Camera) {
	worldRotationPoint := camera.Position
	s.RotatedPos = s.Position.RotateAround(camera.WorldRotation, worldRotationPoint)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Rotate(0)
	op.ColorScale.Scale(1, 1, 1, 1)

	s.internalImage.Clear()
	w, h := float64(s.Sprite.Bounds().Dx()), float64(s.Sprite.Bounds().Dy())
	op.GeoM.Translate(-float64(w)/2, -float64(h)/2)
	op.GeoM.Translate(float64(w)/2, float64(h)/2)
	op.GeoM.Translate(-float64(w)/2, -float64(h)/2)
	op.GeoM.Translate(
		float64(s.internalImage.Bounds().Dx())/2,
		float64(s.internalImage.Bounds().Dy())/2)

	s.internalImage.DrawImage(s.Sprite, op)

	op = &ebiten.DrawImageOptions{}
	op = camera.GetTranslation(op, s.RotatedPos.X-float64(s.internalImage.Bounds().Dx())/2, s.RotatedPos.Y-float64(s.internalImage.Bounds().Dy())/2)
	sp := &ebiten.DrawRectShaderOptions{}
	sp.GeoM = op.GeoM
	sp.Uniforms = map[string]any{
		"OutlineThickness": float32(s.OutlineThickness),
		"OutlineColor":     []float32{float32(s.OutlineColor.R), float32(s.OutlineColor.G), float32(s.OutlineColor.B), float32(s.OutlineColor.A)},
	}
	sp.Images[0] = s.internalImage
	camera.Surface.DrawRectShader(s.internalImage.Bounds().Dx(), s.internalImage.Bounds().Dy(), s.outlineShader, sp)
}

// NewSpriteStack returns a *SpriteStack
func NewSpriteStack(spriteSheet *SpriteSheet, rotation float64, position, rotationPoint *Vector) *SpriteStack {
	size := math.Max(float64(spriteSheet.SpriteWidth), float64(spriteSheet.SpriteHeight))

	loadOutlineShader()

	s := &SpriteStack{
		SpriteSheet:   spriteSheet,
		Rotation:      rotation,
		Position:      position,
		RotationPoint: rotationPoint,
		RotatedPos:    NewVector(0, 0),

		internalImage:    ebiten.NewImage(int(size)*2, int(size)*2),
		outlineShader:    outlineShader,
		OutlineThickness: 0,
		OutlineColor:     color.RGBA{0, 0, 0, 0},
	}

	return s
}

// Draw draws a rotated texture
func (s *SpriteStack) Draw(camera *Camera) {
	rotation := camera.WorldRotation + s.Rotation
	rotation = math.Atan2(math.Sin(rotation), math.Cos(rotation))
	worldRotationPoint := camera.Position
	s.RotatedPos = s.Position.RotateAround(camera.WorldRotation, worldRotationPoint)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Rotate(0)
	op.ColorScale.Scale(1, 1, 1, 1)

	s.internalImage.Clear()
	w, h := s.SpriteSheet.SpriteWidth, s.SpriteSheet.SpriteHeight
	op.GeoM.Translate(-float64(w)/2, -float64(h)/2)
	op.GeoM.Rotate(s.Rotation)
	op.GeoM.Translate(float64(w)/2, float64(h)/2)
	op.GeoM.Translate(-float64(w)/2, -float64(h)/2)
	op.GeoM.Translate(
		float64(s.internalImage.Bounds().Dx())/2,
		float64(s.internalImage.Bounds().Dy())/2)

	for i := s.SpriteSheet.SpritesHigh - 1; i >= 0; i-- {
		sprite := s.SpriteSheet.GetSprite(0, i)
		op.GeoM.Translate(0, math.Min(-1, -float64(s.SpriteSheet.Scale)+0.5))
		s.internalImage.DrawImage(sprite, op)
	}

	op = &ebiten.DrawImageOptions{}
	op = camera.GetTranslation(op, s.RotatedPos.X-float64(s.internalImage.Bounds().Dx())/2, s.RotatedPos.Y-float64(s.internalImage.Bounds().Dy())/2)
	sp := &ebiten.DrawRectShaderOptions{}
	sp.GeoM = op.GeoM
	sp.Uniforms = map[string]any{
		"OutlineThickness": float32(s.OutlineThickness),
		"OutlineColor":     []float32{float32(s.OutlineColor.R), float32(s.OutlineColor.G), float32(s.OutlineColor.B), float32(s.OutlineColor.A)},
	}
	sp.Images[0] = s.internalImage
	camera.Surface.DrawRectShader(s.internalImage.Bounds().Dx(), s.internalImage.Bounds().Dy(), s.outlineShader, sp)
}

// NewFloor returns a *Floor
func NewFloor(sprite *ebiten.Image, rotation float64, position, rotationPoint *Vector) *Floor {
	size := math.Max(float64(sprite.Bounds().Dx()), float64(sprite.Bounds().Dy()))

	loadOutlineShader()

	return &Floor{
		Sprite:        sprite,
		Rotation:      rotation,
		Position:      position,
		RotationPoint: rotationPoint,
		RotatedPos:    NewVector(0, 0),

		internalImage:    ebiten.NewImage(int(size)*2, int(size)*2),
		outlineShader:    outlineShader,
		OutlineThickness: 0,
		OutlineColor:     color.RGBA{0, 0, 0, 0},
	}
}

// Draw draws a rotated texture
func (f *Floor) Draw(camera *Camera) {
	rotation := camera.WorldRotation + f.Rotation
	rotation = math.Atan2(math.Sin(rotation), math.Cos(rotation))
	worldRotationPoint := camera.Position
	f.RotatedPos = f.Position.RotateAround(camera.WorldRotation, worldRotationPoint)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Rotate(0)
	op.ColorScale.Scale(1, 1, 1, 1)

	f.internalImage.Clear()
	w, h := float64(f.Sprite.Bounds().Dx()), float64(f.Sprite.Bounds().Dy())
	op.GeoM.Translate(-float64(w)/2, -float64(h)/2)
	op.GeoM.Rotate(rotation)
	op.GeoM.Translate(float64(w)/2, float64(h)/2)
	op.GeoM.Translate(-float64(w)/2, -float64(h)/2)
	op.GeoM.Translate(
		float64(f.internalImage.Bounds().Dx())/2,
		float64(f.internalImage.Bounds().Dy())/2)

	f.internalImage.DrawImage(f.Sprite, op)

	op = &ebiten.DrawImageOptions{}
	op = camera.GetTranslation(op, f.RotatedPos.X-float64(f.internalImage.Bounds().Dx())/2, f.RotatedPos.Y-float64(f.internalImage.Bounds().Dy())/2)
	sp := &ebiten.DrawRectShaderOptions{}
	sp.GeoM = op.GeoM
	sp.Uniforms = map[string]any{
		"OutlineThickness": float32(f.OutlineThickness),
		"OutlineColor":     []float32{float32(f.OutlineColor.R), float32(f.OutlineColor.G), float32(f.OutlineColor.B), float32(f.OutlineColor.A)},
	}
	sp.Images[0] = f.internalImage
	camera.Surface.DrawRectShader(f.internalImage.Bounds().Dx(), f.internalImage.Bounds().Dy(), f.outlineShader, sp)
}

// NewWall returns a *Wall
func NewWall(topSprite *ebiten.Image, wallSprites []*ebiten.Image, height, rotation float64, position, rotationPoint *Vector) *Wall {
	size := math.Max(float64(topSprite.Bounds().Dx()), float64(topSprite.Bounds().Dy()))

	loadOutlineShader()

	w := &Wall{
		TopSprite:     topSprite,
		WallSprites:   wallSprites,
		Height:        height,
		Rotation:      rotation,
		Position:      position,
		RotationPoint: rotationPoint,
		RotatedPos:    NewVector(0, 0),

		internalImage:    ebiten.NewImage(int(size)*2+int(height)*2, int(size)*2+int(height)*2),
		outlineShader:    outlineShader,
		OutlineThickness: 0,
		OutlineColor:     color.RGBA{0, 0, 0, 0},
	}

	return w
}

// Draw draws a textured cube
func (t *Wall) Draw(camera *Camera) {
	rotation := camera.WorldRotation + t.Rotation
	rotation = math.Atan2(math.Sin(rotation), math.Cos(rotation)) // clip rotation value
	worldRotationPoint := camera.Position
	t.RotatedPos = t.Position.RotateAround(camera.WorldRotation, worldRotationPoint)

	op := &ebiten.DrawImageOptions{}
	t.internalImage.Clear()
	// t.internalImage.Fill(color.RGBA{64, 0, 64, 64})

	w, d, h := float64(t.TopSprite.Bounds().Dx()), float64(t.TopSprite.Bounds().Dy()), t.Height
	tx, ty := float64(t.internalImage.Bounds().Dx())/2,
		float64(t.internalImage.Bounds().Dy())/2

	tr := NewVector(t.RotatedPos.X+w/2, t.RotatedPos.Y-d/2).RotateAround(t.Rotation, t.RotationPoint).RotateAround(camera.WorldRotation, t.RotatedPos)
	bl := NewVector(t.RotatedPos.X-w/2, t.RotatedPos.Y+d/2).RotateAround(t.Rotation, t.RotationPoint).RotateAround(camera.WorldRotation, t.RotatedPos)
	br := NewVector(t.RotatedPos.X+w/2, t.RotatedPos.Y+d/2).RotateAround(t.Rotation, t.RotationPoint).RotateAround(camera.WorldRotation, t.RotatedPos)
	tl := NewVector(t.RotatedPos.X-w/2, t.RotatedPos.Y-d/2).RotateAround(t.Rotation, t.RotationPoint).RotateAround(camera.WorldRotation, t.RotatedPos)

	// draw faces clockwise to prevent image flipping
	drawFace := func(p1, p2 *Vector, img *ebiten.Image) {
		op = &ebiten.DrawImageOptions{}
		op.ColorScale.Scale(1, 1, 1, 1)
		op = camera.GetScale(op, (p2.X-p1.X)/(float64(t.TopSprite.Bounds().Dx())), 1)
		op = camera.GetSkew(op, 0, p1.AngleTo(p2))
		op = camera.GetTranslation(op, p1.X, p1.Y-d)
		// camera.Surface.DrawImage(img, op)

		op.GeoM.Translate(
			-float64(camera.Surface.Bounds().Dx())/2-t.RotatedPos.X+camera.Position.X+tx,
			-float64(camera.Surface.Bounds().Dy())/2-t.RotatedPos.Y+camera.Position.Y+ty+d/2)
		t.internalImage.DrawImage(img, op)
	}

	if math.Abs(rotation) <= math.Pi/2 {
		// front
		drawFace(bl, br, t.WallSprites[0])
	}
	if rotation > 0 && rotation <= math.Pi {
		// right
		drawFace(br, tr, t.WallSprites[3])
	}
	if math.Abs(rotation) >= math.Pi/2 {
		// back
		drawFace(tr, tl, t.WallSprites[2])
	}
	if rotation < 0 {
		// left
		drawFace(tl, bl, t.WallSprites[1])
	}

	op = &ebiten.DrawImageOptions{}
	op.ColorScale.Scale(1, 1, 1, 1)
	op = camera.GetRotation(op, rotation, -w/2, -d/2)
	op.GeoM.Translate(d/2+tx/2, ty/2)
	t.internalImage.DrawImage(t.TopSprite, op)

	op = &ebiten.DrawImageOptions{}
	op = camera.GetTranslation(op, t.RotatedPos.X-w/2, t.RotatedPos.Y-d/2-h)
	op.GeoM.Translate(-d/2-tx/2, -ty/2)

	sp := &ebiten.DrawRectShaderOptions{}
	sp.GeoM = op.GeoM
	sp.Uniforms = map[string]any{
		"OutlineThickness": float32(t.OutlineThickness),
		"OutlineColor":     []float32{float32(t.OutlineColor.R), float32(t.OutlineColor.G), float32(t.OutlineColor.B), float32(t.OutlineColor.A)},
	}
	sp.Images[0] = t.internalImage
	camera.Surface.DrawRectShader(t.internalImage.Bounds().Dx(), t.internalImage.Bounds().Dy(), t.outlineShader, sp)
	// camera.Surface.DrawImage(t.internalImage, op)
}
