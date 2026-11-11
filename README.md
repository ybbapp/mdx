# mdx

在终端阅读 Markdown 和 MDX。支持基本文档排版以及常见 MDX 组件的终端交互，不需要浏览器或图像协议。

## 安装

### Homebrew

macOS 或 Linux：

```sh
brew install ybbapp/tap/mdx
```

升级时运行 `brew upgrade ybbapp/tap/mdx`。Homebrew 安装不需要 Go。

### Go

需要 Go 1.27.3 或更新版本：

```sh
go install github.com/ybbapp/mdx@latest
```

也可以从 [Releases](https://github.com/ybbapp/mdx/releases) 下载 macOS、Linux 或 Windows 预编译程序。
## 使用

```sh
mdx document.mdx       # 交互式阅读
mdx cat document.mdx   # 直接输出，适合管道
mdx examples/sample.mdx # 在源码仓库中试用示例
```

通过 Homebrew 安装后，可以运行 `mdx "$(brew --prefix)/share/mdx/sample.mdx"` 打开随包提供的示例。

`cat` 展开已支持的标签和折叠内容，不执行交互；重定向到文件或管道时不输出颜色控制符。

交互式阅读器支持 Vim 风格的导航：`j` / `k` 或方向键逐行滚动，`Ctrl-d` / `Ctrl-u` 翻半页，`Ctrl-f` / `Ctrl-b` 翻整页，`gg` 跳到开头，`G` 跳到末尾。按 `/` 搜索，`n` / `N` 查找下一个或上一个匹配项；Tab 选择交互项，Enter 操作，`q` 退出。文件修改后自动刷新。

目前支持单独占一行的 `<Tabs>` / `<TabItem label="...">`、`<Accordion title="...">` 和 `<Callout title="...">`。未知的独占行组件会显示提示并保留子内容。不会执行 MDX 表达式、import 或自定义 React 组件，也不支持任意 JSX 布局。

## 发布

在干净且已推送的 main 分支运行 `scripts/release.sh vX.Y.Z`。脚本本地交叉编译 macOS、Linux 和 Windows 的 amd64、arm64 二进制文件，并通过 GitHub CLI 上传；不依赖 GitHub Actions。发布新版本后，还需要更新 [Homebrew formula](https://github.com/ybbapp/homebrew-tap/blob/main/Formula/mdx.rb) 中的版本、各平台下载地址和 SHA-256，然后推送 tap。
