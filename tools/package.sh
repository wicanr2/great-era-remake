#!/usr/bin/env bash
# Isolated package build. All workloads run in bounded containers.
set -euo pipefail
package_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
package_project="${DSDS_PACKAGE_PROJECT_ROOT:-$package_root}"
package_source=/work
package_version="${1:?請提供完整版號，例如 v.1.0.0-20261012}"
package_reference="${DSDS_PACKAGE_REFERENCE_VERSION:-}"
if [ -n "$package_reference" ]; then
  [[ "$package_reference" =~ ^v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}$ ]]
  test -f "$package_project/dist-all/$package_reference/SHA256SUMS.json"
fi
package_kind="${2:-patch}"
case "$package_kind" in patch) package_full=0; package_suffix= ;; full-local) package_full=1; package_suffix=-full ;; *) echo '封包種類須為 patch 或 full-local。' >&2; exit 2 ;; esac
[[ "$package_version" =~ ^v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}$ ]] || { echo '版號格式無效。' >&2; exit 2; }
package_out="$package_project/dist-all/$package_version"
package_channel="${DSDS_PACKAGE_RELEASE_CHANNEL:-candidate}"
case "$package_channel" in candidate|stable) ;; *) echo '發行管道須為 candidate 或 stable。' >&2; exit 2 ;; esac
package_modules="$package_project/workplace/.gocache/pkg"
package_go=logh3-hd-preview:20261011-r1
package_mac=hr-osxcross:1.26.7-15.5-r1
package_appimage=hr-appimage:runtime-recovery-r1
package_user="$(id -u):$(id -g)"
test -d "$package_modules"
test -f "$package_root/tools/release/package.py"
test ! -e "$package_out/$package_kind"
test -d "$package_project/workplace/orig"
test -d "$package_project/workplace/eten"
mkdir -p -- "$package_out"
test "$(stat -c '%u:%g' "$package_out")" = "$package_user"
package_stage="$(mktemp -d "$package_out/.dsds-release-stage.XXXXXX")"
package_stage_name="${package_stage##*/}"
cleanup_package() {
  timeout 30 docker run --rm --network none --memory 256m --cpus 1 --pids-limit 64 \
    --user "$package_user" -v "$package_out:/output:rw" --entrypoint /bin/bash "$package_go" \
    -c 'case "$1" in .dsds-release-stage.*) rm -rf -- "/output/$1" ;; *) exit 2 ;; esac' bash "$package_stage_name"
}
trap cleanup_package EXIT
package_go_id="$(docker image inspect "$package_go" --format '{{.Id}}')"
package_mac_id="$(docker image inspect "$package_mac" --format '{{.Id}}')"
package_appimage_id="$(docker image inspect "$package_appimage" --format '{{.Id}}')"
timeout 360 docker run --rm --network none --memory 4g --cpus 4 --pids-limit 512 \
  --user "$package_user" -v "$package_root:/work:ro" -v "$package_project:/src:ro" -v "$package_modules:/go/pkg:ro" \
  -v "$package_project/workplace/orig:/work/workplace/orig:ro" -v "$package_project/workplace/eten:/work/workplace/eten:ro" \
  -v "$package_stage:/stage:rw" -w "$package_source" -e GOCACHE=/tmp/package-go-cache -e GOPROXY=off \
  -e DSDS_VERSION="$package_version" -e DSDS_GO_IMAGE="$package_go_id" \
  -e DSDS_REFERENCE_VERSION="$package_reference" \
  -e DSDS_PACKAGE_FULL_LOCAL="$package_full" \
  -e DSDS_PACKAGE_RELEASE_CHANNEL="$package_channel" \
  -e DSDS_MAC_IMAGE="$package_mac_id" -e DSDS_APPIMAGE_IMAGE="$package_appimage_id" \
  --entrypoint /bin/bash "$package_go" -c '
    set -euo pipefail
    python3 tools/release/package.py prepare "$DSDS_VERSION"
    Xvfb :93 -screen 0 1280x720x24 -nolisten tcp >/tmp/package-xvfb.log 2>&1 &
    package_xvfb_pid=$!
    trap '\''kill "$package_xvfb_pid" 2>/dev/null || true'\'' EXIT
    export DISPLAY=:93
    go test -count=1 ./... > /stage/full-tests.log 2>&1
    bash tools/check_no_cgo.sh > /stage/no-cgo.log
    if [ -n "$DSDS_REFERENCE_VERSION" ]; then
      CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -trimpath -buildvcs=false \
        -ldflags="-s -w -X main.buildVersion=$DSDS_REFERENCE_VERSION" -o /stage/build/reference-linux ./cmd/dsds
      CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -buildvcs=false \
        -ldflags="-s -w -X main.buildVersion=$DSDS_REFERENCE_VERSION" -o /stage/build/reference-windows ./cmd/dsds
    fi
    CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -trimpath -buildvcs=false \
      -ldflags="-s -w -X main.buildVersion=$DSDS_VERSION" -o /stage/build/dsds-linux ./cmd/dsds
    /stage/build/dsds-linux -version > /stage/linux-version.txt
    CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -buildvcs=false \
      -ldflags="-s -w -X main.buildVersion=$DSDS_VERSION" -o /stage/build/dsds.exe ./cmd/dsds
    cat /stage/full-tests.log
  '
