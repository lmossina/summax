package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

var _ = strings.Repeat("f unused imports", 1)

// === Game Logic ===
const (
	nRows = 9
	nCols = 16
)

type anchorPoint struct {
	row    int
	col    int
	active bool
}

type Game struct {
	nrows           int
	ncols           int
	board           [][]int
	cursor          [2]int
	selectArea      [][]int // binary array to track selected elements
	toggleSelection bool
	totalScore      int
	// Only track the first point selected, selectionAnchor.
	// All the others will be calculated from the selectRange:
	//  if selectRange == {-1, 2}, take all elements that are from the current
	//  selectionAnchor up to minus 1 (that is, up one row),
	//  and all elements that are from the current selectionAnchor up to 2 (that is, two columns on the right)
	selectRange [2]int
	anchorPoint anchorPoint // the first point selected
}

func (game *Game) initializeBoard(seed ...int64) {
	var r *rand.Rand
	if len(seed) > 0 && seed[0] > 0 {
		r = rand.New(rand.NewSource(seed[0]))
	} else {
		r = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	for i := 0; i < nRows; i++ {
		for j := 0; j < nCols; j++ {
			game.board[i][j] = r.Intn(9) + 1
		}
	}
	for i := 0; i < nRows; i++ {
		for j := 0; j < nCols; j++ {
			game.selectArea[i][j] = 0
		}
	}
}

func (g *Game) clearSelection() {
	for i := 0; i < g.nrows; i++ {
		for j := 0; j < g.ncols; j++ {
			g.selectArea[i][j] = 0
		}
	}
}

func (g *Game) selectionCoordinates() [][]int {
	if !g.toggleSelection {
		panic("Broken logic: selection is not toggled on")
	}
	if !g.anchorPoint.active {
		panic("Broken logic: anchorPoint not active, this state is not admissible")
	}
	ancRow, ancCol := g.anchorPoint.row, g.anchorPoint.col
	cursorRow, cursorCol := g.cursor[0], g.cursor[1]

	upperLeftRow, upperLeftCol := min(ancRow, cursorRow), min(ancCol, cursorCol)
	lowerRightRow, lowerRightCol := max(ancRow, cursorRow), max(ancCol, cursorCol)

	return [][]int{{upperLeftRow, upperLeftCol}, {lowerRightRow, lowerRightCol}}
}

func (g *Game) updateSelectionArea() {
	if g.toggleSelection {
		if !g.anchorPoint.active {
			panic("anchorPoint not active: this state is not admissible")
		}

		coords := g.selectionCoordinates()

		g.clearSelection()
		for i := coords[0][0]; i <= coords[1][0]; i++ {
			for j := coords[0][1]; j <= coords[1][1]; j++ {
				g.selectArea[i][j] = 1
			}
		}
	}
}

func (g *Game) evaluateSelection() {
	// 1. compute score of selected cells
	score := 0
	for i := 0; i < g.nrows; i++ {
		for j := 0; j < g.ncols; j++ {
			if g.selectArea[i][j] == 1 {
				score += g.board[i][j]
			}
		}
	}
	if score == 10 {
		g.totalScore += score
		// 2. delete selected elements when they sum to 10
		for i := 0; i < g.nrows; i++ {
			for j := 0; j < g.ncols; j++ {
				if g.selectArea[i][j] == 1 {
					g.board[i][j] = 0
				}
			}
		}
	}
}

func (g *Game) moveCursor(direction string) {
	switch direction {
	case "up":
		g.moveCursorUp()
	case "down":
		g.moveCursorDown()
	case "left":
		g.moveCursorLeft()
	case "right":
		g.moveCursorRight()
	}
	g.updateSelectionArea()
}

func (g *Game) moveCursorUp(jumpSize ...int) {
	jump := 1
	if len(jumpSize) == 1 {
		jump = max(1, jumpSize[0])
	}

	if g.cursor[0] >= jump {
		g.cursor[0] -= jump
	} else if g.cursor[0] >= 0 {
		g.cursor[0] = 0
	}
	g.updateSelectionArea()
}

func (g *Game) moveCursorDown(jumpSize ...int) {
	jump := 1
	if len(jumpSize) == 1 {
		jump = max(1, jumpSize[0])
	}
	if g.cursor[0] < g.nrows-jump {
		g.cursor[0] += jump
	} else if g.cursor[0] <= g.nrows-1 {
		g.cursor[0] = g.nrows - 1
	}
	g.updateSelectionArea()
}

func (g *Game) moveCursorRight(jumpSize ...int) {
	jump := 1
	if len(jumpSize) == 1 {
		jump = max(1, jumpSize[0])
	}
	if g.cursor[1] < g.ncols-jump {
		g.cursor[1] += jump
	} else if g.cursor[1] <= g.ncols-1 {
		g.cursor[1] = g.ncols - 1
	}
	g.updateSelectionArea()
}

func (g *Game) moveCursorLeft(jumpSize ...int) {
	jump := 1
	if len(jumpSize) == 1 {
		jump = max(1, jumpSize[0])
	}

	if g.cursor[1] >= jump {
		g.cursor[1] -= jump
	} else if g.cursor[1] >= 0 {
		g.cursor[1] = 0
	}
	g.updateSelectionArea()
}

func initGame() Game {
	game := Game{
		nrows:      nRows,
		ncols:      nCols,
		board:      make([][]int, nRows),
		cursor:     [2]int{0, 0},
		selectArea: make([][]int, nRows),
	}
	for i := range game.board {
		game.board[i] = make([]int, nCols)
	}

	for i := range game.selectArea {
		game.selectArea[i] = make([]int, nCols)
	}
	game.initializeBoard()
	return game
}

func (game *Game) handleSelect() {
	if !game.toggleSelection {
		game.toggleSelection = !game.toggleSelection

		game.anchorPoint = anchorPoint{
			row:    game.cursor[0],
			col:    game.cursor[1],
			active: true,
		}
		game.selectArea[game.cursor[0]][game.cursor[1]] = 1
	} else {
		game.evaluateSelection()
		game.clearSelection()
		game.toggleSelection = !game.toggleSelection
		game.anchorPoint = anchorPoint{}
	}
}

func main() {
	// Force truecolor mode
	os.Setenv("TCELL_TRUECOLOR", "1")

	game := initGame()
	tui := TUI{1, [2]rune{' ', ' '}, false, false, false}
	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("Error creating screen: %v", err)
	}

	if err := screen.Init(); err != nil {
		log.Fatalf("Error initializing screen: %v", err)
	}
	defer screen.Fini()

	for { // Main event loop
		screen.Clear()
		displayBoard(&game, &tui, screen, tui.printLogs, tui.printDebug)
		screen.Show()

		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyCtrlC:
				return // Exit the program
			case tcell.KeyEscape:
				if game.toggleSelection {
					game.clearSelection()
					game.toggleSelection = !game.toggleSelection
				}
			case tcell.KeyRight:
				game.moveCursor("right")
			case tcell.KeyLeft:
				game.moveCursor("left")
			case tcell.KeyDown:
				game.moveCursor("down")
			case tcell.KeyUp:
				game.moveCursor("up")
			case tcell.KeyCtrlV:
				game.handleSelect()
			case tcell.KeyRune:
				switch r := ev.Rune(); {
				// hold jump values (0..9)
				case r >= '0' && r <= '9':
					tui.updateBuffer(int(r - '0'))
				case r == 'h':
					game.moveCursorLeft(tui.jumpBuffer)
					tui.updateLastMove(r)
					tui.updateBuffer(1)
				case r == 'j':
					game.moveCursorDown(tui.jumpBuffer)
					tui.updateLastMove(r)
					tui.updateBuffer(1)
				case r == 'k':
					game.moveCursorUp(tui.jumpBuffer)
					tui.updateLastMove(r)
					tui.updateBuffer(1)
				case r == 'l':
					game.moveCursorRight(tui.jumpBuffer)
					tui.updateLastMove(r)
					tui.updateBuffer(1)
				case r == 'v' || r == ' ':
					game.handleSelect()
				}
			}
		}
	}
}

