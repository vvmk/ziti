package ziti

import (
	"syscall"
	"time"

	termbox "github.com/nsf/termbox-go"
)

/* Ziti -- A very simple editor in less than 1000 lines of Go code .
 *
 * -----------------------------------------------------------------------
 *
 * Copyright (c) 2018, Kristofer Younger <kryounger at gmail dot com>
 * (who ripped out all the highlighting stuff)
 * based on the work of https://github.com/antirez/kilo
 * which was marked
 * Copyright (C) 2016 Salvatore Sanfilippo <antirez at gmail dot com>
 *
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are
 *
 *  *  Redistributions of source code must retain the above copyright
 *     notice, this list of conditions and the following disclaimer.
 * met:
 *
 *  *  Redistributions in binary form must reproduce the above copyright
 *     notice, this list of conditions and the following disclaimer in the
 *     documentation and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
 * "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
 * LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTabILITY AND FITNESS FOR
 * A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
 * HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
 * SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
 * LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES LOSS OF USE,
 * DATA, OR PROFITS OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
 * THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
 * OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */

const zitiVersion = "1.2"

// Ziti is the top-level exported type
type Ziti struct {
	ziti *editor
}

/* This structure represents a single line of the file we are editing. */
type erow struct {
	idx    int    /* Row index in the file, zero-based. */
	size   int    /* Size of the row, excluding the null term. */
	rsize  int    /* Size of the rendered row. */
	runes  []rune //string /* Row content. */
	render []rune /* Row content "rendered" for screen (for Tabs). */
}

type cursor struct {
	r  int /* Cursor row position */
	c  int /* Cursor column position */
	ro int /* Cursor rowoffset */
	co int /* Cursor coloffset */
}
type editor struct {
	events        chan termbox.Event
	point         cursor
	mark          cursor
	markSet       bool
	screenrows    int /* Number of rows that we can show */
	screencols    int /* Number of cols that we can show */
	numrows       int /* Number of rows in file */
	TextSize      int
	quitTimes     int
	done          bool
	row           []*erow /* Rows */
	dirty         bool    /* File modified but not saved. */
	filename      string  /* Currently open filename */
	statusmsg     string
	statusmsgTime time.Time
	pastebuffer   string
	fgcolor       termbox.Attribute
	bgcolor       termbox.Attribute
}

func (e *editor) checkErr(er error) {
	if er != nil {
		e.editorSetStatusMessage("%s", er)
	}
}

// KEY constants
const (
	KeyNull   = 0  /* NULL ctrl-space set mark */
	CtrlA     = 1  /* Ctrl-a BOL */
	CtrlC     = 3  /* Ctrl-c  cop */
	CtrlE     = 5  /* Ctrl-e  EOL */
	CtrlD     = 4  /* Ctrl-d del forward? */
	CtrlF     = 6  /* Ctrl-f find */
	CtrlH     = 8  /* Ctrl-h del backward*/
	Tab       = 9  /* Tab */
	CtrlL     = 12 /* Ctrl+l redraw */
	Enter     = 13 /* Enter */
	CtrlQ     = 17 /* Ctrl-q quit*/
	CtrlS     = 19 /* Ctrl-s save*/
	CtrlU     = 21 /* Ctrl-u number of times??*/
	CtrlV     = 22 /* Ctrl-V paste */
	CtrlW     = 23
	CtrlX     = 24 /* Ctrl-X cut */
	CtrlY     = 25
	CtrlZ     = 26
	Esc       = 27  /* Escape */
	Space     = 32  /* Space */
	Backspace = 127 /* Backspace */
)

// Cursor movement keys
const (
	ArrowLeft  = termbox.KeyArrowLeft
	ArrowRight = termbox.KeyArrowRight
	ArrowUp    = termbox.KeyArrowUp
	ArrowDown  = termbox.KeyArrowDown
	DelKey     = termbox.KeyDelete
	HomeKey    = termbox.KeyHome
	EndKey     = termbox.KeyEnd
	PageUp     = termbox.KeyPgup
	PageDown   = termbox.KeyPgdn
)

// State contains the state of a terminal.
type State struct {
	termios syscall.Termios
}

// NewEditor generates a new editor for use
func (e *editor) initEditor() {
	e.done = false
	e.point.c = 0
	e.point.r = 0
	e.point.ro = 0
	e.point.co = 0
	e.numrows = 0
	e.row = nil
	e.dirty = false
	e.filename = ""
	e.screencols, e.screenrows = termbox.Size()
	e.screenrows -= 2 /* Get room for status bar. */
	e.quitTimes = 3
	e.fgcolor = termbox.ColorDefault //termbox.ColorBlack
	e.bgcolor = termbox.ColorDefault //termbox.ColorWhite
}

// Start runs an editor
func (z *Ziti) Start(filename string) {

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	z.ziti = &editor{}
	e := z.ziti
	e.initEditor()

	e.editorOpen(filename)
	termbox.SetOutputMode(termbox.OutputNormal)
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	e.editorSetStatusMessage("HELP: Ctrl-S = save | Ctrl-Q = quit | Ctrl-F = find")

	e.events = make(chan termbox.Event, 20)
	go func() {
		for {
			e.events <- termbox.PollEvent()
		}
	}()
	for e.done != true {
		e.editorRefreshScreen()
		select {
		case ev := <-e.events:
			e.editorProcessEvent(ev)
			termbox.Flush()
		}
	}

}
