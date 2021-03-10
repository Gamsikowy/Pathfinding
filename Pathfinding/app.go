package main

import (
	"fmt"
	"math"

	pq "github.com/kyroy/priority-queue"
	"github.com/veandco/go-sdl2/sdl"
)

type mouseDetails struct {
	leftButton  bool
	rightButton bool
	x           int32
	y           int32
}

func mouseDetailsHandler() mouseDetails {
	xMouse, yMouse, mouseBtn := sdl.GetMouseState()
	leftBtn := mouseBtn & sdl.ButtonLMask()
	rightBtn := mouseBtn & sdl.ButtonRMask()
	var md mouseDetails
	md.leftButton = (leftBtn == 1)   // 1 - leftBtn
	md.rightButton = (rightBtn == 4) // 4 - rightBtn
	md.x = int32(xMouse)
	md.y = int32(yMouse)
	return md
}

const (
	size         int32 = 400 // assume the height is the same size as the width
	fieldsNumber int32 = 40
)

//Data - the structure used when adding an item to the priority queue
type Data struct {
	sq *square
}

type square struct {
	width        int32  // the width of the individual square
	row          int32  // square row index
	col          int32  // square column index
	x            int32  // x coordinate on the window
	y            int32  // y coordinate on the window
	member       string // neuter, start, end, open, closed, obstacle, path
	neighbors    []*square
	rowsQuantity int32 // number of lines in the grid
}

// constructor
func newSquare(width, row, col, rowsQuantity int32) *square {
	s := new(square)
	s.width = width
	s.row = row
	s.col = col
	s.x = row * width
	s.y = col * width
	s.member = "neuter"
	s.rowsQuantity = rowsQuantity
	return s
}

func (s *square) isNeuter() bool {
	return s.member == "neuter"
}

func (s *square) isObstacle() bool {
	return s.member == "obstacle"
}

func (s *square) isEnd() bool {
	return s.member == "end"
}

func (s *square) makeNeuter() {
	s.member = "neuter"
}

func (s *square) makeStart() {
	s.member = "start"
}

func (s *square) makeEnd() {
	s.member = "end"
}

func (s *square) makeOpen() {
	s.member = "open"
}

func (s *square) makeClosed() {
	s.member = "closed"
}

func (s *square) makeObstacle() {
	s.member = "obstacle"
}

func (s *square) makePath() {
	s.member = "path"
}

func (s *square) getPosition() []int32 {
	return []int32{s.row, s.col}
}

// adding four neighbors to the neighbors slices
func (s *square) neighborsManagement(grid [][]*square) {
	if s.row < s.rowsQuantity-1 && !grid[int(s.row)+1][int(s.col)].isObstacle() {
		s.neighbors = append(s.neighbors, grid[int(s.row)+1][int(s.col)]) // add up
	}
	if s.col > 0 && !grid[int(s.row)][int(s.col)-1].isObstacle() {
		s.neighbors = append(s.neighbors, grid[int(s.row)][int(s.col)-1]) // add left
	}
	if s.col < s.rowsQuantity-1 && !grid[int(s.row)][int(s.col)+1].isObstacle() {
		s.neighbors = append(s.neighbors, grid[int(s.row)][int(s.col)+1]) // add right
	}
	if s.row > 0 && !grid[int(s.row)-1][int(s.col)].isObstacle() {
		s.neighbors = append(s.neighbors, grid[int(s.row)-1][int(s.col)]) // add down
	}
}

func (s *square) drawSquare(renderer *sdl.Renderer) {
	switch s.member {
	case "neuter":
		renderer.SetDrawColor(255, 255, 255, 1) // white
	case "start":
		renderer.SetDrawColor(255, 191, 0, 1) // orange
	case "end":
		renderer.SetDrawColor(128, 0, 255, 1) // purple
	case "open":
		renderer.SetDrawColor(0, 255, 128, 1) // green
	case "closed":
		renderer.SetDrawColor(0, 4, 255, 1) // blue
	case "obstacle":
		renderer.SetDrawColor(0, 0, 0, 1) // black
	case "path":
		renderer.SetDrawColor(255, 0, 191, 1) // pink
	default:
		renderer.SetDrawColor(255, 255, 255, 1)
		fmt.Println("The member field of square cannot be recognized")
	}

	rect := sdl.Rect{s.x, s.y, s.width, s.width}
	renderer.FillRect(&rect)
}

// calculate distance from end to square (heuristic function)
// manhattan distance
func h(a, b []int32) float64 {
	x1, y1, x2, y2 := a[0], a[1], b[0], b[1]
	return math.Abs(float64(x1)-float64(x2)) + math.Abs(float64(y1)-float64(y2))
}

