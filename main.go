// mdx reads Markdown and MDX documents in a terminal.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ybbapp/mdx/internal/termview"
)

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("用法: mdx <cat|lens> 文件.mdx")
	}
	if args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		fmt.Fprintln(os.Stdout, "用法: mdx <cat|lens> 文件.mdx\n\n  cat   输出到终端滚动历史\n  lens  在全屏阅读器中打开")
		return nil
	}
	mode := args[0]
	if mode != "cat" && mode != "lens" {
		return fmt.Errorf("未知命令 %q；使用 mdx help 查看用法", mode)
	}
	flags := flag.NewFlagSet(mode, flag.ContinueOnError)
	flags.Usage = func() { fmt.Fprintf(flags.Output(), "用法: mdx %s 文件.mdx\n", mode) }
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if flags.NArg() != 1 {
		flags.Usage()
		return fmt.Errorf("请指定一个文件")
	}
	switch mode {
	case "cat":
		return termview.Print(flags.Arg(0), os.Stdout)
	default:
		return termview.Run(flags.Arg(0))
	}
}
func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "mdx:", err)
		os.Exit(1)
	}
}
