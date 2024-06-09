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
	dungeon := zen.NewDungeon(w, h)
	dungeon.Border = 2
	dungeon.MinDoorSize = 1
	dungeon.MaxDoorSize = 2
	dungeon.MaxRoomWidth = 8
	dungeon.MaxRoomHeight = 8
	dungeon.MinRoomWidth = 4
	dungeon.MinRoomHeight = 4
	dungeon.ShowErrorMessages = true

	style := Dungeon
	var err error
	switch style {
	case RandomWalk:
		dungeon.WallThickness = 2
		dungeon.Border = dungeon.WallThickness
		dungeon.MinIslandSize = 26 // 26 is default
		err = dungeon.GenerateRandomWalk((80 * 80) / 4)
		// clean up the lil floaters
		dungeon.CleanIslands()
		dungeon.CleanWalls(5)
		dungeon.CleanWalls(5)
		dungeon.CleanIslands()
		dungeon.CleanWalls(6)
		dungeon.CleanWalls(6)

		dungeon.AddWalls()
	case DungeonGrid:
		dungeon.WallThickness = 2
		dungeon.Border = dungeon.WallThickness
		dungeon.AllowRandomCorridorOffset = false
		err = dungeon.GenerateDungeonGrid(10)
		dungeon.AddWalls()
	case Dungeon:
		dungeon.WallThickness = 1
		dungeon.Border = dungeon.WallThickness
		dungeon.AllowRandomCorridorOffset = true
		err = dungeon.GenerateDungeon(10)
		dungeon.AddWalls()
	}

	if err != nil {
		log.Println(err)
	}

	// Replace the dungeon's tiles with some debug emojis to see door and room placement
	for door := range dungeon.Doors {
		for dx := door.X; dx < door.X+door.W; dx++ {
			for dy := door.Y; dy < door.Y+door.H; dy++ {
				dungeon.Tiles[dy][dx] = zen.DungeonTileDoor
			}
		}
	}
	for room := range dungeon.Rooms {
		dungeon.Tiles[room.Y][room.X] = zen.DungeonTileRoomBegin
		dungeon.Tiles[room.Y+room.H-1][room.X+room.W-1] = zen.DungeonTileRoomEnd
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			fmt.Print(dungeon.Tiles[y][x])
		}
		fmt.Println()
	}
}
