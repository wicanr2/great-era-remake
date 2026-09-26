#!/usr/bin/env bash
# 建立不攜入原版素材的 Windows、macOS 與 Linux AppImage release 候選包。
#
#   tools/package.sh [版本字串] [輸出目錄] [目標清單]
#
# 目標清單以逗號分隔：windows-amd64、darwin-arm64、linux-appimage；可額外指定
# linux-amd64 產生舊式 tar.gz 供 CI／除錯使用。所有編譯、歸檔與 AppImage 組裝都
# 在既有 Docker image 內進行；主機只負責建立唯一暫存掛載點與呼叫 Docker。
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GO_IMAGE="${DSDS_RELEASE_IMAGE:-dsds-go:1.25}"
MAC_IMAGE="${DSDS_MACOS_IMAGE:-u5cht/osxcross:latest}"
APPIMAGE_IMAGE="${DSDS_APPIMAGE_IMAGE:-u5cht/appimage:latest}"
VERSION="${1:-dev}"
# 所有可交付候選包統一集中在 dist-all；推廣影片另置於 dist-all/promo，
# 兩者僅共用交付根目錄，影片絕不被複製進 release archive。
OUT="${2:-$ROOT/dist-all}"
TARGETS="${3:-${DSDS_RELEASE_TARGETS:-windows-amd64,darwin-arm64,linux-appimage}}"
CACHE="$ROOT/workplace/.gocache"
OVERWRITE="${DSDS_RELEASE_OVERWRITE:-0}"

