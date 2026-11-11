package termview

import (
	"fmt"
	"regexp"
	"strings"
)

type node struct {
	kind, tag, text, label string
	children               []*node
	id                     int
}
type action struct {
	id                  int
	kind, label, target string
	choice              int
}
type document struct {
	root    *node
	choices map[int]int
	opened  map[int]bool
}

var tagLine = regexp.MustCompile(`^\s*<(/?)([A-Za-z][A-Za-z0-9.]*)\b([^>]*)>\s*$`)
var attr = regexp.MustCompile(`\b(label|title|value|name)\s*=\s*["']([^"']*)["']`)
var link = regexp.MustCompile(`\[([^\]\n]+)\]\(([^\s)]+)(?:\s+"[^"]*")?\)`)

// parseDocument recognizes line-oriented JSX components, without executing JS.
func parseDocument(src string) *document {
	root := &node{kind: "root"}
	stack := []*node{root}
	lines := strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n")
	fence := ""
	inFront := len(lines) > 0 && strings.TrimSpace(lines[0]) == "---"
	for i, line := range lines {
		if inFront {
			if i > 0 && strings.TrimSpace(line) == "---" {
				inFront = false
			}
			continue
		}
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "```") || strings.HasPrefix(trim, "~~~") {
			marker := trim[:3]
			if fence == "" {
				fence = marker
			} else if fence == marker {
				fence = ""
			}
		}
		parent := stack[len(stack)-1]
		if fence == "" {
			if m := tagLine.FindStringSubmatch(line); m != nil {
				if m[1] == "/" {
					if len(stack) > 1 && stack[len(stack)-1].tag == m[2] {
						stack = stack[:len(stack)-1]
						continue
					}
				} else {
					label := ""
					for _, a := range attr.FindAllStringSubmatch(m[3], -1) {
						if label == "" || a[1] == "label" || a[1] == "title" {
							label = a[2]
						}
					}
					n := &node{kind: "component", tag: m[2], label: label, id: i + 1}
					parent.children = append(parent.children, n)
					if !strings.HasSuffix(strings.TrimSpace(m[3]), "/") {
						stack = append(stack, n)
					}
					continue
				}
			}
		}
		parent.children = append(parent.children, &node{kind: "text", text: line, id: i + 1})
	}
	return &document{root: root, choices: map[int]int{}, opened: map[int]bool{}}
}

func (d *document) markdown() (string, []action) { return d.render(false) }

func (d *document) render(inline bool) (string, []action) {
	var out strings.Builder
	var actions []action
	var walk func(*node)
	emitText := func(n *node) {
		line := n.text
		if strings.HasPrefix(strings.TrimSpace(line), "import ") || strings.HasPrefix(strings.TrimSpace(line), "export ") {
			out.WriteString("> ⚠ MDX import/export 未执行\n\n")
			return
		}
		out.WriteString(line)
		out.WriteByte('\n')
	}
	walk = func(n *node) {
		if n.kind == "text" {
			emitText(n)
			return
		}
		switch n.tag {
		case "Tabs":
			var items []*node
			for _, child := range n.children {
				if child.tag == "TabItem" || child.tag == "Tab" {
					items = append(items, child)
				}
			}
			if len(items) == 0 {
				out.WriteString("> ⚠ Tabs 没有可识别的 TabItem\n\n")
				for _, c := range n.children {
					walk(c)
				}
				return
			}
			if inline {
				for i, item := range items {
					label := item.label
					if label == "" {
						label = fmt.Sprintf("Tab %d", i+1)
					}
					fmt.Fprintf(&out, "\n**%s**\n\n", label)
					for _, c := range item.children {
						walk(c)
					}
				}
				return
			}
			choice := d.choices[n.id]
			if choice >= len(items) {
				choice = 0
			}
			out.WriteString("\n**标签：** ")
			for i, item := range items {
				label := item.label
				if label == "" {
					label = fmt.Sprintf("Tab %d", i+1)
				}
				actions = append(actions, action{id: n.id, kind: "tab", label: label, choice: i})
				marker := "○"
				if i == choice {
					marker = "●"
				}
				fmt.Fprintf(&out, "%s %s `[%d]`  ", marker, label, len(actions))
			}
			out.WriteString("\n\n")
			for _, c := range items[choice].children {
				walk(c)
			}
		case "Accordion", "AccordionItem", "details":
			label := n.label
			if label == "" {
				label = "展开内容"
			}
			if inline {
				fmt.Fprintf(&out, "\n**%s**\n\n", label)
				for _, c := range n.children {
					walk(c)
				}
				return
			}
			actions = append(actions, action{id: n.id, kind: "toggle", label: label})
			marker := "▸"
			if d.opened[n.id] {
				marker = "▾"
			}
			fmt.Fprintf(&out, "\n**%s %s** `[%d]`\n\n", marker, label, len(actions))
			if d.opened[n.id] {
				for _, c := range n.children {
					walk(c)
				}
			}
		case "Callout", "Alert", "Note":
			label := n.label
			if label == "" {
				label = n.tag
			}
			fmt.Fprintf(&out, "\n> **%s**\n\n", label)
			for _, c := range n.children {
				walk(c)
			}
		case "", "TabItem", "Tab":
			for _, c := range n.children {
				walk(c)
			}
		default:
			fmt.Fprintf(&out, "\n> ⚠ 未支持组件 `%s`；以下显示其文本内容。\n\n", n.tag)
			for _, c := range n.children {
				walk(c)
			}
		}
	}
	// Inject link controls only into ordinary text, not fenced code.
	walk(d.root)
	// Links remain visible as Markdown; enumerate them separately for navigation.
	text := out.String()
	if inline {
		return text, nil
	}
	fence := false
	for _, line := range strings.Split(text, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "```") || strings.HasPrefix(trim, "~~~") {
			fence = !fence
			continue
		}
		if fence {
			continue
		}
		for _, m := range link.FindAllStringSubmatch(line, -1) {
			actions = append(actions, action{kind: "link", label: m[1], target: m[2]})
		}
	}
	return text, actions
}
