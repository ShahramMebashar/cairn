#!/bin/sh
# cairn installer — downloads the right prebuilt binary for this machine.
#   curl -fsSL https://github.com/ShahramMebashar/cairn/releases/latest/download/install.sh | sh
# Override with: REPO, VERSION (default latest), BINDIR (default /usr/local/bin).
set -eu

REPO="${REPO:-ShahramMebashar/cairn}"
VERSION="${VERSION:-latest}"
BINDIR="${BINDIR:-/usr/local/bin}"

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
case "$arch" in
  x86_64 | amd64) arch="amd64" ;;
  arm64 | aarch64) arch="arm64" ;;
  *) echo "cairn: unsupported architecture: $arch" >&2; exit 1 ;;
esac
case "$os" in
  linux | darwin) ;;
  *) echo "cairn: unsupported OS: $os (use the Windows zip from Releases, or the desktop app)" >&2; exit 1 ;;
esac

if [ "$VERSION" = "latest" ]; then
  base="https://github.com/$REPO/releases/latest/download"
else
  base="https://github.com/$REPO/releases/download/$VERSION"
fi
archive="cairn_${os}_${arch}.tar.gz"
url="$base/$archive"
checksums_url="$base/checksums.txt"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

echo "cairn: downloading $url"
curl -fsSL "$url" -o "$tmp/$archive"
curl -fsSL "$checksums_url" -o "$tmp/checksums.txt"

expected="$(awk -v file="$archive" '$2 == file { print $1 }' "$tmp/checksums.txt")"
if [ -z "$expected" ]; then
  echo "cairn: checksum not found for $archive" >&2
  exit 1
fi

if command -v sha256sum >/dev/null 2>&1; then
  actual="$(sha256sum "$tmp/$archive" | awk '{ print $1 }')"
elif command -v shasum >/dev/null 2>&1; then
  actual="$(shasum -a 256 "$tmp/$archive" | awk '{ print $1 }')"
else
  echo "cairn: need sha256sum or shasum to verify download" >&2
  exit 1
fi

if [ "$actual" != "$expected" ]; then
  echo "cairn: checksum mismatch for $archive" >&2
  exit 1
fi

tar -xzf "$tmp/$archive" -C "$tmp"

if [ -w "$BINDIR" ]; then
  install -m 0755 "$tmp/cairn" "$BINDIR/cairn"
else
  echo "cairn: $BINDIR not writable, using sudo"
  sudo install -m 0755 "$tmp/cairn" "$BINDIR/cairn"
fi

echo "cairn: installed to $BINDIR/cairn"
"$BINDIR/cairn" version
