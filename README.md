## Ebiten-Zen
A simple library to make game development with ebitengine zen-like.

### Features:
- Tiles
    - Dungeon generation
    - 2.5d cube/wall + floor sprite rendering
    - Isometric/orthographic projection + world rotation
- Camera
    - Look at
    - Screen and world rotation (for 2.5d)
    - Zoom
    - Easy to use coordinate system
- Spritesheets + Animation
    - Simple spritesheet creation
    - Use a spritesheet to create multiple animations
    - Can be used with other Zen functions for convenience
    - 🚧 Prerendered sprite stacks for 2.5d (can be exported from MagicaVoxel)
- Dungeon Generation
    - 3 styles:
        - Random walk, like the desert from Nuclear Throne
        - DungeonGrid, like the old Lost Halls from RotMG
        - Dungeon, like the typical dungeon from any other rogue-like
    - A few config options, like wall thickness, corridor width, number of rooms + size
- Collision Detection
    - Uses a simple spatially partitioned hash
    - Rects, Circles, Points
    - Efficient enough!
- Vectors
    - Can Add, Subtract, Rotate, RotateAround
    - Used internally by Zen too
- 🚧 Pathfinding
- 🚧 Scenes
    - A simple way to set up and switch game scenes
- 🚧 UI
    - Buttons
    - Inputs
    - Inventory/bank system (drag and drop)