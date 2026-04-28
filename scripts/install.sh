#!/bin/sh
set -eu

repo="${YUNXIAO_INSTALL_REPO:-nick3/yunxiao-cli}"
api_url="${YUNXIAO_INSTALL_API_URL:-https://api.github.com}"
download_base_url="${YUNXIAO_INSTALL_DOWNLOAD_BASE_URL:-https://github.com}"
version=""
bin_dir=""
dry_run=0
uninstall=0

usage() {
  cat <<'EOF'
Usage: install.sh [--version <version>] [--bin-dir <dir>] [--dry-run] [--uninstall]

Options:
  --version <version>  Install a specific stable release tag, for example v0.1.0.
  --bin-dir <dir>     Install or uninstall from this directory.
  --dry-run           Print resolved values without installing or uninstalling.
  --uninstall         Remove only the yunxiao binary from the selected directory.
  -h, --help          Show this help.
EOF
}

fail() {
  code="$1"
  shift
  printf 'yunxiao installer: %s\n' "$*" >&2
  exit "$code"
}

info() {
  printf 'yunxiao installer: %s\n' "$*" >&2
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --version)
      [ "$#" -ge 2 ] || fail 2 "--version requires a value"
      version="$2"
      shift 2
      ;;
    --bin-dir)
      [ "$#" -ge 2 ] || fail 2 "--bin-dir requires a value"
      bin_dir="$2"
      shift 2
      ;;
    --dry-run)
      dry_run=1
      shift
      ;;
    --uninstall)
      uninstall=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      fail 2 "unknown argument: $1"
      ;;
  esac
done

resolve_os() {
  detected="${YUNXIAO_INSTALL_OS:-}"
  if [ -z "$detected" ]; then
    detected="$(uname -s 2>/dev/null || true)"
  fi

  case "$detected" in
    linux|Linux) printf 'linux' ;;
    darwin|Darwin) printf 'darwin' ;;
    *) fail 2 "unsupported OS: $detected" ;;
  esac
}

resolve_arch() {
  detected="${YUNXIAO_INSTALL_ARCH:-}"
  if [ -z "$detected" ]; then
    detected="$(uname -m 2>/dev/null || true)"
  fi

  case "$detected" in
    amd64|x86_64) printf 'amd64' ;;
    arm64|aarch64) printf 'arm64' ;;
    *) fail 2 "unsupported architecture: $detected" ;;
  esac
}

resolve_bin_dir() {
  if [ -n "$bin_dir" ]; then
    printf '%s' "$bin_dir"
    return
  fi
  if [ -n "${YUNXIAO_INSTALL_DIR:-}" ]; then
    printf '%s' "$YUNXIAO_INSTALL_DIR"
    return
  fi
  [ -n "${HOME:-}" ] || fail 5 "HOME is not set and no install directory was provided"
  printf '%s/.local/bin' "$HOME"
}

join_url() {
  base="$1"
  path="$2"
  case "$base" in
    */) printf '%s%s' "$base" "$path" ;;
    *) printf '%s/%s' "$base" "$path" ;;
  esac
}

download_file() {
  url="$1"
  dest="$2"
  err="$dest.err"
  rm -f "$err"
  if command -v curl >/dev/null 2>&1; then
    if ! curl -fsSL "$url" -o "$dest" 2>"$err"; then
      detail="$(sed -n '1,3p' "$err" | tr '\n' ' ')"
      fail 3 "failed to download $url${detail:+: $detail}"
    fi
    rm -f "$err"
    return
  fi
  if command -v wget >/dev/null 2>&1; then
    if ! wget -O "$dest" "$url" 2>"$err"; then
      detail="$(sed -n '1,3p' "$err" | tr '\n' ' ')"
      fail 3 "failed to download $url${detail:+: $detail}"
    fi
    rm -f "$err"
    return
  fi
  fail 3 "curl or wget is required"
}

resolve_latest_version() {
  if [ -n "${YUNXIAO_INSTALL_LATEST_VERSION:-}" ]; then
    printf '%s' "$YUNXIAO_INSTALL_LATEST_VERSION"
    return
  fi

  tmp="$(mktemp "${TMPDIR:-/tmp}/yunxiao-latest.XXXXXX")" || fail 5 "failed to create temporary file"
  latest_url="$(join_url "$api_url" "repos/$repo/releases/latest")"
  download_file "$latest_url" "$tmp"
  latest="$(sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$tmp" | sed -n '1p')"
  rm -f "$tmp"
  [ -n "$latest" ] || fail 3 "failed to parse latest stable release from $latest_url; pass --version vX.Y.Z to install a known release"
  printf '%s' "$latest"
}

sha256_file() {
  file="$1"
  if command -v sha256sum >/dev/null 2>&1; then
    hash_output="$(sha256sum "$file")" || fail 4 "failed to calculate SHA-256 for $file"
    printf '%s\n' "$hash_output" | awk '{ print $1 }'
    return
  fi
  if command -v shasum >/dev/null 2>&1; then
    hash_output="$(shasum -a 256 "$file")" || fail 4 "failed to calculate SHA-256 for $file"
    printf '%s\n' "$hash_output" | awk '{ print $1 }'
    return
  fi
  fail 4 "sha256sum or shasum is required"
}

