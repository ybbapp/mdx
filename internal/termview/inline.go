package termview

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/x/ansi"
	"golang.org/x/term"
)

// RunInline keeps the reader in the normal screen and writes revisions to scrollback.
func RunInline(path string) error {
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return errors.New("交互模式需要交互式终端")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	raw, err := os.ReadFile(abs)
	if err != nil {
		return err
	}
	last := string(raw)
	doc := parseDocument(last)
	old, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	defer term.Restore(int(os.Stdin.Fd()), old)
	defer func() { fmt.Print("\x1b[?25h\r\n") }()
	input := make(chan byte, 128)
	go func() {
		b := make([]byte, 1)
		for {
			if _, e := os.Stdin.Read(b); e != nil {
				return
			}
			input <- b[0]
		}
	}()
	var actions []action
	selected := 0
	width, height := 0, 0
	status := ""
	revision := true
	redraw := func() error {
		w, h, e := term.GetSize(int(os.Stdout.Fd()))
		if e != nil {
			return e
		}
		if w < 30 || h < 8 {
			return errors.New("终端窗口太小")
		}
		if w != width || h != height {
			if width != 0 {
				status = "窗口已调整；按 r 重新排版"
			}
			width, height = w, h
		}
		md, a := doc.markdown()
		actions = a
		if selected >= len(actions) {
			selected = 0
		}
		if revision {
			renderer, e := glamour.NewTermRenderer(glamour.WithStandardStyle("dark"), glamour.WithWordWrap(width-2))
			if e != nil {
				return e
			}
			out, e := renderer.Render(md)
			if e != nil {
				return e
			}
			lines := strings.Split(strings.ReplaceAll(out, "\r\n", "\n"), "\n")
			for len(lines) > 0 && lines[len(lines)-1] == "" {
				lines = lines[:len(lines)-1]
			}
			// Start at the bottom so each revision goes into normal terminal scrollback.
			fmt.Fprintf(os.Stdout, "\x1b[%d;1H\r\n", height)
			for _, line := range lines {
				fmt.Fprintf(os.Stdout, "%s\x1b[K\r\n", line)
			}
			revision = false
		}
		info := "Tab:选择 Enter:操作 r:重排 q:退出"
		if len(actions) > 0 {
			info += fmt.Sprintf(" %d/%d %s", selected+1, len(actions), actions[selected].label)
		}
		if status != "" {
			info += "  " + status
		} else {
			info += ""
		}
		fmt.Fprintf(os.Stdout, "\x1b[%d;1H\x1b[2K%s", height, ansi.Truncate(info, width-1, "…"))
		return nil
	}
	activate := func() {
		if len(actions) == 0 {
			return
		}
		a := actions[selected]
		switch a.kind {
		case "tab":
			doc.choices[a.id] = a.choice
			revision = true
		case "toggle":
			doc.opened[a.id] = !doc.opened[a.id]
			revision = true
		case "link":
			if err := openLink(filepath.Dir(abs), a.target); err != nil {
				status = err.Error()
			} else {
				status = "已打开: " + a.label
			}
		}
	}
	if err = redraw(); err != nil {
		return err
	}
	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			needsRedraw := false
			w, h, sizeErr := term.GetSize(int(os.Stdout.Fd()))
			if sizeErr == nil && (w != width || h != height) {
				needsRedraw = true
			}
			b, e := os.ReadFile(abs)
			if e == nil && string(b) != last {
				next := parseDocument(string(b))
				next.choices = doc.choices
				next.opened = doc.opened
				doc = next
				last = string(b)
				revision = true
				status = "文件已刷新"
				needsRedraw = true
			}
			if needsRedraw {
				if err = redraw(); err != nil {
					return err
				}
			}
		case key := <-input:
			status = ""
			switch key {
			case 'r':
				revision = true
				status = "已重新排版"
			case 'q', 3:
				return nil
			case '\t', 'l':
				if len(actions) > 0 {
					selected = (selected + 1) % len(actions)
				}
			case 'h':
				if len(actions) > 0 {
					selected = (selected + len(actions) - 1) % len(actions)
				}
			case 13, 10:
				activate()
			}
			if err = redraw(); err != nil {
				return err
			}
		}
	}
}