// === UI ===
type TUI struct {
	jumpBuffer  int
	lastMove    [2]rune
	brailleMode bool
	printLogs   bool
	printDebug  bool
}

func (tui *TUI) updateBuffer(val int) {
	tui.jumpBuffer = val
}

func (tui *TUI) updateLastMove(move rune) {
	// warning: this only works with VIM motions
	// FIXME for directions keys operations
	tui.lastMove = [2]rune{rune('0' + tui.jumpBuffer), move}
}

type RGBColors struct {
	red, blue, purple tcell.Color
}

// TODO: test hard-coded rgb colors for UI:
// maybe w/ option to keep user's theme pref?
var red_rgb = tcell.NewRGBColor(255, 0, 0) // Bright red

// these are styles, not colors: FIXME var names
var (
	blue     = tcell.StyleDefault.Foreground(tcell.ColorBlue)
	yellow   = tcell.StyleDefault.Foreground(tcell.ColorYellow)
	purple   = tcell.StyleDefault.Foreground(tcell.ColorPurple)
	green    = tcell.StyleDefault.Foreground(tcell.ColorGreen)
	brown    = tcell.StyleDefault.Foreground(tcell.ColorSaddleBrown)
	gray     = tcell.StyleDefault.Foreground(tcell.ColorGray)
	darkGray = tcell.StyleDefault.Foreground(tcell.ColorDimGray)
	antique  = tcell.StyleDefault.Foreground(tcell.ColorAntiqueWhite)
	white    = tcell.StyleDefault.Foreground(tcell.ColorWhite)
	red      = tcell.StyleDefault.Foreground(tcell.ColorRed)
)