checksum_for_asset() {
  checksums="$1"
  asset="$2"
  awk -v name="$asset" '$2 == name { print $1; exit }' "$checksums"
}

ensure_safe_target() {
  dir="$1"

  [ -n "$dir" ] || fail 7 "install directory is empty"
  [ ! -L "$dir" ] || fail 7 "refusing to uninstall from symlink directory: $dir"
  if [ -d "$dir" ]; then
    physical_dir="$(cd "$dir" 2>/dev/null && pwd -P)" || fail 7 "failed to resolve install directory: $dir"
    [ "$physical_dir" != "/" ] || fail 7 "refusing to uninstall from root directory"
  fi
}

print_path_guidance() {
  dir="$1"
  case ":${PATH:-}:" in
    *":$dir:"*) return ;;
  esac

  info "$dir is not on PATH. Add this line to your shell profile, then restart your shell: export PATH=\"$dir:\$PATH\""
}

os="$(resolve_os)"
arch="$(resolve_arch)"
resolved_bin_dir="$(resolve_bin_dir)"
binary="yunxiao"
target="$resolved_bin_dir/$binary"

if [ "$uninstall" -eq 1 ]; then
  ensure_safe_target "$resolved_bin_dir"
  if [ "$dry_run" -eq 1 ]; then
    printf 'action=uninstall\n'
    printf 'os=%s\n' "$os"
    printf 'arch=%s\n' "$arch"
    printf 'target=%s\n' "$target"
    exit 0
  fi

  if [ -d "$target" ]; then
    fail 7 "refusing to uninstall directory target: $target"
  fi
  if [ ! -e "$target" ]; then
    info "$target is already not installed"
    exit 0
  fi
  rm -f "$target" || fail 7 "failed to remove $target"
  info "removed $target"
  exit 0
fi

if [ -z "$version" ]; then
  version="$(resolve_latest_version)"
fi

asset="yunxiao_${version}_${os}_${arch}.tar.gz"
release_path="releases/download/$version"
archive_url="$(join_url "$download_base_url" "$repo/$release_path/$asset")"
checksums_url="$(join_url "$download_base_url" "$repo/$release_path/checksums.txt")"

if [ "$dry_run" -eq 1 ]; then
  printf 'action=install\n'
  printf 'os=%s\n' "$os"
  printf 'arch=%s\n' "$arch"
  printf 'version=%s\n' "$version"
  printf 'asset=%s\n' "$asset"
  printf 'archive_url=%s\n' "$archive_url"
  printf 'checksums_url=%s\n' "$checksums_url"
  printf 'target=%s\n' "$target"
  exit 0
fi

if [ -e "$resolved_bin_dir" ] && [ ! -d "$resolved_bin_dir" ]; then
  fail 5 "$resolved_bin_dir exists but is not a directory"
fi
mkdir -p "$resolved_bin_dir" || fail 5 "failed to create $resolved_bin_dir"
[ -w "$resolved_bin_dir" ] || fail 5 "$resolved_bin_dir is not writable"
if [ -d "$target" ]; then
  fail 5 "target is a directory: $target"
fi

tmp_dir="$(mktemp -d 2>/dev/null || mktemp -d -t yunxiao-installer)"
tmp_target=""
cleanup() {
  if [ -n "$tmp_target" ]; then
    rm -f "$tmp_target" 2>/dev/null || true
  fi
  rm -rf "$tmp_dir" 2>/dev/null || true
}
trap cleanup EXIT HUP INT TERM

archive="$tmp_dir/$asset"
checksums="$tmp_dir/checksums.txt"
extract_dir="$tmp_dir/extract"
mkdir -p "$extract_dir"

download_file "$archive_url" "$archive"
download_file "$checksums_url" "$checksums"
expected="$(checksum_for_asset "$checksums" "$asset" || true)"
[ -n "$expected" ] || fail 4 "missing checksum entry for $asset in $checksums_url"
actual="$(sha256_file "$archive")"
if [ "$actual" != "$expected" ]; then
  fail 4 "checksum mismatch for $asset; expected $expected from $checksums_url, got $actual from $archive_url; existing install was not overwritten"
fi

tar -xzf "$archive" -C "$extract_dir" "$binary" || fail 1 "archive does not contain expected $binary binary"
[ -f "$extract_dir/$binary" ] || fail 1 "archive does not contain expected $binary binary"
chmod +x "$extract_dir/$binary" || fail 5 "failed to make $binary executable"

tmp_target="$(mktemp "$resolved_bin_dir/.yunxiao.tmp.XXXXXX")" || fail 5 "failed to create staging file in $resolved_bin_dir"
cp "$extract_dir/$binary" "$tmp_target" || fail 5 "failed to stage $target"
chmod 755 "$tmp_target" || fail 5 "failed to set executable permissions on $target"

verify_log="$tmp_dir/verify.log"
if ! "$tmp_target" commands --json >"$verify_log" 2>&1; then
  detail="$(sed -n '1,5p' "$verify_log" | tr '\n' ' ')"
  fail 6 "downloaded yunxiao failed verification before install: $tmp_target commands --json${detail:+; output: $detail}"
fi
mv "$tmp_target" "$target" || fail 5 "failed to install $target"
tmp_target=""
print_path_guidance "$resolved_bin_dir"
info "installed $target"
