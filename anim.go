// Package zen is the root for all ebiten-zen files
package zen

import (
	"image"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// SpriteSheet stores the image and information about the sizing of the SpriteSheet
type SpriteSheet struct {
	Image       *ebiten.Image // original image which was passed on creation
	PaddedImage *ebiten.Image
	Sprites     []*ebiten.Image

	SpriteWidth  int // how big each sprite is
	SpriteHeight int
	SpritesWide  int // how many sprites are in the sheet
	SpritesHigh  int

	Scale            int
	OutlineThickness int
	OutlineColor     color.RGBA
}

// SpriteSheetOptions are the options which are passed to the NewSpriteSheet function
type SpriteSheetOptions struct {
	Scale            int
	OutlineThickness int
	OutlineColor     color.RGBA
}

// NewSpriteSheet returns a new SpriteSheet
func NewSpriteSheet(img *ebiten.Image, origSpriteWidth, origSpriteHeight int, options SpriteSheetOptions) *SpriteSheet {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()

	defaultOptions := SpriteSheetOptions{
		Scale:            1,
		OutlineThickness: 0,
		OutlineColor:     color.RGBA{255, 255, 255, 255},
	}

	if options.Scale == 0 {
		options.Scale = defaultOptions.Scale
	}
	if options.OutlineThickness == 0 {
		options.OutlineThickness = defaultOptions.OutlineThickness
		options.OutlineColor = defaultOptions.OutlineColor
	}

	s := &SpriteSheet{
		Image:        img,
		SpriteWidth:  origSpriteWidth,
		SpriteHeight: origSpriteHeight,
		SpritesWide:  w / origSpriteWidth,
		SpritesHigh:  h / origSpriteHeight,
		Scale:        options.Scale,
	}

	// all white copy of image without any opacity which could ruin outline
	imgWhite := ebiten.NewImage(img.Bounds().Dx(), img.Bounds().Dy())
	op := &ebiten.DrawImageOptions{}
	op.ColorScale.Scale(0, 0, 0, 0xff)
	op.ColorM.Translate(0xff, 0xff, 0xff, 0)
	imgWhite.DrawImage(img, op)

	p := 2 + options.OutlineThickness*2
	paddedImg := ebiten.NewImage(
		(w+(s.SpritesWide+1)*p)*options.Scale,
		(h+(s.SpritesHigh+1)*p)*options.Scale)
	outlineImg := ebiten.NewImage(
		(w+(s.SpritesWide+1)*p)*options.Scale,
		(h+(s.SpritesHigh+1)*p)*options.Scale)
	eraser := ebiten.NewImage(
		origSpriteWidth+options.OutlineThickness*2,
		origSpriteHeight+options.OutlineThickness*2)
	eraser.Fill(color.RGBA{255, 255, 255, 255})

	c := options.OutlineColor
	s.Sprites = make([]*ebiten.Image, s.SpritesWide*s.SpritesHigh)
	for x := 0; x < s.SpritesWide; x++ {
		for y := 0; y < s.SpritesHigh; y++ {
			dx := float64(origSpriteWidth)*float64(x) + float64(p)*(float64(x)+1)
			dy := float64(origSpriteHeight)*float64(y) + float64(p)*(float64(y)+1)

			// draw padding first
			d := func(op *ebiten.DrawImageOptions) {
				paddedImg.DrawImage(img.SubImage(
					image.Rect(
						x*s.SpriteWidth,
						y*s.SpriteHeight,
						(x+1)*s.SpriteWidth,
						(y+1)*s.SpriteHeight,
					)).(*ebiten.Image), op)
			}
			for zx := -p / 2; zx <= p/2; zx++ {
				if zx != 0 {
					op := &ebiten.DrawImageOptions{}
					op.GeoM.Translate(dx+float64(zx), dy)
					op.GeoM.Scale(float64(options.Scale), float64(options.Scale))
					if options.OutlineThickness > 0 {
						op.ColorM.Scale(0, 0, 0, float64(c.A)/0xff)
						op.ColorM.Translate(float64(c.R)/0xff, float64(c.G)/0xff, float64(c.B)/0xff, 0)
					}
					d(op)
				}
			}
			for zy := -p / 2; zy <= p/2; zy++ {
				if zy != 0 {
					op := &ebiten.DrawImageOptions{}
					op.GeoM.Translate(dx, dy+float64(zy))
					op.GeoM.Scale(float64(options.Scale), float64(options.Scale))
					if options.OutlineThickness > 0 {
						op.ColorM.Scale(0, 0, 0, float64(c.A)/0xff)
						op.ColorM.Translate(float64(c.R)/0xff, float64(c.G)/0xff, float64(c.B)/0xff, 0)
					}
					d(op)
				}
			}

			// clear area, if a tile isn't full width, it'll be the wrong size (2px will be increased to 4px wide!)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(dx-float64(options.OutlineThickness), dy-float64(options.OutlineThickness))
			op.GeoM.Scale(float64(options.Scale), float64(options.Scale))
			op.CompositeMode = ebiten.CompositeModeClear
			paddedImg.DrawImage(eraser, op)

			// draw outline to the outlineImg
			for zy := -options.OutlineThickness; zy <= options.OutlineThickness; zy++ {
				for zx := -options.OutlineThickness; zx <= options.OutlineThickness; zx++ {
					op := &ebiten.DrawImageOptions{}
					op.GeoM.Translate(
						dx+float64(zx)/float64(options.Scale),
						dy+float64(zy)/float64(options.Scale))
					op.GeoM.Scale(float64(options.Scale), float64(options.Scale))

					outlineImg.DrawImage(imgWhite.SubImage(
						image.Rect(
							x*s.SpriteWidth,
							y*s.SpriteHeight,
							(x+1)*s.SpriteWidth,
							(y+1)*s.SpriteHeight,
						)).(*ebiten.Image), op)
				}
			}

			// cut out sprite from the outline
			op = &ebiten.DrawImageOptions{}
			op.GeoM.Translate(
				dx, dy)
			op.GeoM.Scale(float64(options.Scale), float64(options.Scale))
			op.ColorM.Scale(0, 0, 0, 100)
			op.ColorM.Translate(1000/0xff, 1000/0xff, 1000/0xff, 0)
			op.CompositeMode = ebiten.CompositeModeDestinationOut
			outlineImg.DrawImage(img.SubImage(
				image.Rect(
					x*s.SpriteWidth,
					y*s.SpriteHeight,
					(x+1)*s.SpriteWidth,
					(y+1)*s.SpriteHeight,
				)).(*ebiten.Image), op)

			// draw the sprite itself
			op = &ebiten.DrawImageOptions{}
			op.GeoM.Translate(
				dx, dy)
			op.GeoM.Scale(float64(options.Scale), float64(options.Scale))
			paddedImg.DrawImage(img.SubImage(
				image.Rect(
					x*s.SpriteWidth,
					y*s.SpriteHeight,
					(x+1)*s.SpriteWidth,
					(y+1)*s.SpriteHeight,
				)).(*ebiten.Image), op)

			// save subimage/reference
			ot := float64(options.OutlineThickness)
			s.Sprites[x+y*s.SpritesWide] = paddedImg.SubImage(
				image.Rect(
					int(dx-ot)*options.Scale,
					int(dy-ot)*options.Scale,
					(int(dx)+s.SpriteWidth+int(ot))*options.Scale,
					(int(dy)+s.SpriteHeight+int(ot))*options.Scale,
				)).(*ebiten.Image)
		}
	}

	// draw outlines with the correct color
	op = &ebiten.DrawImageOptions{}
	op.ColorM.Scale(0, 0, 0, float64(c.A)/0xff)
	op.ColorM.Translate(float64(c.R)/0xff, float64(c.G)/0xff, float64(c.B)/0xff, 0)
	paddedImg.DrawImage(outlineImg, op)

	s.PaddedImage = paddedImg
	s.SpriteWidth += options.OutlineThickness
	s.SpriteHeight += options.OutlineThickness
	s.SpriteWidth *= options.Scale
	s.SpriteHeight *= options.Scale

	eraser.Dispose()

	return s
}

// GetSprite returns the sprite at the position x,y in the tilesheet
func (s *SpriteSheet) GetSprite(x, y int) *ebiten.Image {
	return s.Sprites[x+y*s.SpritesWide]
}

// Frame stores a single frame of an Animation. It contains an image and how long it should be drawn for
type Frame struct {
	Image    *ebiten.Image
	Duration time.Duration // how long to draw this frame for
}

// NewFrame returns a new Frame
func NewFrame(image *ebiten.Image, duration time.Duration) Frame {
	return Frame{
		Image:    image,
		Duration: duration,
	}
}

// Animation stores a list of Frames and other data regarding timing
type Animation struct {
	Frames        []Frame
	CurrentFrame  int
	LastFrameTime time.Time
	Paused        bool
}

// NewAnimation returns a new Animation
func NewAnimation(frames []Frame) *Animation {
	return &Animation{
		Frames: frames,
		Paused: false,
	}
}

// Update updates
func (a *Animation) Update() {
	if a.Paused {
		return
	}

	now := time.Now()
	if (now.Sub(a.LastFrameTime)) > a.Frames[a.CurrentFrame].Duration {
		a.LastFrameTime = now
		a.CurrentFrame++
		if a.CurrentFrame >= len(a.Frames) {
			a.CurrentFrame = 0
		}
	}
}

// Draw draws the animation to the surface with the provided DrawImageOptions
func (a *Animation) Draw(surface *ebiten.Image, op *ebiten.DrawImageOptions) {
	surface.DrawImage(a.Frames[a.CurrentFrame].Image, op)
}

// Pause pauses the animation
func (a *Animation) Pause() {
	a.Paused = true
}

// Play resumes the animation
func (a *Animation) Play() {
	a.Paused = false
}
