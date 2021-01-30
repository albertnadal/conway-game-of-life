///////////////////////////////////////////////////////////////////////////////
/**
 * @file   main.go
 * @author Albert Nadal Garriga (anadalg@gmail.com)
 * @date   17-01-2021
 * @brief  Go implementation of the Conway Game of Life
 * @usage  go run main.go --fps=60 --threads=16 --file=queenbeeturner.rle
 */
///////////////////////////////////////////////////////////////////////////////

package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/kbinani/screenshot"
)

type GameOfLife struct {
	ScreenWidth                 int32
	ScreenHeight                int32
	WorldWidth                  int32
	WorldHeight                 int32
	Cells                       [][]bool
	CurrentGenerationCellsIndex int
	Canvas                      rl.RenderTexture2D
	ThreadWaitGroup             sync.WaitGroup
	FragmentWidth               int32
	FragmentHeight              int32
	MaxThreads                  int32
}

var filename = flag.String("file", "", "File with a Game Of Life map in Extended RLE format.")
var fps = flag.String("fps", "", "Frames per second.")
var maxThreads = flag.String("threads", "16", "Max threads used for processing (16 by default).")

func main() {

	flag.Parse()
	fmt.Println("Game of Life")

	n := screenshot.NumActiveDisplays()
	if n < 1 {
		fmt.Println("No screens found.")
		os.Exit(1)
	}

	// Get boundaries of the first available screen
	bounds := screenshot.GetDisplayBounds(0)
	screenWidth := int32(bounds.Dx())
	screenHeight := int32(bounds.Dy())

	// Set-up the Go runtime to use all the available CPU cores
	totalCores := runtime.NumCPU()
	fmt.Printf("- Multi-threaded cores available: %d\n", totalCores)
	runtime.GOMAXPROCS(totalCores)

	rl.InitWindow(screenWidth, screenHeight, "Game of Life")
	rl.SetWindowPosition(0, 30)

	// Limit the fps to adjust renderization speed
	if len(*fps) > 0 {
		fps_, _ := strconv.Atoi(*fps)
		rl.SetTargetFPS(int32(fps_))
	}

	// Set-up the total amount of threads (goroutines) used for processing
	var maxThreads_ int = 16
	if len(*maxThreads) > 0 {
		maxThreads_, _ = strconv.Atoi(*maxThreads)
	}

	gameOfLife := GameOfLife{ScreenWidth: screenWidth, ScreenHeight: screenHeight, MaxThreads: int32(maxThreads_)}
	if len(*filename) == 0 {
		// Run the Game of Life with a random generated pattern
		gameOfLife.Init()
	} else {
		// Run the Game of Life using a pattern from file in Extended RLE format
		gameOfLife.InitWithFile(*filename)
	}

	for !rl.WindowShouldClose() {
		gameOfLife.Draw()
		gameOfLife.Update()
	}

	rl.UnloadTexture(gameOfLife.Canvas.Texture)
	rl.CloseWindow()
}

// GameOfLife functions

