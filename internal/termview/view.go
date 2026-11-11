// Package termview implements a terminal-native, non-executing MDX reader.
package termview

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
	"golang.org/x/term"
)

// Print writes a static rendering to the terminal's normal scrollback.
// All known tab panels and accordion sections are expanded in this mode.
func Print(path string, output io.Writer) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	doc := parseDocument(string(raw))
	md, _ := doc.render(true)
	width := 80
	color := false
	if f, ok := output.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		if w, _, e := term.GetSize(int(f.Fd())); e == nil && w >= 20 {
			width = w
		}
		color = true
	}
	opts := []glamour.TermRendererOption{glamour.WithStandardStyle("dark"), glamour.WithWordWrap(width - 2)}
	if !color {
		opts = append(opts, glamour.WithColorProfile(termenv.Ascii))
	}
	renderer, err := glamour.NewTermRenderer(opts...)
	if err != nil {
		return err
	}
	rendered, err := renderer.Render(md)
	if err != nil {
		return err
	}
	if !color {
		rendered = ansi.Strip(rendered)
	}
	_, err = io.WriteString(output, rendered)
	return err
}

func openLink(base, target string) error {
	u, err := url.Parse(target)
	if err != nil {
		return err
	}
	if u.Scheme != "" && u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("仅支持 HTTP(S) 和本地文件链接")
	}
	if u.Scheme == "" {
		if strings.HasPrefix(target, "#") {
			return errors.New("页内锚点暂不支持")
		}
		target = filepath.Join(base, filepath.FromSlash(u.Path))
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", target)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", target)
	default:
		cmd = exec.Command("xdg-open", target)
	}
	return cmd.Start()
}
func Run(path string) error {
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return errors.New("终端预览需要交互式终端")
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	doc := parseDocument(string(raw))
	last := string(raw)
	old, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	defer term.Restore(int(os.Stdin.Fd()), old)
	fmt.Print("\x1b[?1049h\x1b[?25l")
	defer fmt.Print("\x1b[?25h\x1b[?1049l")
	input := make(chan byte, 32)
	go func() {
		b := make([]byte, 1)
		for {
			if _, e := os.Stdin.Read(b); e != nil {
				return
			}
			select {
			case input <- b[0]:
			default:
			}
		}
	}()
	offset, selected := 0, 0
	var actions []action
	var lines []string
	message := ""
	search := ""
	searching := false
	searchIndex := -1
	pendingG := false
	draw := func() error {
		width, height, e := term.GetSize(int(os.Stdout.Fd()))
		if e != nil {
			return e
		}
		if width < 20 || height < 6 {
			return errors.New("终端窗口太小")
		}
		md, a := doc.markdown()
		actions = a
		if selected >= len(actions) {
			selected = 0
		}
		r, e := glamour.NewTermRenderer(glamour.WithStandardStyle("dark"), glamour.WithWordWrap(width-2))
		if e != nil {
			return e
		}
		rendered, e := r.Render(md)
		if e != nil {
			return e
		}
		lines = strings.Split(strings.ReplaceAll(rendered, "\r\n", "\n"), "\n")
		page := height - 3
		max := len(lines) - page
		if max < 0 {
			max = 0
		}
		if offset > max {
			offset = max
		}
		if offset < 0 {
			offset = 0
		}
		var b strings.Builder
		b.WriteString("\x1b[H\x1b[2J")
		fmt.Fprintf(&b, " mdx · %s\x1b[K\r\n", ansi.Truncate(filepath.Base(path), width-12, "…"))
		for i := 0; i < page; i++ {
			if offset+i < len(lines) {
				b.WriteString(lines[offset+i])
			}
			b.WriteString("\x1b[K\r\n")
		}
		info := "↑↓/jk 滚动 · Tab 选择 · Enter 操作 · / 搜索 · q 退出"
		if len(actions) > 0 {
			a := actions[selected]
			info = fmt.Sprintf("[%d/%d] %s: %s · Tab 选择 · Enter 操作", selected+1, len(actions), a.kind, a.label)
		}
		if message != "" {
			info = message
		}
		if searching {
			info = "搜索: " + search
		}
		b.WriteString(ansi.Truncate(info, width-1, "…"))
		b.WriteString("\x1b[K")
		_, e = os.Stdout.WriteString(b.String())
		return e
	}
	find := func(direction int) {
		q := strings.ToLower(search)
		if q == "" || len(lines) == 0 {
			return
		}
		start := searchIndex
		if start < 0 {
			start = offset - direction
		}
		for step := 1; step <= len(lines); step++ {
			i := (start + direction*step) % len(lines)
			if i < 0 {
				i += len(lines)
			}
			if strings.Contains(strings.ToLower(ansi.Strip(lines[i])), q) {
				searchIndex = i
				offset = i
				message = "找到: " + search
				return
			}
		}
		message = "未找到: " + search
	}

	if err = draw(); err != nil {
		return err
	}
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if b, e := os.ReadFile(path); e == nil && string(b) != last {
				next := parseDocument(string(b))
				next.choices = doc.choices
				next.opened = doc.opened
				doc = next
				last = string(b)
				message = "文件已刷新"
			}
			if err = draw(); err != nil {
				return err
			}
		case key := <-input:
			if searching {
				switch key {
				case 13, 10:
					searching = false
					searchIndex = -1
					find(1)
				case 27:
					searching = false
					search = ""
				case 127, 8:
					if len(search) > 0 {
						runes := []rune(search)
						search = string(runes[:len(runes)-1])
					}
				default:
					if key >= 32 {
						search += string([]byte{key})
					}
				}
			} else {
				message = ""
				if key != 'g' {
					pendingG = false
				}
				switch key {
				case 'q', 3:
					return nil
				case 'j':
					offset++
				case 'k':
					offset--
				case 'g':
					if pendingG {
						offset = 0
						pendingG = false
					} else {
						pendingG = true
					}
				case 'G':
					offset = len(lines)
				case ' ', 6: // Ctrl-f
					_, h, _ := term.GetSize(int(os.Stdout.Fd()))
					offset += h - 3
				case 2: // Ctrl-b
					_, h, _ := term.GetSize(int(os.Stdout.Fd()))
					offset -= h - 3
				case 4: // Ctrl-d
					_, h, _ := term.GetSize(int(os.Stdout.Fd()))
					offset += (h - 3) / 2
				case 21: // Ctrl-u
					_, h, _ := term.GetSize(int(os.Stdout.Fd()))
					offset -= (h - 3) / 2
				case '/':
					search = ""
					searching = true
				case 'n':
					searchIndex = -1
					find(1)
				case '\t':
					if len(actions) > 0 {
						selected = (selected + 1) % len(actions)
					}
				case 13, 10:
					if len(actions) > 0 {
						a := actions[selected]
						switch a.kind {
						case "tab":
							doc.choices[a.id] = a.choice
						case "toggle":
							doc.opened[a.id] = !doc.opened[a.id]
						case "link":
							if e := openLink(filepath.Dir(path), a.target); e != nil {
								message = e.Error()
							} else {
								message = "已打开链接: " + a.label
							}
						}
					}
				case 27:
					select {
					case b := <-input:
						if b == '[' {
							select {
							case c := <-input:
								if c == 'A' {
									offset--
								} else if c == 'B' {
									offset++
								}
							case <-time.After(40 * time.Millisecond):
							}
						}
					case <-time.After(40 * time.Millisecond):
					}
				}
			}
			if err = draw(); err != nil {
				return err
			}
		}
	}
}
