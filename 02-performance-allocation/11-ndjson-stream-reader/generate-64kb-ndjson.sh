#!/usr/bin/env bash

set -euo pipefail

target_bytes=65536
output_file="${1:-input-64kb.ndjson}"

for payload_len in $(seq 0 4096); do
  payload=$(printf '%*s' "$payload_len" '' | tr ' ' x)
  line=$(printf '{"id":1,"message":"%s"}\n' "$payload")
  line_len=${#line}

  if (( target_bytes % line_len == 0 )); then
    line_count=$((target_bytes / line_len))
    : > "$output_file"

    for _ in $(seq 1 "$line_count"); do
      printf '%s' "$line" >> "$output_file"
    done

    printf 'Wrote %s bytes to %s using %s lines of %s bytes each\n' "$target_bytes" "$output_file" "$line_count" "$line_len"
    exit 0
  fi
done

printf 'Could not construct an NDJSON line length that evenly divides %s bytes\n' "$target_bytes" >&2
exit 1