var customStyle = tcell.StyleDefault
var defaultStyle = customStyle // tcell.StyleDefault
var highlightStyle = tcell.StyleDefault.Foreground(tcell.ColorPurple).Background(tcell.ColorWhite)
var cursorHighlight = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow)
var cursorPosition = tcell.StyleDefault.Foreground(tcell.ColorBlack)

func drawString(screen tcell.Screen, row, col int, text string) {
	x := col
	for _, char := range text {
		screen.SetContent(x, row, char, nil, yellow)
		x += runewidth.RuneWidth(char)
	}
}

func drawBorder(screen tcell.Screen, nHeight, nWidth int, frameAnchor [2]int) {
	horiLine := strings.Repeat("─", nWidth-2)
	row0, col0 := frameAnchor[0], frameAnchor[1]

	// draw top row (w/ corners)
	topper := "╭" + horiLine + "╮"
	x := col0
	for _, char := range topper {
		screen.SetContent(x, row0, char, nil, yellow)
		x += runewidth.RuneWidth(char)
	}

	// draw bottom row (w/ corners)
	bottomer := "╰" + horiLine + "╯"
	x = col0
	for _, ch := range bottomer {
		screen.SetContent(x, row0+nHeight-1, ch, nil, green)
		x += runewidth.RuneWidth(ch)
	}

	for i := row0 + 1; i < nHeight+row0-1; i++ {
		for j := col0; j < nWidth+col0; j++ {
			// draw left border line
			if j == col0 {
				screen.SetContent(j, i, '│', nil, red)
			}

			// draw left border line
			if j == nWidth+col0-1 {
				screen.SetContent(j, i, '│', nil, purple)
			}
		}
	}
}

