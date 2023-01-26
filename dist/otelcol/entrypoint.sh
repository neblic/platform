#!/bin/sh
set -euo pipefail

# set default configuration settings
OTELCOL_CONFIG_PATH="${OTELCOL_CONFIG_PATH:-/etc/neblic/otelcol/config.yaml}"

# Neblic
export NEBLIC_DATA_PLANE_ENDPOINT="${NEBLIC_DATA_PLANE_ENDPOINT:-0.0.0.0:4317}"
export NEBLIC_CONTROL_PLANE_ENDPOINT="${NEBLIC_CONTROL_PLANE_ENDPOINT:-0.0.0.0:8899}"
export NEBLIC_STORAGE_PATH="${NEBLIC_STORAGE_PATH:-/var/lib/otelcol}"

[[ ! -z "$NEBLIC_STORAGE_PATH" ]] &&
    echo "Setting Neblic storage path ${NEBLIC_STORAGE_PATH}" &&
    mkdir -p $NEBLIC_STORAGE_PATH &&
    chown -R otelcol $NEBLIC_STORAGE_PATH

# Grafana Loki exporter
LOKI_ENDPOINT="${LOKI_ENDPOINT:-http://loki:3100}"
LOKI_LOG_PUSH_PATH="${LOKI_LOG_PUSH_PATH:-/loki/api/v1/push}"
export LOKI_LOG_PUSH_ENDPOINT=${LOKI_ENDPOINT}${LOKI_LOG_PUSH_PATH}

echo "Starting otelcol"
exec /usr/bin/sudo -u otelcol /bin/otelcol --config $OTELCOL_CONFIG_PATH