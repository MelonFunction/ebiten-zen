// Package zen is the root for all ebiten-zen files
package zen

import (
	_ "embed"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed outlineshader.go
var outline_go []byte

// IsometricDrawable is an interface that Wall and Floor must satisfy
type IsometricDrawable interface {
	Draw(camera *Camera)
}

// Billboard represents a sprite that always faces the camera
type Billboard struct {
	Position *Vector

	RotatedPos *Vector

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
	outlineThickness int // stolen from spritestack, added in later as shader
	outlineColor     color.RGBA

	internalImage *ebiten.Image
	spriteSheet   *SpriteSheet
	Sprites       []*ebiten.Image // if len() == 1, same will be used for all walls
}

// Floor represents a floor tile in the world
type Floor struct {
	Position *Vector

	Rotation      float64 // individual rotation
	RotationPoint *Vector
	RotatedPos    *Vector

	Sprite *ebiten.Image
}

// Wall represents a wall tile in the world
type Wall struct {
	Position *Vector
	Height   float64 // height of the top face of the tile

	Rotation      float64 // individual rotation
	RotationPoint *Vector
	RotatedPos    *Vector

	TopSprite   *ebiten.Image
	WallSprites []*ebiten.Image // if len() == 1, same will be used for all walls
}

// NewBillboard returns a *Billboard
func NewBillboard(sprite *ebiten.Image, position *Vector) *Billboard {
	s := &Billboard{
		Position:   position,
		RotatedPos: NewVector(0, 0),
		Sprite:     sprite,
	}

	return s
}

// Draw draws a rotated texture
func (s *Billboard) Draw(camera *Camera) {
	worldRotationPoint := camera.Position
	s.RotatedPos = s.Position.RotateAround(camera.WorldRotation, worldRotationPoint)
	w, d := float64(s.Sprite.Bounds().Dx()), float64(s.Sprite.Bounds().Dy())
	op := &ebiten.DrawImageOptions{}
	op.ColorScale.Scale(1, 1, 1, 1)
	op = camera.GetTranslation(op, s.RotatedPos.X-w/2, s.RotatedPos.Y-d/2)
	camera.Surface.DrawImage(s.Sprite, op)
}

// NewSpriteStack returns a *SpriteStack
func NewSpriteStack(spriteSheet *SpriteSheet, rotation float64, position, rotationPoint *Vector) *SpriteStack {
	size := math.Max(float64(spriteSheet.SpriteWidth), float64(spriteSheet.SpriteHeight))

	adjustedSpriteSheet := spriteSheet
	if spriteSheet.OutlineThickness > 0 {
		adjustedSpriteSheet = NewSpriteSheet(
			spriteSheet.Image,
			spriteSheet.OrigSpriteWidth,
			spriteSheet.OrigSpriteHeight,
			SpriteSheetOptions{
				Scale: spriteSheet.Scale,
			})
	}

	shader, err := ebiten.NewShader(outline_go)
	if err != nil {
		log.Fatal(err)
	}

	s := &SpriteStack{
		spriteSheet:      adjustedSpriteSheet,
		Rotation:         rotation,
		Position:         position,
		RotationPoint:    rotationPoint,
		RotatedPos:       NewVector(0, 0),
		internalImage:    ebiten.NewImage(int(size)*2, int(size)*2),
		outlineShader:    shader,
		outlineThickness: spriteSheet.OutlineThickness,
		outlineColor:     spriteSheet.OutlineColor,
	}

	return s
}

// PreRender draws the stack to an offscreen image, allowing for outline
func (s *SpriteStack) PreRender(image *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Rotate(0)
	op.ColorScale.Scale(1, 1, 1, 1)

	s.internalImage.Clear()
	w, h := s.spriteSheet.SpriteWidth, s.spriteSheet.SpriteHeight
	op.GeoM.Translate(-float64(w)/2, -float64(h)/2)
	op.GeoM.Rotate(s.Rotation)
	op.GeoM.Translate(float64(w)/2, float64(h)/2)
	op.GeoM.Translate(-float64(w)/2, -float64(h)/2)
	op.GeoM.Translate(
		float64(s.internalImage.Bounds().Dx())/2,
		float64(s.internalImage.Bounds().Dy())/2)

	for i := s.spriteSheet.SpritesHigh - 1; i >= 0; i-- {
		sprite := s.spriteSheet.GetSprite(0, i)
		op.GeoM.Translate(0, math.Min(-1, -float64(s.spriteSheet.Scale)+0.5))
		image.DrawImage(sprite, op)
	}
}

// Draw draws a rotated texture
func (s *SpriteStack) Draw(camera *Camera) {
	rotation := camera.WorldRotation + s.Rotation
	rotation = math.Atan2(math.Sin(rotation), math.Cos(rotation))
	worldRotationPoint := camera.Position
	s.RotatedPos = s.Position.RotateAround(camera.WorldRotation, worldRotationPoint)

	op := &ebiten.DrawImageOptions{}
	op.ColorScale.Scale(1, 1, 1, 1)
	op = camera.GetTranslation(op, s.RotatedPos.X-float64(s.internalImage.Bounds().Dx())/2, s.RotatedPos.Y-float64(s.internalImage.Bounds().Dy())/2)
	s.internalImage.Clear()
	s.PreRender(s.internalImage)
	// camera.Surface.DrawImage(s.internalImage, op)

	sp := &ebiten.DrawRectShaderOptions{}
	sp.GeoM = op.GeoM
	sp.Uniforms = map[string]any{
		"OutlineThickness": float32(s.outlineThickness),
		"OutlineColor":     []float32{float32(s.outlineColor.R), float32(s.outlineColor.G), float32(s.outlineColor.B), float32(s.outlineColor.A)},
	}
	sp.Images[0] = s.internalImage
	camera.Surface.DrawRectShader(s.internalImage.Bounds().Dx(), s.internalImage.Bounds().Dy(), s.outlineShader, sp)

}

// NewFloor returns a *Floor
func NewFloor(sprite *ebiten.Image, rotation float64, position, rotationPoint *Vector) *Floor {
	return &Floor{
		Sprite:        sprite,
		Rotation:      rotation,
		Position:      position,
		RotationPoint: rotationPoint,
		RotatedPos:    NewVector(0, 0),
	}
}

// Draw draws a rotated texture
func (f *Floor) Draw(camera *Camera) {
	rotation := camera.WorldRotation + f.Rotation
	rotation = math.Atan2(math.Sin(rotation), math.Cos(rotation))
	worldRotationPoint := camera.Position
	f.RotatedPos = f.Position.RotateAround(camera.WorldRotation, worldRotationPoint)
	w, d := float64(f.Sprite.Bounds().Dx()), float64(f.Sprite.Bounds().Dy())
	op := &ebiten.DrawImageOptions{}
	op.ColorScale.Scale(1, 1, 1, 1)
	op = camera.GetRotation(op, rotation, -w/2, -d/2)
	op = camera.GetTranslation(op, f.RotatedPos.X-w/2, f.RotatedPos.Y-d/2)
	camera.Surface.DrawImage(f.Sprite, op)
}

// NewWall returns a *Wall
func NewWall(topSprite *ebiten.Image, wallSprites []*ebiten.Image, height, rotation float64, position, rotationPoint *Vector) *Wall {
	return &Wall{
		TopSprite:     topSprite,
		WallSprites:   wallSprites,
		Height:        height,
		Rotation:      rotation,
		Position:      position,
		RotationPoint: rotationPoint,
		RotatedPos:    NewVector(0, 0),
	}
}

// Draw draws a textured cube
func (t *Wall) Draw(camera *Camera) {
	rotation := camera.WorldRotation + t.Rotation
	rotation = math.Atan2(math.Sin(rotation), math.Cos(rotation)) // clip rotation value
	worldRotationPoint := camera.Position
	t.RotatedPos = t.Position.RotateAround(camera.WorldRotation, worldRotationPoint)
	op := &ebiten.DrawImageOptions{}
	w, d, h := float64(t.TopSprite.Bounds().Dx()), float64(t.TopSprite.Bounds().Dy()), t.Height

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
		camera.Surface.DrawImage(img, op)
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

	w, d = float64(t.TopSprite.Bounds().Dx()), float64(t.TopSprite.Bounds().Dy())
	op = &ebiten.DrawImageOptions{}
	op.ColorScale.Scale(1, 1, 1, 1)
	op = camera.GetRotation(op, rotation, -w/2, -d/2)
	op = camera.GetTranslation(op, t.RotatedPos.X-w/2, t.RotatedPos.Y-d/2-h)
	camera.Surface.DrawImage(t.TopSprite, op)
}
