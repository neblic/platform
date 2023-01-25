#!/bin/sh
set -euo pipefail

# set default configuration settings
OTELCOL_CONFIG_PATH="${OTELCOL_CONFIG_PATH:-/etc/neblic/otelcol/config.yaml}"

# Neblic
export NEBLIC_DATA_PLANE_ENDPOINT="${NEBLIC_DATA_PLANE_ENDPOINT:-0.0.0.0:4317}"
export NEBLIC_CONTROL_PLANE_ENDPOINT="${NEBLIC_CONTROL_PLANE_ENDPOINT:-0.0.0.0:8899}"

# Grafana Loki exporter
LOKI_ENDPOINT="${LOKI_ENDPOINT:-http://loki:3100}"
LOKI_LOG_PUSH_PATH="${LOKI_LOG_PUSH_PATH:-/loki/api/v1/push}"
export LOKI_LOG_PUSH_ENDPOINT=${LOKI_ENDPOINT}${LOKI_LOG_PUSH_PATH}

exec /bin/otelcol --config $OTELCOL_CONFIG_PATH