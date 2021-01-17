///////////////////////////////////////////////////////////////////////////////
/**
 * @file   main.go
 * @author Albert Nadal Garriga (anadalg@gmail.com)
 * @date   17-01-2021
 * @brief  Go implementation of the Conway Game of Life
 */
///////////////////////////////////////////////////////////////////////////////

package main

import (
	"fmt"
	"github.com/gen2brain/raylib-go/raylib"
	"math/rand"
)

const SCREEN_WIDTH int32 = 1920
const SCREEN_HEIGHT int32 = 1080

type GameOfLife struct {
	WorldWidth                  int32
	WorldHeight                 int32
	Cells                       [][]bool
	CurrentGenerationCellsIndex int
	Canvas                      rl.RenderTexture2D
}

func main() {

	fmt.Println("Game of Life\n")
	rl.InitWindow(SCREEN_WIDTH, SCREEN_HEIGHT, "Game of Life")
	rl.SetTargetFPS(60)

	gameOfLife := GameOfLife{}
	gameOfLife.Init()

	for !rl.WindowShouldClose() {
		gameOfLife.Draw()
		gameOfLife.Update()
	}

	rl.UnloadTexture(gameOfLife.Canvas.Texture)
	rl.CloseWindow()
}

// GameOfLife functions

func (m *GameOfLife) Init() {
	m.WorldWidth = SCREEN_WIDTH / 2
	m.WorldHeight = SCREEN_HEIGHT / 2
	m.Canvas = rl.LoadRenderTexture(m.WorldWidth, m.WorldHeight)
	m.CurrentGenerationCellsIndex = 0

	// Initialize cell vectors
	totalCells := m.WorldWidth * m.WorldHeight
	m.Cells = make([][]bool, 2)
	m.Cells[0] = make([]bool, totalCells)
	m.Cells[1] = make([]bool, totalCells)
	for i := int32(0); i < totalCells; i++ {
		if rand.Intn(2) == 0 {
			m.Cells[0][i] = true
		} else {
			m.Cells[0][i] = false
		}
		m.Cells[1][i] = false // No live cell here
	}
}

func (m *GameOfLife) Update() {
	nextGenerationCells := m.Cells[(m.CurrentGenerationCellsIndex+1)%2]

	for x := int32(0); x < m.WorldWidth; x++ {
		for y := int32(0); y < m.WorldHeight; y++ {
			cellIndex := (m.WorldWidth * y) + x
			liveNeighboursCount := m.GetLiveNeighboursCount(x, y)
			cellIsLive := m.Cells[m.CurrentGenerationCellsIndex][cellIndex]

			// Rule 1: Any live cell with two or three live neighbours survives.
			// Rule 2: Any dead cell with three live neighbours becomes a live cell.
			// Rule 3: All other live cells die in the next generation. Similarly, all other dead cells stay dead.
			if ((liveNeighboursCount == 2) || (liveNeighboursCount == 3)) && cellIsLive {
				nextGenerationCells[cellIndex] = true // survives
			} else if (liveNeighboursCount == 3) && !cellIsLive {
				nextGenerationCells[cellIndex] = true // becomes a live cell
			} else {
				nextGenerationCells[cellIndex] = false // dies or stay dead
			}
		}
	}

	m.CurrentGenerationCellsIndex = (m.CurrentGenerationCellsIndex + 1) % 2
}

func (m *GameOfLife) GetLiveNeighboursCount(x, y int32) int {
	var count int = 0

	type coord struct {
		x, y int32
	}

	neighbourCoordinates := []coord{
		coord{x: -1, y: -1},
		coord{x: 0, y: -1},
		coord{x: 1, y: -1},
		coord{x: -1, y: 0},
		coord{x: 1, y: 0},
		coord{x: -1, y: 1},
		coord{x: 0, y: 1},
		coord{x: 1, y: 1},
	}

	for _, c := range neighbourCoordinates {
		_x := mod(x+c.x, m.WorldWidth)
		_y := mod(y+c.y, m.WorldHeight)
		if m.Cells[m.CurrentGenerationCellsIndex][(m.WorldWidth*_y)+_x] {
			count++
		}
	}

	return count
}

func (m *GameOfLife) Draw() {
	rl.BeginDrawing()
	rl.ClearBackground(rl.RayWhite)

	rl.BeginTextureMode(m.Canvas)
	rl.ClearBackground(rl.RayWhite)
	for x := int32(0); x < m.WorldWidth; x++ {
		for y := int32(0); y < m.WorldHeight; y++ {
			if m.Cells[m.CurrentGenerationCellsIndex][(m.WorldWidth*y)+x] {
				rl.DrawPixel(x, y, rl.NewColor(0, 0, 0, 255))
			}
		}
	}
	rl.EndTextureMode()

	rl.DrawTexturePro(m.Canvas.Texture, rl.NewRectangle(0, 0, float32(m.Canvas.Texture.Width), float32(m.Canvas.Texture.Height)), rl.NewRectangle(0, 0, float32(SCREEN_WIDTH), float32(SCREEN_HEIGHT)), rl.NewVector2(float32(0), float32(0)), 0, rl.RayWhite)
	rl.EndDrawing()
}

func mod(a, b int32) int32 {
	m := a % b
	if a < 0 && b < 0 {
		m -= b
	}
	if a < 0 && b > 0 {
		m += b
	}
	return m
}
