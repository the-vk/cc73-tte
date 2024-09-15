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
	CSI_ERASE_LINE_CURSOR_TO_END          = CSI + "0K"
	CSI_CURSOR_POSIION                    = CSI + "%d;%dH"
)

const (
	CMD_UP    = byte('k')
	CMD_RIGHT = byte('l')
	CMD_DOWN  = byte('j')
	CMD_LEFT  = byte('h')

	CMD_MODE_INSERT = byte('i')
)

const (
	KEY_ESC = 0x1b
)

const (
	MODE_CONTROL = iota
	MODE_INSERT
)

var (
	in     *os.File
	out    *os.File
	inbuf  = make([]byte, 16)
	outbuf bytes.Buffer

	line = 1
	col  = 1

	mode = MODE_CONTROL

	textbuf = []string{"", "hello", "world"}
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
		check(printStatus())

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
	switch mode {
	case MODE_CONTROL:
		err = dispatchControl(buf)
	case MODE_INSERT:
		err = dispatchInsert(buf)
	}
	if err != nil {
		return err
	}
	line = clamp(line, 1, min(h, len(textbuf)))
	col = clamp(col, 1, min(w, len(textbuf[line-1])+1))
	return nil
}

func dispatchControl(buf []byte) error {
	switch buf[0] {
	case CMD_UP:
		line -= 1
	case CMD_DOWN:
		line += 1
	case CMD_LEFT:
		col -= 1
	case CMD_RIGHT:
		col += 1
	case CMD_MODE_INSERT:
		mode = MODE_INSERT
	}
	return nil
}

func dispatchInsert(buf []byte) error {
	switch buf[0] {
	case KEY_ESC:
		mode = MODE_CONTROL
	default:
		textbuf[line-1] = insert(textbuf[line-1], col-1, string(buf))
		col += len(buf)
	}
	return nil
}

func insert(s string, i int, t string) string {
	return s[:i] + t + s[i:]
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
	l, c := 1, 1
	for _, line := range textbuf {
		moveCursor(l, c)
		outbuf.WriteString(CSI_ERASE_LINE_CURSOR_TO_END)
		outbuf.WriteString(line)
		l++
	}
	for i := l; i <= h-1; i++ {
		moveCursor(i, 1)
		outbuf.WriteString("~")
	}
	return flush()
}

func printStatus() error {
	cur_l, cur_col := line, col
	_, h, err := size()
	if err != nil {
		return err
	}
	moveCursor(h, 1)
	outbuf.WriteString(fmt.Sprintf("mode: %s -- %d:%d", formatMode(), cur_l, cur_col))
	line, col = cur_l, cur_col
	return nil
}

func formatMode() string {
	switch mode {
	case MODE_CONTROL:
		return "control"
	case MODE_INSERT:
		return "insert"
	}
	return ""
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