/*
// euclidean distance
func h(a, b []int32) float64 {
	x1, y1, x2, y2 := a[0], a[1], b[0], b[1]
	return math.Sqrt(math.Pow(float64(x2-x1), 2) + math.Pow(float64(y2-y1), 2))
}
*/

// checking if an element is in slice
func isInSlices(element *square, set []*square) bool {
	result := false
	for _, e := range set {
		if e == element {
			result = true
		}
	}
	return result
}

// checking if an element is in map
func isInMap(element *square, set map[*square]*square) bool {
	result := false
	for key := range set {
		if key == element {
			result = true
		}
	}
	return result
}

func aStar(window *sdl.Window, renderer *sdl.Renderer, grid [][]*square, startingSquare, endingSquare *square) bool {
	// filling g and f maps with positive infinities
	g := map[*square]float64{} // the distance between current and the starting square
	for _, row := range grid {
		for _, square := range row {
			g[square] = math.Inf(1)
		}
	}
	// setting g cost of starting square to 0
	g[startingSquare] = 0

	f := map[*square]float64{} // sum of g and h
	for _, row := range grid {
		for _, square := range row {
			f[square] = math.Inf(1)
		}
	}
	// the distance between the square and the target
	f[startingSquare] = h(startingSquare.getPosition(), endingSquare.getPosition())

	// adding start square to set of squares from which we choose the next one
	openSet := pq.NewPriorityQueue()
	openSet.Insert(Data{startingSquare}, f[startingSquare])

	localMemory := map[*square]*square{} // memory of the squares we walked through, using when rebuilt the path
	var openSetMirror []*square          // stores the squares that are in the priority queue
	openSetMirror = append(openSetMirror, startingSquare)

	for openSet.Len() > 0 {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				// let the user close the window
				return false
			}
		}

		currentData := openSet.PopLowest().(Data)
		current := currentData.sq
		// removing current square
		for i, square := range openSetMirror {
			if square == current {
				openSetMirror = append(openSetMirror[:i], openSetMirror[i+1:]...)
			}
		}

		if current == endingSquare {
			// creating the shortest final path
			for isInMap(current, localMemory) {
				current = localMemory[current]
				current.makePath()
				draw(window, renderer, grid)
			}
			return true
		}

		gTmp := g[current] + 1 // the distance between us and the starting square will increase by one
		for _, neighbor := range current.neighbors {
			if gTmp < g[neighbor] {
				localMemory[neighbor] = current
				g[neighbor] = gTmp
				f[neighbor] = gTmp + h(neighbor.getPosition(), endingSquare.getPosition())

				// check if neighbor is in openSetMirror
				if !isInSlices(neighbor, openSetMirror) {
					openSet.Insert(Data{neighbor}, f[neighbor])
					openSetMirror = append(openSetMirror, neighbor)
					neighbor.makeOpen()
				}
			}
		}
		draw(window, renderer, grid)

		if current != startingSquare {
			current.makeClosed()
		}
	}
	return false
}

func bfs(window *sdl.Window, renderer *sdl.Renderer, grid [][]*square, startingSquare, endingSquare *square) bool {
	var id int32 = 0 // recognizing which item was added first
	// adding start square to set of squares from which we choose the next one
	openSet := pq.NewPriorityQueue()
	openSet.Insert(Data{startingSquare}, float64(id))

	startingSquare.makeOpen()

	// filling g and with positive infinities
	g := map[*square]float64{} // the distance between current and the starting square
	for _, row := range grid {
		for _, square := range row {
			g[square] = math.Inf(1)
		}
	}
	// setting g cost of starting square to 0
	g[startingSquare] = 0

	localMemory := map[*square]*square{} // memory of the squares we walked through

	for openSet.Len() > 0 {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				// let the user close the window
				return false
			}
		}

		currentData := openSet.PopLowest().(Data)
		current := currentData.sq

		if current == endingSquare {
			// creating the shortest final path
			for isInMap(current, localMemory) {
				current = localMemory[current]
				current.makePath()
				draw(window, renderer, grid)
			}
			return true
		}
		fmt.Println(len(current.neighbors))
		for _, neighbor := range current.neighbors {
			if neighbor.isNeuter() || neighbor.isEnd() {
				id++
				neighbor.makeOpen()
				current.makeClosed()
				g[neighbor] = g[current] + 1 // the distance between us and the starting square will increase by one
				localMemory[neighbor] = current
				openSet.Insert(Data{neighbor}, float64(id))
			}
		}
		draw(window, renderer, grid)
	}
	return false
}

