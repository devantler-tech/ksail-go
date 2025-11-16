#!/usr/bin/env bash
set -euo pipefail

error() {
  echo "collect-metrics: $*" >&2
  exit 1
}

summary_file="${GITHUB_STEP_SUMMARY:-}"
[[ -n "$summary_file" ]] || error "GITHUB_STEP_SUMMARY is not set"

touch "$summary_file" 2>/dev/null || error "Unable to write to $summary_file"

start_time="${METRICS_START_TIME:-}"
[[ -n "$start_time" ]] || error "METRICS_START_TIME is required"
[[ "$start_time" =~ ^[0-9]+$ ]] || error "METRICS_START_TIME must be an integer"

if [[ -n "${METRICS_END_TIME:-}" ]]; then
  end_time="$METRICS_END_TIME"
  [[ "$end_time" =~ ^[0-9]+$ ]] || error "METRICS_END_TIME must be an integer when provided"
else
  end_time="$(date +%s)"
fi

duration=$((end_time - start_time))
if [[ $duration -lt 0 ]]; then
  duration=0
fi

cache_raw="${METRICS_CACHE_HIT:-}"
cache_status="n/a"
if [[ -n "$cache_raw" ]]; then
  lowered_cache="$(printf '%s' "$cache_raw" | tr '[:upper:]' '[:lower:]')"
  case "$lowered_cache" in
    true|hit|1|yes)
      cache_status="hit"
      ;;
    false|miss|0|no)
      cache_status="miss"
      ;;
    n/a|na|none)
      cache_status="n/a"
      ;;
    *)
      cache_status="$cache_raw"
      ;;
  esac
fi

artifact_checksum="${METRICS_ARTIFACT_CHECKSUM:-n/a}"
[[ -n "$artifact_checksum" ]] || artifact_checksum="n/a"

metrics_json="{\"durationSeconds\":${duration},\"cacheStatus\":\"${cache_status}\",\"artifactChecksum\":\"${artifact_checksum}\"}"

if [[ -n "${METRICS_JOB_NAME:-}" ]]; then
  metrics_json="{\"job\":\"${METRICS_JOB_NAME}\",\"durationSeconds\":${duration},\"cacheStatus\":\"${cache_status}\",\"artifactChecksum\":\"${artifact_checksum}\"}"
fi

if [[ -n "${METRICS_OUTPUT:-}" ]]; then
  printf 'metrics=%s\n' "$metrics_json" >> "$METRICS_OUTPUT"
fi

{
  echo "### Job Metrics"
  echo "- Duration: ${duration}s"
  echo "- Cache: ${cache_status}"
  echo "- Artifact SHA256: ${artifact_checksum}"
} >>"$summary_file"
