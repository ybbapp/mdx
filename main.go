// mdx reads Markdown and MDX documents in a terminal.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ybbapp/mdx/internal/termview"
)

func parseArgs(args []string) (string, string, error) {
	if len(args) == 0 {
		return "", "", fmt.Errorf("请指定一个文件；使用 mdx --help 查看用法")
	}
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help" || args[0] == "help") {
		return "help", "", nil
	}
	if args[0] == "cat" {
		if len(args) != 2 {
			return "", "", fmt.Errorf("用法: mdx cat 文件.mdx")
		}
		return "cat", args[1], nil
	}
	if len(args) != 1 || strings.HasPrefix(args[0], "-") {
		return "", "", fmt.Errorf("用法: mdx 文件.mdx 或 mdx cat 文件.mdx")
	}
	return "view", args[0], nil
}
func run(args []string) error {
	mode, path, err := parseArgs(args)
	if err != nil {
		return err
	}
	switch mode {
	case "help":
		fmt.Fprintln(os.Stdout, "用法: mdx 文件.mdx      交互式阅读\n      mdx cat 文件.mdx  直接输出到终端或管道")
		return nil
	case "cat":
		return termview.Print(path, os.Stdout)
	default:
		return termview.Run(path)
	}
}
func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "mdx:", err)
		os.Exit(1)
	}
}