func designGrid() [][]*square {
	var i, j int32
	fieldSize := size / fieldsNumber
	grid := make([][]*square, fieldsNumber)

	for i = 0; i < fieldsNumber; i++ {
		grid[i] = make([]*square, fieldsNumber)
		for j = 0; j < fieldsNumber; j++ {
			square := newSquare(fieldSize, i, j, fieldsNumber)
			grid[i][j] = square
		}
	}
	return grid
}

func drawGrid(renderer *sdl.Renderer) {
	var i int32
	fieldSize := size / fieldsNumber  // truncated as it should be
	renderer.SetDrawColor(0, 0, 0, 1) // black

	for i = 0; i < fieldsNumber; i++ {
		err := renderer.DrawLine(0, i*fieldSize, size, i*fieldSize)
		if err != nil {
			panic(fmt.Errorf("Renderer while drawing rows: %v", err))
		}
		err = renderer.DrawLine(i*fieldSize, 0, i*fieldSize, size)
		if err != nil {
			panic(fmt.Errorf("Renderer while drawing columns: %v", err))
		}
	}
	renderer.Present()
}

// draw a background and fill the squares
func draw(window *sdl.Window, renderer *sdl.Renderer, grid [][]*square) {
	renderer.Clear()
	renderer.SetDrawColor(255, 255, 255, 1) // white
	renderer.FillRect(&sdl.Rect{0, 0, size, size})

	for _, row := range grid {
		for _, square := range row {
			square.drawSquare(renderer)
		}
	}

	drawGrid(renderer)
	renderer.Present()
}

// specify in which row and column the event occurred
func clickEvent(md mouseDetails) (int32, int32) {
	fieldSize := size / fieldsNumber
	row := md.x / fieldSize
	col := md.y / fieldSize
	return row, col
}

func main() {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(fmt.Errorf("SDL initialize: %v", err))
	}

	window, err := sdl.CreateWindow("Pathfinding", sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, size, size, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(fmt.Errorf("Window creation: %v", err))
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(fmt.Errorf("Renderer creation: %v", err))
	}
	defer renderer.Destroy()

	grid := designGrid()

	var md mouseDetails // storing the details of the mouse click event
	var startingSquare *square
	var endingSquare *square
	starting, ending := false, false
	running := true
	algRunning := false // prevents user interaction during the execution of the algorithm (e.g. changing obstacles)

	for running {
		draw(window, renderer, grid)
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			if algRunning {
				continue
			}
			switch e := event.(type) {
			case *sdl.QuitEvent:
				// let the user close the window
				running = false
				break
			case *sdl.KeyboardEvent:
				// fmt.Printf("[%d ms] Keyboard\ttype:%d\tsym:%c\tmodifiers:%d\tstate:%d\trepeat:%d\n",
				// 	t.Timestamp, t.Type, t.Keysym.Sym, t.Keysym.Mod, t.State, t.Repeat)
				// if the user presses a key and the algorithm is not currently executed, run A*
				// e.Keysym.Sym == 'a' recognizes whether the a key has been pressed
				// e.Type == 768 recognizes the keydown (not keyup) event
				if !algRunning && starting && ending && e.Keysym.Sym == 'a' && e.Type == 768 {
					for _, row := range grid {
						for _, square := range row {
							square.neighborsManagement(grid)
						}
					}
					algRunning = true
					_ = aStar(window, renderer, grid, startingSquare, endingSquare)
					algRunning = false
				} else if !algRunning && starting && ending && e.Keysym.Sym == 'b' && e.Type == 768 {
					for _, row := range grid {
						for _, square := range row {
							square.neighborsManagement(grid)
						}
					}
					algRunning = true
					_ = bfs(window, renderer, grid, startingSquare, endingSquare)
					algRunning = false
				} else if !algRunning && e.Keysym.Sym == 'c' && e.Type == 768 {
					// the c key clears the entire window
					startingSquare, endingSquare = nil, nil
					starting, ending = false, false
					grid = designGrid()
				}
			}

			md = mouseDetailsHandler()
			if md.leftButton {
				row, col := clickEvent(md)
				square := grid[row][col]
				if !starting {
					startingSquare = square
					starting = true
					startingSquare.makeStart()
				} else if !ending && square != startingSquare {
					endingSquare = square
					ending = true
					endingSquare.makeEnd()
				} else if starting && ending && square != startingSquare && square != endingSquare {
					square.makeObstacle()
				}
			} else if md.rightButton {
				row, col := clickEvent(md)
				square := grid[row][col]
				square.makeNeuter()
				if square == startingSquare {
					starting = false
				}
				if square == endingSquare {
					ending = false
				}
			}
			renderer.Present()
		}
		renderer.Present()
	}
}
