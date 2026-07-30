#!/bin/sh
set -eu

release_directory=${1:-.}
repository=${2:-lovecatisgood-sudo/Free-Opensource-SEO-Screaming-Toad-not-Frog-tool-with-100million-url-crawl-potential}

cd "$release_directory"
test -f SHA256SUMS
sha256sum -c SHA256SUMS

for archive in *.tar.gz *.zip; do
    [ -e "$archive" ] || continue
    if command -v gh >/dev/null 2>&1; then
        gh attestation verify "$archive" -R "$repository"
    fi
done

echo "release checksums verified"
if ! command -v gh >/dev/null 2>&1; then
    echo "GitHub CLI not found; provenance verification was skipped"
fi