if [[ "$OUT" != /* ]]; then
	OUT="$ROOT/$OUT"
fi
case "$VERSION" in
	""|*[!A-Za-z0-9._-]*) echo "[package] 版本字串含不安全字元：$VERSION" >&2; exit 2 ;;
esac
case "$TARGETS" in
	""|*[!A-Za-z0-9_,-]*) echo "[package] 目標清單含不安全字元：$TARGETS" >&2; exit 2 ;;
esac
case "$OVERWRITE" in 0|1) ;; *) echo "[package] DSDS_RELEASE_OVERWRITE 只能是 0 或 1" >&2; exit 2 ;; esac

need_linux=0
need_windows=0
need_macos=0
need_appimage=0
IFS=, read -r -a target_list <<<"$TARGETS"
for target in "${target_list[@]}"; do
	case "$target" in
		linux-amd64) need_linux=1 ;;
		linux-appimage) need_linux=1; need_appimage=1 ;;
		windows-amd64) need_windows=1 ;;
		darwin-arm64) need_macos=1 ;;
		*) echo "[package] 未知 release 目標：$target" >&2; exit 2 ;;
	esac
done

mkdir -p "$OUT" "$CACHE/pkg" "$CACHE/build"
expected="$(id -u):$(id -g)"
for path in "$OUT" "$CACHE" "$CACHE/pkg" "$CACHE/build"; do
	owner="$(stat -c '%u:%g' "$path")"
	if [[ "$owner" != "$expected" ]]; then
		echo "[package] 拒絕寫入非目前使用者擁有的路徑：$path（$owner，預期 $expected）" >&2
		exit 1
	fi
done

# 只允許刪除此輪 mktemp 建出的 stage。清理仍透過 Docker 完成，避免把主機
# shell 變成 build／封裝工具鏈的一部分。
STAGE="$(mktemp -d "$OUT/.dsds-release-stage.XXXXXX")"
stage_base="$(basename "$STAGE")"
cleanup_stage() {
	docker run --rm --network none --memory 256m --cpus 0.5 --pids-limit 64 \
		-v "$OUT:/dist:rw" -u "$expected" "$GO_IMAGE" bash -eu -c '
			case "$1" in .dsds-release-stage.*) ;; *) exit 2 ;; esac
			rm -rf -- "/dist/$1"
		' bash "$stage_base" >/dev/null 2>&1 || true
}
trap cleanup_stage EXIT

docker run --rm --network none --memory 2g --cpus 2 --pids-limit 512 \
	-v "$ROOT:/work:ro" \
	-v "$CACHE/pkg:/go/pkg:rw" \
	-v "$CACHE/build:/.cache/go-build:rw" \
	-v "$STAGE:/stage:rw" \
	-u "$expected" -w /work \
	-e GOCACHE=/.cache/go-build -e GOFLAGS=-buildvcs=false \
	-e DSDS_VERSION="$VERSION" -e DSDS_NEED_LINUX="$need_linux" \
	-e DSDS_NEED_WINDOWS="$need_windows" \
	"$GO_IMAGE" bash -eu -o pipefail -c '
		mkdir -p /stage/build /stage/common/docs/release /stage/common/docs/licenses /stage/common/docs/promo
		cp README.md /stage/common/README.md
		cp -R translations /stage/common/translations
		cp docs/release/README.md /stage/common/docs/release/README.md
		cp docs/licenses/OFL-1.1.txt docs/licenses/modern-font-atlas.md docs/licenses/portraits-NOTICE.md /stage/common/docs/licenses/
		cp docs/promo/README.md docs/promo/storyboard.md /stage/common/docs/promo/
		cat >/stage/common/RUNNING.md <<EOF
# 執行《大時代的故事》remake

本包只包含 clean-room 引擎、Modern 字型 atlas、語系與說明文件，**不包含原版
遊戲資料、原版美術、原版音樂、原版執行檔或倚天字型**。

請自行準備合法取得的原版資料目錄與字庫，並使用本平台對應的執行檔：

    dsds -game /絕對路徑/遊戲資料 -eten /絕對路徑/eten \\
      -theme modern -resolution high -audio modern \\
      -locale translations/zh-Hant

AppImage 的 AppRun 會先切到隨包的 Resources 目錄，因此同樣的 translations 相對
路徑可直接使用；Windows 與 macOS 請從包內根目錄（或 app bundle 的 Resources）啟動。
沒有合法的 -game 資料目錄時，本包刻意不宣稱可玩。

版本：$DSDS_VERSION
EOF
		if [ "$DSDS_NEED_LINUX" = 1 ]; then
			command -v gcc >/dev/null 2>&1
			CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=gcc /usr/local/go/bin/go build \
				-trimpath -buildvcs=false -ldflags="-s -w" -o /stage/build/dsds-linux-amd64 ./cmd/dsds
		fi
		if [ "$DSDS_NEED_WINDOWS" = 1 ]; then
			CGO_ENABLED=0 GOOS=windows GOARCH=amd64 /usr/local/go/bin/go build \
				-trimpath -buildvcs=false -ldflags="-s -w" -o /stage/build/dsds-windows-amd64.exe ./cmd/dsds
		fi
	'

if [[ "$need_macos" == 1 ]]; then
	docker run --rm --network none --memory 2g --cpus 2 --pids-limit 512 \
		-v "$ROOT:/work:ro" \
		-v "$CACHE/pkg:/go/pkg:rw" \
		-v "$CACHE/build:/go/build:rw" \
		-v "$STAGE:/stage:rw" \
		-u "$expected" -w /work \
		-e GOPATH=/go -e GOMODCACHE=/go/pkg/mod -e GOCACHE=/go/build \
		-e GOFLAGS=-buildvcs=false \
		"$MAC_IMAGE" bash -eu -o pipefail -c '
			command -v oa64-clang >/dev/null 2>&1
			CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 CC=oa64-clang /usr/local/go/bin/go build \
				-trimpath -buildvcs=false -ldflags="-s -w" -o /stage/build/dsds-darwin-arm64 ./cmd/dsds
			file /stage/build/dsds-darwin-arm64 | grep -q "Mach-O 64-bit arm64"
		'
fi

docker run --rm --network none --memory 2g --cpus 2 --pids-limit 512 \
	-v "$ROOT:/work:ro" \
	-v "$CACHE/pkg:/go/pkg:rw" \
	-v "$CACHE/build:/.cache/go-build:rw" \
	-v "$STAGE:/stage:rw" \
	-v "$OUT:/dist:rw" \
	-u "$expected" -w /work \
	-e GOCACHE=/.cache/go-build -e GOFLAGS=-buildvcs=false \
	-e DSDS_VERSION="$VERSION" -e DSDS_TARGETS="$TARGETS" \
	-e DSDS_NEED_LINUX="$need_linux" -e DSDS_NEED_WINDOWS="$need_windows" \
	-e DSDS_NEED_MACOS="$need_macos" -e DSDS_NEED_APPIMAGE="$need_appimage" \
	-e DSDS_OVERWRITE="$OVERWRITE" \
	"$GO_IMAGE" bash -eu -o pipefail -c '
		name="great-era-remake"
		version="$DSDS_VERSION"
		mkdir -p /stage/assemble

		ensure_output() {
			path="$1"
			if [ -e "$path" ] && [ "$DSDS_OVERWRITE" != 1 ]; then
				echo "[package] 拒絕覆蓋既有產物：$path（若確認需要，設 DSDS_RELEASE_OVERWRITE=1）" >&2
				exit 5
			fi
		}
		for target in ${DSDS_TARGETS//,/ }; do
			case "$target" in
				windows-amd64) ensure_output "/dist/$name-windows-amd64-$version.zip" ;;
				darwin-arm64) ensure_output "/dist/$name-macos-arm64-$version.zip" ;;
				linux-appimage) ensure_output "/dist/$name-linux-x86_64-$version.AppImage" ;;
				linux-amd64) ensure_output "/dist/$name-linux-amd64-$version.tar.gz" ;;
			esac
		done
		ensure_output "/dist/SHA256SUMS-$version.txt"

		forbidden() {
			dir="$1"; allow="$2"
			bad="$(find "$dir" -type f \
				\( -iname "*.tpc" -o -iname "*.15" -o -iname "*.rgb" -o -iname "*.mus" -o -iname "*.tim" \
				-o -iname "*.dat" -o -iname "*.glb" -o -iname "*.gtb" -o -iname "*.dt1" -o -iname "*.dt2" \
				-o -iname "*.sav" -o -iname "*.cps" -o -iname "*.bgi" -o -iname "*.ogg" -o -iname "*.wav" \
				-o -iname "*.mid" -o -iname "*.midi" -o -iname "*.exe" \) ! -path "$allow" -print)"
			if [ -n "$bad" ]; then
				echo "[package] release 包含禁止的原版／衍生資產：" >&2
				printf "%s\\n" "$bad" >&2
				exit 4
			fi
			bad="$(find "$dir" -type d \( -name workplace -o -name game \) -print)"
			if [ -n "$bad" ]; then
				echo "[package] release 包含原版資料目錄：" >&2
				printf "%s\\n" "$bad" >&2
				exit 4
			fi
		}

		write_manifest() {
			dir="$1"; target="$2"; runtime="$3"; manifest="$4"
			dirty=no
			git diff --quiet --ignore-submodules -- || dirty=yes
			{
				printf "release_version=%s\\n" "$version"
				printf "target=%s\\n" "$target"
				printf "commit=%s\\n" "$(git rev-parse HEAD 2>/dev/null || printf unknown)"
				printf "source_tree_dirty=%s\\n" "$dirty"
				printf "contains_original_game_data=no\\n"
				printf "contains_original_audio_or_font=no\\n"
				printf "requires_user_supplied_legal_game_data=yes\\n"
				printf "runtime_validation=%s\\n" "$runtime"
				printf "\\n[file_sha256]\\n"
				( cd "$dir" && find . -type f ! -name "$(basename "$manifest")" -print0 | sort -z | xargs -0 sha256sum )
			} >"$manifest"
		}

		copy_common() { cp -R /stage/common/. "$1/"; }

		if [ "$DSDS_NEED_WINDOWS" = 1 ]; then
			dir=/stage/assemble/windows/GreatEraRemake
			mkdir -p "$dir"
			cp /stage/build/dsds-windows-amd64.exe "$dir/dsds.exe"
			copy_common "$dir"
			forbidden "$dir" "$dir/dsds.exe"
			write_manifest "$dir" windows-amd64 "cross-compiled; Windows host smoke pending" "$dir/RELEASE-MANIFEST.txt"
			archive="/dist/$name-windows-amd64-$version.zip"
			/usr/local/go/bin/go run ./tools/release_zip.go "$dir" "$archive"
			/usr/local/go/bin/go run ./tools/release_zip.go verify "$archive" "GreatEraRemake/dsds.exe"
		fi

		if [ "$DSDS_NEED_MACOS" = 1 ]; then
			app=/stage/assemble/macos/GreatEraRemake.app
			mkdir -p "$app/Contents/MacOS" "$app/Contents/Resources"
			cp /stage/build/dsds-darwin-arm64 "$app/Contents/MacOS/dsds"
			copy_common "$app/Contents/Resources"
			cp tools/release/dsds.svg "$app/Contents/Resources/dsds.svg"
			cat >"$app/Contents/Info.plist" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
<key>CFBundleDevelopmentRegion</key><string>zh-Hant</string>
<key>CFBundleExecutable</key><string>dsds</string>
<key>CFBundleIdentifier</key><string>org.great-era-remake.dsds</string>
<key>CFBundleInfoDictionaryVersion</key><string>6.0</string>
<key>CFBundleName</key><string>Great Era Remake</string>
<key>CFBundlePackageType</key><string>APPL</string>
<key>CFBundleShortVersionString</key><string>$version</string>
<key>CFBundleVersion</key><string>$version</string>
</dict></plist>
EOF
			forbidden "$app" /dev/null
			write_manifest "$app" macos-arm64 "cross-compiled; macOS device smoke, signing and notarization pending" "$app/Contents/Resources/RELEASE-MANIFEST.txt"
			archive="/dist/$name-macos-arm64-$version.zip"
			/usr/local/go/bin/go run ./tools/release_zip.go "$app" "$archive"
			/usr/local/go/bin/go run ./tools/release_zip.go verify "$archive" "GreatEraRemake.app/Contents/MacOS/dsds"
		fi

		if [ "$DSDS_NEED_LINUX" = 1 ] && [[ ",$DSDS_TARGETS," == *,linux-amd64,* ]]; then
			dir=/stage/assemble/linux/GreatEraRemake
			mkdir -p "$dir"
			cp /stage/build/dsds-linux-amd64 "$dir/dsds"
			copy_common "$dir"
			forbidden "$dir" /dev/null
			write_manifest "$dir" linux-amd64 "Docker build and archive integrity smoke" "$dir/RELEASE-MANIFEST.txt"
			( cd /stage/assemble/linux && tar --sort=name --mtime="UTC 1970-01-01" --owner=0 --group=0 --numeric-owner -cf - GreatEraRemake | gzip -n >"/dist/$name-linux-amd64-$version.tar.gz" )
		fi

		if [ "$DSDS_NEED_APPIMAGE" = 1 ]; then
			app=/stage/appimage/AppDir
			mkdir -p "$app/usr/bin" "$app/usr/share/great-era" "$app/usr/share/icons/hicolor/scalable/apps"
			cp /stage/build/dsds-linux-amd64 "$app/usr/bin/dsds"
			copy_common "$app/usr/share/great-era"
			cp tools/release/dsds.desktop "$app/dsds.desktop"
			cp tools/release/dsds.svg "$app/dsds.svg"
			cp tools/release/dsds.svg "$app/usr/share/icons/hicolor/scalable/apps/dsds.svg"
			cat >"$app/AppRun" <<EOF
#!/usr/bin/env sh
APPDIR="\$(CDPATH= cd -- "\$(dirname -- "\$0")" && pwd)"
cd "\$APPDIR/usr/share/great-era"
exec "\$APPDIR/usr/bin/dsds" "\$@"
EOF
			chmod 0755 "$app/AppRun" "$app/usr/bin/dsds"
			forbidden "$app" /dev/null
			write_manifest "$app" linux-x86_64-appimage "AppImage extract and payload integrity smoke" "$app/RELEASE-MANIFEST.txt"
		fi
	'

if [[ "$need_appimage" == 1 ]]; then
	docker run --rm --network none --memory 1g --cpus 1.5 --pids-limit 256 \
		-v "$STAGE:/stage:rw" -v "$OUT:/dist:rw" \
		-u "$expected" -w /stage/appimage \
		-e ARCH=x86_64 -e VERSION="$VERSION" \
		"$APPIMAGE_IMAGE" bash -eu -o pipefail -c '
			/opt/appimagetool.d/usr/bin/appimagetool AppDir "/dist/great-era-remake-linux-x86_64-$VERSION.AppImage"
			mkdir -p /stage/appimage-verify
			cd /stage/appimage-verify
			"/dist/great-era-remake-linux-x86_64-$VERSION.AppImage" --appimage-extract >/dev/null
			test -x squashfs-root/usr/bin/dsds
			test -f squashfs-root/RELEASE-MANIFEST.txt
			bad="$(find squashfs-root -type f \( -iname "*.tpc" -o -iname "*.15" -o -iname "*.rgb" -o -iname "*.mus" -o -iname "*.tim" -o -iname "*.dat" -o -iname "*.glb" -o -iname "*.gtb" -o -iname "*.dt1" -o -iname "*.dt2" -o -iname "*.sav" -o -iname "*.cps" -o -iname "*.bgi" -o -iname "*.ogg" -o -iname "*.wav" -o -iname "*.mid" -o -iname "*.midi" -o -iname "*.exe" \) -print)"
			[ -z "$bad" ] || { echo "$bad" >&2; exit 4; }
		'
fi

docker run --rm --network none --memory 512m --cpus 1 --pids-limit 128 \
	-v "$OUT:/dist:rw" -u "$expected" "$GO_IMAGE" bash -eu -o pipefail -c '
		version="$1"; targets="$2"
		outputs=""
		for target in ${targets//,/ }; do
			case "$target" in
				windows-amd64) path="/dist/great-era-remake-windows-amd64-$version.zip" ;;
				darwin-arm64) path="/dist/great-era-remake-macos-arm64-$version.zip" ;;
				linux-appimage) path="/dist/great-era-remake-linux-x86_64-$version.AppImage" ;;
				linux-amd64) path="/dist/great-era-remake-linux-amd64-$version.tar.gz" ;;
			esac
			test -s "$path"
			outputs="$outputs $path"
		done
		# shellcheck disable=SC2086
		sha256sum $outputs >"/dist/SHA256SUMS-$version.txt"
	' bash "$VERSION" "$TARGETS"

for path in "$OUT"/*"$VERSION"* "$OUT/SHA256SUMS-$VERSION.txt"; do
	[[ -e "$path" ]] || continue
	owner="$(stat -c '%u:%g' "$path")"
	if [[ "$owner" != "$expected" ]]; then
		echo "[package] 產物擁有權錯誤：$path（$owner，預期 $expected）" >&2
		exit 1
	fi
done

echo "[package] 完成目標：$TARGETS"
echo "[package] 輸出：$OUT（不含原版遊戲資料／音訊／字型）"
