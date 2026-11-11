#!/bin/sh
set -eu
version=${1:?Usage: scripts/release.sh v1.2.3}
case "$version" in v[0-9]*) ;; *) echo 'Version must start with v and a digit' >&2; exit 1;; esac
if [ -n "$(git status --porcelain)" ]; then echo 'Working tree must be clean' >&2; exit 1; fi
if [ "$(git branch --show-current)" != main ] || [ "$(git rev-parse HEAD)" != "$(git rev-parse origin/main)" ]; then echo 'Release from pushed main' >&2; exit 1; fi
out="$(pwd)/release/$version"; mkdir -p "$out"
for target in darwin/arm64 darwin/amd64 linux/arm64 linux/amd64 windows/arm64 windows/amd64; do
  os=${target%/*}; arch=${target#*/}; name="mdx_${version}_${os}_${arch}"; temp=$(mktemp -d); exe=mdx
  if [ "$os" = windows ]; then exe=mdx.exe; fi
  GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o "$temp/$exe" .
  cp README.md "$temp/README.md"
  if [ "$os" = windows ]; then (cd "$temp" && zip -q "$out/$name.zip" "$exe" README.md); else tar -czf "$out/$name.tar.gz" -C "$temp" "$exe" README.md; fi
  rm -rf "$temp"
done
(cd "$out" && shasum -a 256 ./*.tar.gz ./*.zip > SHA256SUMS)
gh release create "$version" "$out"/* --target "$(git rev-parse HEAD)" --title "$version" --generate-notes