func (m *GameOfLife) Init() {
	// By making the world 3 times smaller then the screen resolution we get a more pixelated texture, so cells are 3 times bigger than the size of a single pixel.
	m.WorldWidth = getMultiple(m.ScreenWidth/3, m.MaxThreads)
	m.WorldHeight = m.ScreenHeight / 3
	m.FragmentWidth = int32(math.Ceil(float64(m.WorldWidth-1) / float64(m.MaxThreads)))
	m.FragmentHeight = m.WorldHeight - 1
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

func (m *GameOfLife) InitWithFile(filename string) {
	f, _ := os.Open(filename)
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanBytes)

	buffer := ""
	var count int32 = 0
	width := 0
	height := 0
	var row int32 = 0

	// Initialize a Game of Life pattern in Extended RLE from the file
	// Parse the first line to get the width and height of the pattern
	for scanner.Scan() {
		b := scanner.Bytes()
		if (b[0] == 13) || (b[0] == 10) {
			// New Line (10) or Carriage Return (13)
			s := strings.Split(buffer, ",")
			width, _ = strconv.Atoi(strings.TrimSpace(strings.Split(s[0], "=")[1]))
			height, _ = strconv.Atoi(strings.TrimSpace(strings.Split(s[1], "=")[1]))
			break
		} else {
			buffer += string(b)
		}
	}

	buffer = ""
	m.WorldWidth = getMultiple(int32(width), m.MaxThreads)
	m.WorldHeight = int32(height)
	m.FragmentWidth = int32(math.Ceil(float64(m.WorldWidth-1) / float64(m.MaxThreads)))
	m.FragmentHeight = m.WorldHeight - 1
	m.Canvas = rl.LoadRenderTexture(m.WorldWidth, m.WorldHeight)
	m.CurrentGenerationCellsIndex = 0

	// Initialize cell vectors
	totalCells := m.WorldWidth * m.WorldHeight
	m.Cells = make([][]bool, 2)
	m.Cells[0] = make([]bool, totalCells)
	m.Cells[1] = make([]bool, totalCells)

	// Parse the encoded content in Extended RLE format
	for scanner.Scan() {
		b := scanner.Bytes()
		if (b[0] >= 48) && (b[0] <= 57) {
			// Numeric character
			buffer += string(b)
		} else if (string(b) == "b") || (string(b) == "o") {
			e := 1
			if buffer != "" {
				e, _ = strconv.Atoi(buffer)
			}

			for i := count; i < count+int32(e); i++ {
				if string(b) == "b" {
					m.Cells[0][i] = false
				} else {
					m.Cells[0][i] = true
				}
			}
			buffer = ""
			count += int32(e)
		} else if (string(b) == "$") || (string(b) == "!") {
			for i := count; i < (row+1)*m.WorldWidth; i++ {
				m.Cells[0][i] = false
			}
			row++
			count = row * int32(m.WorldWidth)
		}
	}

	f.Close()
}

func (m *GameOfLife) Update() {
	rl.SetWindowTitle(fmt.Sprintf("Game of Life. FPS: %f\n", rl.GetFPS()))
	for i := int32(0); i < m.MaxThreads; i++ {
		m.ThreadWaitGroup.Add(1)
		go m.UpdateFragment(i)
	}
	m.ThreadWaitGroup.Wait()
	m.CurrentGenerationCellsIndex = (m.CurrentGenerationCellsIndex + 1) % 2
}

func (m *GameOfLife) UpdateFragment(thread_index int32) {
	defer m.ThreadWaitGroup.Done()
	nextGenerationCells := m.Cells[(m.CurrentGenerationCellsIndex+1)%2]
	for x := thread_index * m.FragmentWidth; x < (thread_index*m.FragmentWidth)+m.FragmentWidth; x++ {
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
}

func (m *GameOfLife) GetLiveNeighboursCount(x, y int32) int {
	var count int = 0

	type coord struct {
		x, y int32
	}

	neighbourCoordinates := []coord{
		{x: -1, y: -1},
		{x: 0, y: -1},
		{x: 1, y: -1},
		{x: -1, y: 0},
		{x: 1, y: 0},
		{x: -1, y: 1},
		{x: 0, y: 1},
		{x: 1, y: 1},
	}

	for _, c := range neighbourCoordinates {
		_x := mod(x+c.x, m.WorldWidth-1)
		_y := mod(y+c.y, m.WorldHeight-1)
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
			if m.Cells[m.CurrentGenerationCellsIndex][(m.WorldWidth*(m.WorldHeight-y-1))+x] {
				rl.DrawPixel(x, y, rl.NewColor(0, 0, 0, 255))
			}
		}
	}
	rl.EndTextureMode()
	rl.DrawTexturePro(m.Canvas.Texture, rl.NewRectangle(0, 0, float32(m.Canvas.Texture.Width), float32(m.Canvas.Texture.Height)), rl.NewRectangle(0, 0, float32(m.ScreenWidth), float32(m.ScreenHeight)), rl.NewVector2(float32(0), float32(0)), 0, rl.RayWhite)
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

func getMultiple(a, b int32) int32 {
	if (a % b) != 0 {
		return a + (b - (a % b))
	}
	return a
}
