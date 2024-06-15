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
	Position *Vector2

	RotationPoint *Vector2
	RotatedPos    *Vector2

	outlineShader    *ebiten.Shader
	OutlineThickness int
	OutlineColor     color.RGBA
	internalImage    *ebiten.Image

	Sprite *ebiten.Image
}

// SpriteStack represents a stack of sprites (or basically floor tiles)
type SpriteStack struct {
	Position *Vector2
	Height   float64 // height of the top face of the tile

	Rotation      float64 // individual rotation
	RotationPoint *Vector2
	RotatedPos    *Vector2

	outlineShader    *ebiten.Shader
	OutlineThickness int
	OutlineColor     color.RGBA
	internalImage    *ebiten.Image
	SpriteSheet      *SpriteSheet // used internally, but public just in case
	// Sprites          []*ebiten.Image // if len() == 1, same will be used for all walls
}

// Floor represents a floor tile in the world
type Floor struct {
	Position *Vector2

	Rotation      float64 // individual rotation
	RotationPoint *Vector2
	RotatedPos    *Vector2

	outlineShader    *ebiten.Shader
	OutlineThickness int
	OutlineColor     color.RGBA
	internalImage    *ebiten.Image

	Sprite *ebiten.Image
}

// Wall represents a wall tile in the world
type Wall struct {
	Position *Vector2
	Height   float64 // height of the top face of the tile

	Rotation      float64 // individual rotation
	RotationPoint *Vector2
	RotatedPos    *Vector2

	outlineShader    *ebiten.Shader
	OutlineThickness int
	OutlineColor     color.RGBA
	internalImage    *ebiten.Image

	TopSprite   *ebiten.Image
	WallSprites []*ebiten.Image
}

