#!/usr/bin/env bash
set -euo pipefail

readonly fixture="testdata/pattern.h264"
readonly scenario="examples/smoke.yml"
readonly binary="./chop"
readonly bind="127.0.0.1:6700"
readonly api="http://${bind}/api/blocks"
readonly bootDeadlineSeconds=10
readonly expectedBlocks=("cam-1" "link-1")

require_fixture() {
    if [[ -f "$fixture" ]]; then return; fi
    echo "smoke: skipping — fixture $fixture not found (see testdata/README.md)" >&2
    exit 0
}

build_artifacts() {
    echo "smoke: building front + binary"
    (cd web && pnpm build >/dev/null)
    mkdir -p internal/dashboard/dist
    cp -r web/dist/* internal/dashboard/dist/
    go build -o "$binary" ./cmd/chop
}

start_chop() {
    "$binary" run --bind "$bind" "$scenario" &
    chopPid=$!
    trap 'kill -TERM "$chopPid" 2>/dev/null || true; wait "$chopPid" 2>/dev/null || true' EXIT
}

wait_for_api() {
    local deadline=$((SECONDS + bootDeadlineSeconds))
    until curl -fsS "$api" >/dev/null 2>&1; do
        if (( SECONDS >= deadline )); then
            echo "smoke: API never came up at $api within ${bootDeadlineSeconds}s" >&2
            exit 1
        fi
        sleep 0.5
    done
}

assert_blocks_listed() {
    local got want
    got="$(curl -fsS "$api" | jq -r '.[].id' | sort | tr '\n' ' ')"
    want="$(printf '%s\n' "${expectedBlocks[@]}" | sort | tr '\n' ' ')"
    if [[ "$got" != "$want" ]]; then
        echo "smoke: blocks mismatch — got: ${got}— want: ${want}" >&2
        exit 1
    fi
}

require_fixture
build_artifacts
start_chop
wait_for_api
assert_blocks_listed
echo "smoke: ok"
