package termview

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/x/ansi"
)

func TestComponentsAndControls(t *testing.T) {
	src := `---
title: Test
---
# Hello
<Tabs>
<TabItem label="Alpha">
First [link](https://example.com)
</TabItem>
<TabItem label="Beta">
Second
</TabItem>
</Tabs>
<Accordion title="More">
Hidden text
</Accordion>
<Callout title="Note">
Read this
</Callout>
<Custom>
Preserved
</Custom>`
	doc := parseDocument(src)
	text, actions := doc.markdown()
	for _, s := range []string{"# Hello", "First", "Note", "Preserved", "未支持组件"} {
		if !strings.Contains(text, s) {
			t.Errorf("missing %q in %s", s, text)
		}
	}
	if strings.Contains(text, "Second") || strings.Contains(text, "Hidden text") {
		t.Fatal("hidden content shown")
	}
	if len(actions) != 4 || actions[0].kind != "tab" || actions[2].kind != "toggle" || actions[3].kind != "link" {
		t.Fatalf("controls: %+v", actions)
	}
	doc.choices[actions[1].id] = actions[1].choice
	doc.opened[actions[2].id] = true
	text, _ = doc.markdown()
	if !strings.Contains(text, "Second") || !strings.Contains(text, "Hidden text") || strings.Contains(text, "First") {
		t.Fatalf("interaction failed: %s", text)
	}
}
func TestCodeFencesAreNotComponents(t *testing.T) {
	doc := parseDocument("```jsx\n<Custom>\n[not a link](https://example.com)\n</Custom>\n```")
	text, actions := doc.markdown()
	if !strings.Contains(text, "<Custom>") || len(actions) != 0 {
		t.Fatalf("code altered: %s, %+v", text, actions)
	}
}

func TestInlineExpandsInteractiveContent(t *testing.T) {
	doc := parseDocument("<Tabs>\n<TabItem label=\"A\">\nFirst\n</TabItem>\n<TabItem label=\"B\">\nSecond\n</TabItem>\n</Tabs>\n<Accordion title=\"Details\">\nHidden\n</Accordion>")
	text, actions := doc.render(true)
	for _, want := range []string{"First", "Second", "Hidden"} {
		if !strings.Contains(text, want) {
			t.Fatalf("inline missing %q: %s", want, text)
		}
	}
	if len(actions) != 0 || strings.Contains(text, "`[1]`") {
		t.Fatalf("interactive controls in inline mode: %s", text)
	}
}

func TestPrintWritesPlainTextToPipe(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file.mdx")
	if err := os.WriteFile(path, []byte("# Hello\n"), 0600); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := Print(path, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Hello") || strings.Contains(out.String(), "\x1b") {
		t.Fatalf("unexpected output: %q", out.String())
	}
}

func TestWrappedURLFragmentsOpenFullDestination(t *testing.T) {
	target := "https://www.1stchoice-formations.co.uk/address-services/london-registered-office-service/"
	renderer, err := glamour.NewTermRenderer(glamour.WithStandardStyle("dark"), glamour.WithWordWrap(40))
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := renderer.Render("[查看服务详情](" + target + ")")
	if err != nil {
		t.Fatal(err)
	}
	result := clickableLinks(rendered, []action{{kind: "link", target: target}})
	if strings.Count(result, "\x1b]8;;"+target+"\x1b\\") < 2 {
		t.Fatalf("wrapped fragments lack full hyperlink: %q", result)
	}
	if ansi.Strip(result) != ansi.Strip(rendered) {
		t.Fatal("link controls changed displayed text")
	}
}

func TestViewerHasNoFooter(t *testing.T) {
	if got := pageLines(30); got != 29 {
		t.Fatalf("page lines = %d, want 29", got)
	}
	if got := pageLines(1); got != 0 {
		t.Fatalf("page lines for short terminal = %d, want 0", got)
	}
	if got := viewerTitle("doc.mdx", "", "", false); got != "mdx · doc.mdx" {
		t.Fatalf("title = %q", got)
	}
	if got := viewerTitle("doc.mdx", "已打开链接", "", false); got != "已打开链接" {
		t.Fatalf("message title = %q", got)
	}
	if got := viewerTitle("doc.mdx", "", "Tabs", true); got != "搜索: Tabs · Enter 查找 · Esc 取消" {
		t.Fatalf("search title = %q", got)
	}
}
