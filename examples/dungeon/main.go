// Package main
package main

import (
	"fmt"
	"log"

	zen "github.com/melonfunction/ebiten-zen"
)

// Generation styles
const (
	RandomWalk int = iota
	DungeonGrid
	Dungeon
)

func main() {
	log.SetFlags(log.Lshortfile)

	w, h := 80, 80 // default terminal width (hopefully)
	world := zen.NewDungeon(w, h)
	world.Border = 2
	world.MinDoorSize = 1
	world.MaxDoorSize = 2
	world.MaxRoomWidth = 8
	world.MaxRoomHeight = 8
	world.MinRoomWidth = 4
	world.MinRoomHeight = 4
	world.ShowErrorMessages = true

	style := Dungeon
	var err error
	switch style {
	case RandomWalk:
		world.WallThickness = 2
		world.Border = world.WallThickness
		world.MinIslandSize = 26 // 26 is default
		err = world.GenerateRandomWalk((80 * 80) / 4)
		// clean up the lil floaters
		world.CleanIslands()
		world.CleanWalls(5)
		world.CleanWalls(5)
		world.CleanIslands()
		world.CleanWalls(6)
		world.CleanWalls(6)

		world.AddWalls()
	case DungeonGrid:
		world.WallThickness = 2
		world.Border = world.WallThickness
		world.AllowRandomCorridorOffset = false
		err = world.GenerateDungeonGrid(10)
		world.AddWalls()
	case Dungeon:
		world.WallThickness = 1
		world.Border = world.WallThickness
		world.AllowRandomCorridorOffset = true
		err = world.GenerateDungeon(10)
		world.AddWalls()
	}

	if err != nil {
		log.Println(err)
	}

	// Replace the world's tiles with some debug emojis to see door and room placement
	for door := range world.Doors {
		for dx := door.X; dx < door.X+door.W; dx++ {
			for dy := door.Y; dy < door.Y+door.H; dy++ {
				world.Tiles[dy][dx] = zen.DungeonTileDoor
			}
		}
	}
	for room := range world.Rooms {
		world.Tiles[room.Y][room.X] = zen.DungeonTileRoomBegin
		world.Tiles[room.Y+room.H-1][room.X+room.W-1] = zen.DungeonTileRoomEnd
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			fmt.Print(world.Tiles[y][x])
		}
		fmt.Println()
	}
}
