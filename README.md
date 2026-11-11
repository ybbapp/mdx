# mdx

在终端阅读 Markdown 和 MDX。支持基本文档排版以及常见 MDX 组件的终端交互，不需要浏览器或图像协议。

## 安装

需要 Go 1.27.3 或更新版本：

```sh
go install github.com/ybbapp/mdx@latest
```

也可以从 [Releases](https://github.com/ybbapp/mdx/releases) 下载预编译程序。

## 使用

```sh
mdx peek document.mdx  # 内联交互，使用终端原生滚动历史
mdx cat document.mdx   # 直接输出，适合管道
mdx lens document.mdx  # 全屏阅读器
```

`peek` 中用 Tab 选择操作项、Enter 操作、`q` 退出。文件变化或组件状态改变时，新版本会追加到终端历史；调整窗口后按 `r` 按新宽度重排。鼠标滚动和文本选择由终端处理；当前不支持鼠标点击组件。

`cat` 展开已支持的标签和折叠内容，不执行交互；重定向到文件或管道时不输出颜色控制符。`lens` 中用方向键或 `j` / `k` 滚动，Tab 选择、Enter 操作，`/` 搜索，`q` 退出。

目前支持单独占一行的 `<Tabs>` / `<TabItem label="...">`、`<Accordion title="...">` 和 `<Callout title="...">`。未知的独占行组件会显示提示并保留子内容。不会执行 MDX 表达式、import 或自定义 React 组件，也不支持任意 JSX 布局。

## 发布

在干净且已推送的 main 分支运行 `scripts/release.sh vX.Y.Z`。脚本本地交叉编译 macOS、Linux 和 Windows 的 amd64、arm64 二进制文件，并通过 GitHub CLI 上传；不依赖 GitHub Actions。