// NewBillboard returns a *Billboard
func NewBillboard(sprite *ebiten.Image, position, rotationPoint *Vector2) *Billboard {
	size := math.Max(float64(sprite.Bounds().Dx()), float64(sprite.Bounds().Dy()))

	loadOutlineShader()

	s := &Billboard{
		Sprite:        sprite,
		Position:      position,
		RotationPoint: rotationPoint,
		RotatedPos:    NewVector2(0, 0),

		internalImage:    ebiten.NewImage(int(size)*4, int(size)*4),
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
	op.GeoM.Translate(-s.RotationPoint.X, -s.RotationPoint.Y)

	op.GeoM.Translate(
		float64(s.internalImage.Bounds().Dx())/2,
		float64(s.internalImage.Bounds().Dy())/2)

	if s.OutlineThickness > 0 {
		s.internalImage.Clear()
		s.internalImage.DrawImage(s.Sprite, op)
	}

	op = camera.GetTranslation(op, s.RotatedPos.X-float64(s.internalImage.Bounds().Dx())/2, s.RotatedPos.Y-float64(s.internalImage.Bounds().Dy())/2)
	if !(s.OutlineThickness > 0) {
		camera.Surface.DrawImage(s.Sprite, op)
	} else {
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
}

// NewSpriteStack returns a *SpriteStack
func NewSpriteStack(spriteSheet *SpriteSheet, rotation float64, position, rotationPoint *Vector2) *SpriteStack {
	size := math.Max(float64(spriteSheet.SpriteWidth), float64(spriteSheet.SpriteHeight))

	loadOutlineShader()

	s := &SpriteStack{
		SpriteSheet:   spriteSheet,
		Rotation:      rotation,
		Position:      position,
		RotationPoint: rotationPoint,
		RotatedPos:    NewVector2(0, 0),

		internalImage:    ebiten.NewImage(int(size)*2, int(size)*2), // TODO this should adjust based on rotationPoint
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
	s.RotatedPos = s.Position.Sub(s.RotationPoint).Rotate(s.Rotation).RotateAround(camera.WorldRotation, worldRotationPoint)

	if s.OutlineThickness > 0 {
		s.internalImage.Clear()
		// s.internalImage.Fill(color.RGBA{64, 0, 0, 128})
	}

	op := &ebiten.DrawImageOptions{}
	if s.OutlineThickness > 0 {
		op.GeoM.Translate(-s.RotationPoint.X, -s.RotationPoint.Y)
		op.GeoM.Rotate(rotation)

		op.GeoM.Translate(
			float64(s.internalImage.Bounds().Dx())/2,
			float64(s.internalImage.Bounds().Dy())/2)
	} else {
		op = camera.GetRotation(op, rotation, 0, 0)
		op = camera.GetTranslation(op, s.RotatedPos.X, s.RotatedPos.Y)
	}

	for i := s.SpriteSheet.SpritesHigh - 1; i >= 0; i-- {
		sprite := s.SpriteSheet.GetSprite(0, i)
		op.GeoM.Translate(0, math.Min(-1, -float64(s.SpriteSheet.Scale)+0.5))
		if s.OutlineThickness > 0 {
			s.internalImage.DrawImage(sprite, op)
		} else {
			camera.Surface.DrawImage(sprite, op)
		}
	}

	if s.OutlineThickness > 0 {
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(s.RotationPoint.Rotate(rotation).Unpack())
		s.RotatedPos = s.Position.Sub(s.RotationPoint).Rotate(s.Rotation).RotateAround(camera.WorldRotation, worldRotationPoint)
		op = camera.GetTranslation(op,
			s.RotatedPos.X-float64(s.internalImage.Bounds().Dx())/2,
			s.RotatedPos.Y-float64(s.internalImage.Bounds().Dy())/2)
		sp := &ebiten.DrawRectShaderOptions{}
		sp.GeoM = op.GeoM
		sp.Uniforms = map[string]any{
			"OutlineThickness": float32(s.OutlineThickness),
			"OutlineColor":     []float32{float32(s.OutlineColor.R), float32(s.OutlineColor.G), float32(s.OutlineColor.B), float32(s.OutlineColor.A)},
		}
		sp.Images[0] = s.internalImage
		camera.Surface.DrawRectShader(s.internalImage.Bounds().Dx(), s.internalImage.Bounds().Dy(), s.outlineShader, sp)
	}
}

// NewFloor returns a *Floor
func NewFloor(sprite *ebiten.Image, rotation float64, position, rotationPoint *Vector2) *Floor {
	size := math.Max(float64(sprite.Bounds().Dx()), float64(sprite.Bounds().Dy()))

	loadOutlineShader()

	return &Floor{
		Sprite:        sprite,
		Rotation:      rotation,
		Position:      position,
		RotationPoint: rotationPoint,
		RotatedPos:    NewVector2(0, 0),

		internalImage:    ebiten.NewImage(int(size)*4, int(size)*4),
		outlineShader:    outlineShader,
		OutlineThickness: 0,
		OutlineColor:     color.RGBA{0, 0, 0, 0},
	}
}

// Draw draws a rotated texture
func (s *Floor) Draw(camera *Camera) {
	rotation := camera.WorldRotation + s.Rotation
	rotation = math.Atan2(math.Sin(rotation), math.Cos(rotation))
	worldRotationPoint := camera.Position
	s.RotatedPos = s.Position.RotateAround(camera.WorldRotation, worldRotationPoint)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-s.RotationPoint.X, -s.RotationPoint.Y)
	op.GeoM.Rotate(rotation)
	op.GeoM.Translate(
		float64(s.internalImage.Bounds().Dx())/2,
		float64(s.internalImage.Bounds().Dy())/2)

	if s.OutlineThickness > 0 {
		s.internalImage.Clear()
		// s.internalImage.Fill(color.RGBA{64, 0, 0, 128})
		s.internalImage.DrawImage(s.Sprite, op)
	}

	if !(s.OutlineThickness > 0) {
		op = camera.GetTranslation(op, s.RotatedPos.X-float64(s.internalImage.Bounds().Dx())/2, s.RotatedPos.Y-float64(s.internalImage.Bounds().Dy())/2)
		camera.Surface.DrawImage(s.Sprite, op)
	} else {
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
}

// NewWall returns a *Wall
func NewWall(topSprite *ebiten.Image, wallSprites []*ebiten.Image, height, rotation float64, position, rotationPoint *Vector2) *Wall {
	size := math.Sqrt(float64(topSprite.Bounds().Dx()*topSprite.Bounds().Dx()) + float64(topSprite.Bounds().Dy()*topSprite.Bounds().Dy()))

	loadOutlineShader()

	w := &Wall{
		TopSprite:     topSprite,
		WallSprites:   wallSprites,
		Height:        height,
		Rotation:      rotation,
		Position:      position,
		RotationPoint: rotationPoint,
		RotatedPos:    NewVector2(0, 0),

		internalImage:    ebiten.NewImage(int(size)*4, int(size)*4),
		outlineShader:    outlineShader,
		OutlineThickness: 0,
		OutlineColor:     color.RGBA{0, 0, 0, 0},
	}

	return w
}

// Draw draws a textured cube
func (s *Wall) Draw(camera *Camera) {
	rotation := camera.WorldRotation + s.Rotation
	rotation = math.Atan2(math.Sin(rotation), math.Cos(rotation))
	worldRotationPoint := camera.Position
	s.RotatedPos = s.Position.RotateAround(camera.WorldRotation, worldRotationPoint)

	if s.OutlineThickness > 0 {
		s.internalImage.Clear()
		// s.internalImage.Fill(color.RGBA{64, 0, 0, 128})
	}

	op := &ebiten.DrawImageOptions{}
	w, d := float64(s.TopSprite.Bounds().Dx()), float64(s.TopSprite.Bounds().Dy())
	tr := NewVector2(w, 0).RotateAround(rotation, s.RotationPoint)
	bl := NewVector2(0, d).RotateAround(rotation, s.RotationPoint)
	br := NewVector2(w, d).RotateAround(rotation, s.RotationPoint)
	tl := NewVector2(0, 0).RotateAround(rotation, s.RotationPoint)
	// draw faces clockwise to prevent image flipping
	drawFace := func(p1, p2 *Vector2, img *ebiten.Image) {
		op = &ebiten.DrawImageOptions{}
		op = camera.GetScale(op, (p2.X-p1.X)/(float64(s.TopSprite.Bounds().Dx())), 1)
		op = camera.GetSkew(op, 0, p1.AngleTo(p2))
		op.GeoM.Translate(p1.X, p1.Y-d)
		op.GeoM.Translate(-s.RotationPoint.X, -s.RotationPoint.Y+s.Height)
		op.GeoM.Translate(
			float64(s.internalImage.Bounds().Dx())/2,
			float64(s.internalImage.Bounds().Dy())/2)
		if s.OutlineThickness > 0 {
			s.internalImage.DrawImage(img, op)
		} else {
			op = camera.GetTranslation(op, s.RotatedPos.X-float64(s.internalImage.Bounds().Dx())/2, s.RotatedPos.Y-float64(s.internalImage.Bounds().Dy())/2-s.Height)
			camera.Surface.DrawImage(img, op)
		}
	}
	if math.Abs(rotation) <= math.Pi/2 {
		drawFace(bl, br, s.WallSprites[0]) // front
	}
	if rotation > 0 && rotation <= math.Pi {
		drawFace(br, tr, s.WallSprites[3]) // right
	}
	if math.Abs(rotation) >= math.Pi/2 {
		drawFace(tr, tl, s.WallSprites[2]) // back
	}
	if rotation < 0 {
		drawFace(tl, bl, s.WallSprites[1]) // left
	}

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-s.RotationPoint.X, -s.RotationPoint.Y)
	op.GeoM.Rotate(rotation)
	op.GeoM.Translate(
		float64(s.internalImage.Bounds().Dx())/2,
		float64(s.internalImage.Bounds().Dy())/2)

	if s.OutlineThickness > 0 {
		s.internalImage.DrawImage(s.TopSprite, op)
	}

	if !(s.OutlineThickness > 0) {
		op = camera.GetTranslation(op, s.RotatedPos.X-float64(s.internalImage.Bounds().Dx())/2, s.RotatedPos.Y-float64(s.internalImage.Bounds().Dy())/2-s.Height)
		camera.Surface.DrawImage(s.TopSprite, op)
	} else {
		op = &ebiten.DrawImageOptions{}
		op = camera.GetTranslation(op, s.RotatedPos.X-float64(s.internalImage.Bounds().Dx())/2, s.RotatedPos.Y-float64(s.internalImage.Bounds().Dy())/2-s.Height)
		sp := &ebiten.DrawRectShaderOptions{}
		sp.GeoM = op.GeoM
		sp.Uniforms = map[string]any{
			"OutlineThickness": float32(s.OutlineThickness),
			"OutlineColor":     []float32{float32(s.OutlineColor.R), float32(s.OutlineColor.G), float32(s.OutlineColor.B), float32(s.OutlineColor.A)},
		}
		sp.Images[0] = s.internalImage
		camera.Surface.DrawRectShader(s.internalImage.Bounds().Dx(), s.internalImage.Bounds().Dy(), s.outlineShader, sp)
	}
}
