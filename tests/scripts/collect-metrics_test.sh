#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")/../.." && pwd)"
SCRIPT_PATH="$ROOT_DIR/.github/scripts/collect-metrics.sh"
SUMMARY_FILE="$(mktemp)"
OUTPUT_FILE="$(mktemp)"
trap 'rm -f "$SUMMARY_FILE" "$OUTPUT_FILE"' EXIT

export GITHUB_STEP_SUMMARY="$SUMMARY_FILE"
export METRICS_START_TIME=1700000000
export METRICS_END_TIME=1700000065
export METRICS_CACHE_HIT=true
export METRICS_ARTIFACT_CHECKSUM="sha256-expected"
export METRICS_OUTPUT="$OUTPUT_FILE"
export METRICS_JOB_NAME="unit-test"

bash "$SCRIPT_PATH"
grep -q "^### Job Metrics$" "$SUMMARY_FILE"
grep -q "^- Duration: 65s$" "$SUMMARY_FILE"
grep -q "^- Cache: hit$" "$SUMMARY_FILE"
grep -q "^- Artifact SHA256: sha256-expected$" "$SUMMARY_FILE"

grep -q '^metrics={"job":"unit-test","durationSeconds":65,"cacheStatus":"hit","artifactChecksum":"sha256-expected"}$' "$OUTPUT_FILE"
