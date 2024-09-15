package main

import (
	"bytes"
	"fmt"
	"os"

	"golang.org/x/term"
)

const (
	ESC                                   = "\033"
	CSI                                   = ESC + "["
	CSI_ERASE_DISPLAY_CURSOR_TO_END       = CSI + "0J"
	CSI_ERASE_DISPLAY_CURSOR_TO_BEGGINING = CSI + "1J"
	CSI_ERASE_DISPLAY                     = CSI + "2J"
	CSI_CURSOR_POSIION                    = CSI + "%d;%dH"
)

const (
	CMD_UP    = byte('k')
	CMD_RIGHT = byte('l')
	CMD_DOWN  = byte('j')
	CMD_LEFT  = byte('h')
)

var (
	in     *os.File
	out    *os.File
	inbuf  = make([]byte, 16)
	outbuf bytes.Buffer

	line = 1
	col  = 1
)

func main() {
	in = os.Stdin
	out = os.Stdout
	oldState, err := term.MakeRaw(int(in.Fd()))
	check(err)
	defer term.Restore(int(out.Fd()), oldState)

	for {
		check(clear())
		check(printTextBuffer())

		moveCursor(line, col)
		check(flush())

		n, err := in.Read(inbuf)
		check(err)

		check(dispatchInput(inbuf[:n]))

		check(flush())
	}
}

func check(e error) {
	if e != nil {
		os.Stderr.WriteString(e.Error())
		os.Exit(1)
	}
}

func dispatchInput(buf []byte) error {
	w, h, err := size()
	if err != nil {
		return err
	}
	switch buf[0] {
	case CMD_UP:
		line = max(1, line-1)
	case CMD_DOWN:
		line = min(h, line+1)
	case CMD_LEFT:
		col = max(1, col-1)
	case CMD_RIGHT:
		col = min(w, col+1)
	}
	return nil
}

func size() (int, int, error) {
	width, height, err := term.GetSize(int(out.Fd()))
	if err != nil {
		return 0, 0, err
	}
	return width, height, nil
}

func moveCursor(x, y int) {
	outbuf.WriteString(fmt.Sprintf(CSI_CURSOR_POSIION, x, y))
}

func printTextBuffer() error {
	_, h, err := size()
	if err != nil {
		return err
	}
	for i := 1; i <= h; i++ {
		moveCursor(i, 1)
		outbuf.WriteString("~")
	}
	return flush()
}

func clear() error {
	outbuf.Reset()
	outbuf.WriteString(CSI_ERASE_DISPLAY)
	return flush()
}

func flush() error {
	_, err := out.Write(outbuf.Bytes())
	outbuf.Reset()
	return err
}
