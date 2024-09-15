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
	if err != nil {
		fmt.Println("failed to make raw:", err)
		os.Exit(1)
	}
	defer term.Restore(int(out.Fd()), oldState)

	if err := clear(); err != nil {
		fmt.Println("failed to clear:", err)
		os.Exit(1)
	}

	for {
		if err := printTextBuffer(); err != nil {
			os.Exit(1)
		}

		n, err := in.Read(inbuf)
		if err != nil {
			break
		}
		os.Stdout.Write([]byte(fmt.Sprintf("input (%d): %s\n", n, string(inbuf))))
	}
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
