package game

import (
	"cadny-cursh/src/utils"
	"time"

	"github.com/nsf/termbox-go"
)

var (
	gameOver = false
)

type board [][]candy

type coord struct {
	x, y int
}

type cursor coord

type displayElements struct {
	points int
	msg    string
}

type level struct {
	board                  board
	posX, posY, xmax, ymax int
	cursor                 cursor
	isSelected             bool
	blinkCh                chan bool
	visuals                displayElements
	movesLeft              int
}

func (d *displayElements) showMsg(msg string) {
	d.msg = msg
	time.Sleep(time.Second * 2)
	d.msg = ""
}

func (d *displayElements) addPoint(points int) {
	d.points += points
}

func (lev *level) startBlink() {
	go lev.blinkCursor(lev.blinkCh)
}

func (lev *level) stopBlink() {
	lev.blinkCh <- true
}

func newLevel(rowCount, colCount, posX, posY int) *level {
	newBoard := make(board, rowCount)
	for i := range newBoard {
		newBoard[i] = make([]candy, colCount)
	}
	l := level{
		board:      newBoard,
		posX:       posX,
		posY:       posY,
		xmax:       colCount - 1,
		ymax:       rowCount - 1,
		cursor:     cursor{},
		isSelected: false,
		blinkCh:    make(chan bool),
		visuals:    displayElements{0, ""},
		movesLeft:  30,
	}
	return &l
}

func (lev *level) handleKeyboardEvent(kEvent keyboardEvent, kbProc *keyboardEvProcess) bool {
	lev.stopBlink()
	lev.repaintCurCell()
	switch kEvent.eventType {
	case NAVIGATE:
		lev.navigate(kEvent)
	case SELECT:
		if selected := lev.toggleSelected(); selected {
			go lev.blinkAdjacent()
		} else {
			adjacentColors.repaintCells(lev)
		}
	case MOVE:
		kbProc.pause()
		lev.move(kEvent)
		kbProc.resume()
	case END:
		return true
	}
	lev.startBlink()
	return false
}

func Start() {
	if err := termbox.Init(); err != nil {
		panic(err)
	}
	defer func() {
		termbox.Clear(defaultColor, defaultColor)
		termbox.Flush()
		termbox.Close()
		utils.Clrscr()
	}()

	lev := newLevel(8, 8, 5, 5)

	lev.initBoard()

	var keyboardChan chan keyboardEvent = make(chan keyboardEvent)

	var kbProc keyboardEvProcess = false

	go listenToKeyboard(&lev.isSelected, keyboardChan, &kbProc)

mainloop:
	for {
		select {
		case e := <-keyboardChan:
			if breakLoop := lev.handleKeyboardEvent(e, &kbProc); breakLoop {
				break mainloop
			}
		default:
			if gameOver {
				renderGameOver(lev.visuals.points)
				<-keyboardChan
				break mainloop
			} else {
				lev.render()
			}
			time.Sleep(time.Millisecond * 10)
		}
	}
}
