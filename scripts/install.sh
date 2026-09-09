#!/bin/sh
set -eu

repository="ishiguro-junya/arca"
asset="arca_darwin_arm64.tar.gz"

if [ "$(uname -s)" != "Darwin" ] || [ "$(uname -m)" != "arm64" ]; then
  echo "ArcaはmacOS Apple Siliconのみ対応しています。" >&2
  exit 1
fi

install_dir="${ARCA_INSTALL_DIR:-${HOME}/.local/bin}"
release_url="$(curl -fsSL -o /dev/null -w '%{url_effective}' "https://github.com/${repository}/releases/latest")"
version="${release_url##*/}"

case "${version}" in
  v[0-9]*.[0-9]*.[0-9]*) ;;
  *)
    echo "最新Releaseのバージョンを取得できませんでした。" >&2
    exit 1
    ;;
esac

work_dir="$(mktemp -d)"
trap 'rm -rf "${work_dir}"' EXIT HUP INT TERM

base_url="https://github.com/${repository}/releases/download/${version}"
curl -fsSL "${base_url}/${asset}" -o "${work_dir}/${asset}"
curl -fsSL "${base_url}/checksums.txt" -o "${work_dir}/checksums.txt"

checksum_line="$(grep "  ${asset}$" "${work_dir}/checksums.txt")"
if [ -z "${checksum_line}" ]; then
  echo "${asset}のチェックサムが見つかりません。" >&2
  exit 1
fi
printf '%s\n' "${checksum_line}" | (cd "${work_dir}" && shasum -a 256 -c -)

tar -xzf "${work_dir}/${asset}" -C "${work_dir}"
mkdir -p "${install_dir}"
install -m 0755 "${work_dir}/arca" "${install_dir}/arca"

echo "Arca ${version}を${install_dir}/arcaへインストールしました。"
