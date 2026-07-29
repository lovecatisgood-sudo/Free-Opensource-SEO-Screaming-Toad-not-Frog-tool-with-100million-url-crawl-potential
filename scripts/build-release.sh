#!/bin/sh
set -eu

repository=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repository"

version=${SEO_AUDITOR_VERSION:-2.0.0-rc.1}
commit=$(git rev-parse HEAD)
source_epoch=${SOURCE_DATE_EPOCH:-$(git show -s --format=%ct HEAD)}
built_at=$(date -u -d "@$source_epoch" +%Y-%m-%dT%H:%M:%SZ)
release_root="$repository/.artifacts/release/$version"

mkdir -p "$release_root"
pnpm --dir web/app build
go test ./...
go vet ./...

ldflags="-s -w -X github.com/seo-auditor/seo-auditor/internal/version.Version=$version -X github.com/seo-auditor/seo-auditor/internal/version.Commit=$commit -X github.com/seo-auditor/seo-auditor/internal/version.BuiltAt=$built_at"

for target in linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64; do
    os=${target%-*}
    arch=${target#*-}
    suffix=""
    if [ "$os" = windows ]; then suffix=".exe"; fi
    directory="$release_root/seo-auditor-$version-$target"
    mkdir -p "$directory"
    for command in seo-auditor seo-auditor-cli seo-auditor-mcp; do
        CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -trimpath -buildvcs=false -ldflags "$ldflags" -o "$directory/$command$suffix" "./cmd/$command"
    done
    cp LICENSE README.md SECURITY.md THIRD_PARTY_NOTICES.md "$directory/"
    cp docs/OPERATIONS.md docs/RULE_CATALOG.md docs/SECURITY_MODEL.md "$directory/"
done

SEO_AUDITOR_VERSION="$version" go run ./cmd/seo-auditor-sbom > "$release_root/sbom.cdx.json"
find "$release_root" -type f ! -name SHA256SUMS -print0 | sort -z | xargs -0 sha256sum | sed "s#$release_root/##" > "$release_root/SHA256SUMS"

echo "release artifacts: $release_root"
echo "signing: not performed; configure CI identity/signing before public publication"
