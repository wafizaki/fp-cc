#!/usr/bin/env bash

# deploy_tenant.sh
# - Reimplements functionality from deploy_tenant.py in pure bash
# - Creates tenant directories and launches filebrowser containers
# - Usage: ./deploy_tenant.sh [--start-port PORT] tenant_a tenant_b ...

set -u

timestamp() { date '+%Y-%m-%d %H:%M:%S'; }

print_usage() {
  echo "Usage: $0 [--start-port PORT] tenant_a tenant_b ..."
}

# Default start port
START_PORT=8000

# Parse optional --start-port
if [ "$#" -eq 0 ]; then
  print_usage
  exit 1
fi

# simple arg parsing
ARGS=()
while [[ $# -gt 0 ]]; do
  case "$1" in
    --start-port)
      shift
      if [[ $# -eq 0 ]]; then
        echo "[ERROR] Missing value for --start-port"
        exit 1
      fi
      START_PORT="$1"
      shift
      ;;
    --help|-h)
      print_usage
      exit 0
      ;;
    --*)
      echo "[ERROR] Unknown option: $1" >&2
      print_usage
      exit 1
      ;;
    *)
      ARGS+=("$1")
      shift
      ;;
  esac
done

if [[ ${#ARGS[@]} -eq 0 ]]; then
  echo "[ERROR] No tenant names provided"
  print_usage
  exit 1
fi

# Check docker
if ! command -v docker >/dev/null 2>&1; then
  echo "[ERROR] docker is not installed or not in PATH."
  exit 1
fi

CURRENT_PORT=${START_PORT}

echo "$(timestamp) === DEPLOYMENT START (LOG SCRAPE ONLY) ==="
echo

for tenant in "${ARGS[@]}"; do
  echo "----------------------------------------------------------------"
  echo "$(timestamp) --- Processing Tenant: ${tenant} ---"

  BASE_PATH="$(pwd)/tenants/${tenant}"
  FILES_PATH="${BASE_PATH}/files"
  CONFIG_PATH="${BASE_PATH}/config"

  # create directories
  if mkdir -p "${FILES_PATH}" "${CONFIG_PATH}"; then
    echo "$(timestamp) Created or verified directories:" 
    echo "  - ${FILES_PATH}"
    echo "  - ${CONFIG_PATH}"
  else
    echo "$(timestamp) [ERROR] Failed to create directories for ${tenant}. Skipping."
    echo
    ((CURRENT_PORT++))
    continue
  fi

  CONTAINER_NAME="files_${tenant}"
  VOLUME_NAME="${tenant}_settings_vol"

  # remove old container if exists
  if docker ps -a --format '{{.Names}}' | grep -q -F "${CONTAINER_NAME}"; then
    echo "$(timestamp) Removing old container ${CONTAINER_NAME}..."
    docker stop "${CONTAINER_NAME}" >/dev/null 2>&1 || true
    docker rm "${CONTAINER_NAME}" >/dev/null 2>&1 || true
  fi

  echo "$(timestamp) Starting ${tenant} on port ${CURRENT_PORT}..."

  # run container
  CONFIG_VOL="${tenant}_config_vol"
  if docker run -d --name "${CONTAINER_NAME}" -p "${CURRENT_PORT}:80" \
      -v "${FILES_PATH}:/srv:rw" \
      -v "${VOLUME_NAME}:/database:rw" \
      -v "${CONFIG_VOL}:/config:rw" \
      # -v "${CONFIG_PATH}:/config:rw" \
      filebrowser/filebrowser >/dev/null; then
    echo "$(timestamp) Container started: ${CONTAINER_NAME}"
  else
    echo "$(timestamp) [ERROR] Failed to start container ${CONTAINER_NAME}."
    echo
    ((CURRENT_PORT++))
    continue
  fi

  echo "$(timestamp) Waiting up to 5s for container logs to contain credentials..."

  # Poll docker logs for a short period to reliably catch early log lines
  PASSWORD=""
  MAX_WAIT=5
  waited=0
  while [ $waited -lt $MAX_WAIT ]; do
    PASS_LINE=$(docker logs "${CONTAINER_NAME}" 2>&1 | grep "generated password" | head -n1 || true)
    if [[ -n "${PASS_LINE}" ]]; then
      PASSWORD=$(echo "${PASS_LINE}" | awk '{print $NF}')
      break
    fi
    sleep 1
    waited=$((waited + 1))
  done

  if [[ -n "${PASSWORD}" ]]; then
    echo "$(timestamp) [SUCCESS] Tenant: ${tenant}"
    echo "  URL: http://localhost:${CURRENT_PORT}"
    echo "  User: admin"
    echo "  Pass: ${PASSWORD}"
  else
    echo "$(timestamp) [WARN] Password not found in logs for ${tenant}."
    echo "  (Reason: volume '${VOLUME_NAME}' might already exist from previous run.)"
    echo "  Action: remove the volume with 'docker volume rm ${VOLUME_NAME}' if you want a fresh password."
  fi

  echo
  ((CURRENT_PORT++))
done

echo "$(timestamp) === DEPLOYMENT FINISHED ==="