func displayBoard(game *Game, tui *TUI, screen tcell.Screen, printLogs, printDebug bool) {
	// TODO consider making offsets depending on terminal size, and center board/text accordingly
	leftOffset := 4
	upperOffset := 4 //len(borders)

	titleRow := 1

	title := "               SummaX"
	for j, char := range title {
		screen.SetContent(j+leftOffset, titleRow, char, nil, yellow)
	}

	// TODO extract to func
	drawBoard := func(rowOffset, columnOffset int) {
		for i := 0; i < game.nrows; i++ {
			for j := 0; j < game.ncols; j++ {
				var str_value string
				cellValue := game.board[i][j]
				if cellValue == 0 {
					str_value = " "
				} else {
					str_value = fmt.Sprintf("%d", cellValue)
					if tui.brailleMode {
						var brailleChar rune
						switch cellValue {
						case 1:
							brailleChar = '⠁'
						case 2:
							brailleChar = '⠂'
						case 3:
							brailleChar = '⠃'
						case 4:
							brailleChar = '⠄'
						case 5:
							brailleChar = '⠅'
						case 6:
							brailleChar = '⠆'
						case 7:
							brailleChar = '⠇'
						case 8:
							brailleChar = '⠈'
						case 9:
							brailleChar = '⠉'
						}
						str_value = fmt.Sprintf("%c", brailleChar)
					}
				}

				x, y := j*2+1+leftOffset, i+upperOffset+1
				value := " " + str_value

				if game.toggleSelection {
					selcoords := game.selectionCoordinates()
					leftBoundary := selcoords[0][1]
					if j == leftBoundary {
						x, y = j*2+2+leftOffset, i+upperOffset+1
						value = str_value
					}
				}

				// select style of cell: color number representation according to their values
				k := 0
				for _, char := range value {
					thisStyle := defaultStyle
					if game.selectArea[i][j] == 1 {
						thisStyle = highlightStyle
					} else {
						switch cellValue {
						case 1:
							thisStyle = blue
						case 2:
							thisStyle = purple
						case 3:
							thisStyle = green
						case 4:
							thisStyle = brown
						case 5:
							thisStyle = gray
						case 6:
							thisStyle = antique
						case 7:
							thisStyle = white
						case 8:
							thisStyle = red
						case 9:
							thisStyle = yellow
						}
					}
					screen.SetContent(x+k, y, char, nil, thisStyle)
					k += runewidth.RuneWidth(char)
				}
			}
		}
	}

	height := nRows + 1 + 1
	width := nCols + (nCols + 1) + 1 + 1
	anchor := [2]int{upperOffset, leftOffset}
	drawBorder(screen, height, width, anchor)

	// width and height of border lines
	drawBoard(anchor[0], anchor[1])

	drawCursor := func(rowOffset, columnOffset int) {
		cRow, cCol := game.cursor[0], game.cursor[1]
		value := fmt.Sprintf("%d", game.board[cRow][cCol])
		cellChar := rune(value[0])

		// hide cells with value zero, which happens only when cell
		// has been used to score a ten
		if cellChar == '0' {
			cellChar = ' '
		}
		screen.SetContent(cCol*2+2+columnOffset-1, cRow+rowOffset, cellChar, nil, cursorHighlight)

		// draw top bar with VIM motions grid: jumps available 1..9
		// TODO vimRowPos should be automatically one line above upper border
		vimRowPos := 3
		for i := 0; i < cCol; i++ {
			countLeft := cCol - i
			if countLeft <= 9 && countLeft > 0 {
				screen.SetContent(cCol*2+2+columnOffset-1-countLeft*2, vimRowPos, rune('0'+countLeft), nil, darkGray)
			} else {

				screen.SetContent(cCol*2+2+columnOffset-1-countLeft*2, vimRowPos, '·', nil, darkGray)
			}
		}
		for i := cCol; i < game.ncols; i++ {
			countRight := game.ncols - 1 - i
			if countRight <= 9 && countRight > 0 {
				screen.SetContent(cCol*2+2+columnOffset-1+countRight*2, vimRowPos, rune('0'+countRight), nil, darkGray)
			} else {
				screen.SetContent(cCol*2+2+columnOffset-1+countRight*2, vimRowPos, '·', nil, darkGray)
			}
		}
		// top row: cursor position on grid
		screen.SetContent(cCol*2+2+columnOffset-1, 3, rune(' '), nil, cursorHighlight)

		// draw left-hand-side bar with VIM motions grid: jumps available 1..9
		vimColPos := 2
		for i := 0; i < cRow; i++ {
			countUp := cRow - i
			if countUp <= 9 && countUp > 0 {
				screen.SetContent(vimColPos, cRow+rowOffset-countUp, rune('0'+countUp), nil, darkGray)
			} else {
				screen.SetContent(vimColPos, cRow+rowOffset-countUp, '·', nil, darkGray)
			}
		}

		for i := cRow; i < game.nrows; i++ {
			countDown := game.nrows - 1 - i
			if countDown <= 9 && countDown > 0 {
				screen.SetContent(vimColPos, cRow+rowOffset+countDown, rune('0'+countDown), nil, darkGray)
			} else {
				screen.SetContent(vimColPos, cRow+rowOffset+countDown, '·', nil, darkGray)
			}
		}
		// left column: cursos position
		screen.SetContent(vimColPos, cRow+rowOffset, rune(' '), nil, cursorHighlight)
	}
	drawCursor(anchor[0]+1, anchor[1]+1)

	tui.drawMessages(game, screen, anchor[0], anchor[1])
}

func (tui *TUI) drawMessages(game *Game, screen tcell.Screen, rowAnchor, colAnchor int) {
	cRow, cCol := rowAnchor+1, colAnchor
	verticalSpace := 1
	messages := []string{
		"",
		"      move: h  j  k  l",
		"            ←  ↓  ↑  →",
		"(de)select",
		"    & eval: ctrl+v, v, or ␣ (space bar)",
		"     score: " + fmt.Sprintf("%d", game.totalScore),
		"    motion: " + fmt.Sprintf(string(tui.lastMove[:])),
		"",
	}

	debugMessages := []string{
		fmt.Sprintf("Cursor: (%d, %d)", game.cursor[0], game.cursor[1]),
	}

	debugMessages = append(debugMessages, "Selection:\n")
	for i := 0; i < nRows; i++ {
		chunk := ""
		for j := 0; j < nCols; j++ {
			chunk += fmt.Sprintf("%d ", game.selectArea[i][j])
		}
		debugMessages = append(debugMessages, chunk)
		messages = append(messages, "  ")
	}
	if tui.printDebug {
		messages = append(messages, debugMessages...)
	}
	if tui.printLogs {
		messages = append(messages, " Logs: ")
		messages = append(messages, " ")
	}
	for i, line := range messages {
		x := cCol
		for _, char := range line {
			screen.SetContent(x, nRows+i+verticalSpace+cRow, char, nil, customStyle) //tcell.StyleDefault)
			x += runewidth.RuneWidth(char)
		}
	}
}
