#!/usr/bin/env bash

set -euo pipefail
export LC_ALL=C

maximum_bytes="${MAX_PORTABLE_PATH_BYTES:-180}"
invalid=0

while IFS= read -r -d '' path; do
  if (( ${#path} > maximum_bytes )); then
    echo "path exceeds ${maximum_bytes} bytes: ${path}" >&2
    invalid=1
  fi

  IFS="/" read -r -a components <<< "${path}"
  for component in "${components[@]}"; do
    if [[ "${component}" =~ [\<\>:\"\\\|\?\*] ]]; then
      echo "path contains a Windows-forbidden character: ${path}" >&2
      invalid=1
      break
    fi
    if [[ "${component}" == *"." || "${component}" == *" " ]]; then
      echo "path component ends with a dot or space: ${path}" >&2
      invalid=1
      break
    fi

    stem="${component%%.*}"
    case "${stem^^}" in
      CON|PRN|AUX|NUL|COM[1-9]|LPT[1-9])
        echo "path contains a Windows-reserved name: ${path}" >&2
        invalid=1
        break
        ;;
    esac
  done
done < <(git ls-files -z)

if (( invalid != 0 )); then
  exit 1
fi

echo "Repository paths are portable (${maximum_bytes}-byte relative-path limit)."