timeout 360 docker run --rm --network none --memory 3g --cpus 3 --pids-limit 512 \
  --user "$package_user" -v "$package_root:/work:ro" -v "$package_modules:/go/pkg:ro" \
  -v "$package_stage:/stage:rw" -w /work -e GOCACHE=/tmp/package-mac-cache -e GOPROXY=off \
  -e DSDS_VERSION="$package_version" -e MACOSX_DEPLOYMENT_TARGET=12.0 \
  -e DSDS_REFERENCE_VERSION="$package_reference" \
  --entrypoint /bin/bash "$package_mac" -c '
    set -euo pipefail
    if [ -n "$DSDS_REFERENCE_VERSION" ]; then
      CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 CC=oa64-clang /usr/local/go/bin/go build \
        -trimpath -buildvcs=false -ldflags="-s -w -X main.buildVersion=$DSDS_REFERENCE_VERSION" \
        -o /stage/build/reference-macos ./cmd/dsds
    fi
    CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 CC=oa64-clang /usr/local/go/bin/go build \
      -trimpath -buildvcs=false -ldflags="-s -w -X main.buildVersion=$DSDS_VERSION" \
      -o /stage/build/dsds-macos ./cmd/dsds
    file /stage/build/dsds-macos
  '
timeout 60 docker run --rm --network none --memory 2g --cpus 2 --pids-limit 128 \
  --user "$package_user" -v "$package_root:/work:ro" -v "$package_stage:/stage:rw" \
  -v "$package_modules:/go/pkg:ro" \
  -v "$package_out:/output:rw" -w /work -e DSDS_PACKAGE_FULL_LOCAL="$package_full" \
  -e DSDS_PACKAGE_RELEASE_CHANNEL="$package_channel" --entrypoint python3 "$package_go" \
  tools/release/package.py assemble "$package_version"
if [ "$package_kind" = full-local ]; then
timeout 90 docker run --rm --network none --memory 2g --cpus 2 --pids-limit 128 \
  --user "$package_user" -v "$package_stage:/stage:rw" -v "$package_out:/output:rw" \
  -e DSDS_VERSION="$package_version" -e DSDS_PACKAGE_KIND="$package_kind" -e DSDS_PACKAGE_SUFFIX="$package_suffix" --entrypoint /bin/bash "$package_appimage" -c '
    set -euo pipefail
    test -x /opt/runtime-x86_64
    mksquashfs /stage/AppDir /stage/payload.squashfs -noappend -comp zstd -b 131072 \
      -all-root -mkfs-time 0 -all-time 0 -processors 2 -no-progress
    package_file="/output/$DSDS_PACKAGE_KIND/great-era-remake-linux-x86_64$DSDS_PACKAGE_SUFFIX-$DSDS_VERSION.AppImage"
    cat /opt/runtime-x86_64 /stage/payload.squashfs > "$package_file"
    chmod 0755 "$package_file"
    sha256sum /opt/runtime-x86_64 > /stage/appimage-runtime.sha256
    mkdir /stage/appimage-extracted
    cd /stage/appimage-extracted
    "$package_file" --appimage-extract >/dev/null
    test -x squashfs-root/AppRun
    test -x squashfs-root/usr/bin/dsds
  '
fi
timeout 60 docker run --rm --network none --memory 2g --cpus 2 --pids-limit 128 \
  --user "$package_user" -v "$package_root:/work:ro" -v "$package_stage:/stage:rw" \
  -v "$package_project:/src:ro" -v "$package_out:/output:rw" -w /work -e DSDS_PACKAGE_FULL_LOCAL="$package_full" \
  -e DSDS_REFERENCE_VERSION="$package_reference" \
  -e DSDS_PACKAGE_RELEASE_CHANNEL="$package_channel" --entrypoint python3 "$package_go" \
  tools/release/package.py verify "$package_version"
echo "封包：$package_out"
