#!/usr/bin/env bash
set -u

timestamp() { date '+%Y-%m-%d %H:%M:%S'; }

# Only use sudo if needed AND available
SUDO=""
if [ "$(id -u)" -ne 0 ] && command -v sudo >/dev/null 2>&1; then
  SUDO="sudo"
fi

echo "$(timestamp) --- REMOVING TENANT CONTAINERS AND VOLUMES ---"
echo

if [ $# -eq 0 ]; then
  echo "[ERROR] Please provide one or more tenant names."
  echo "Usage: $0 tenant_a tenant_b"
  exit 1
fi

for name in "$@"; do
  echo "----------------------------------------------------------------"
  echo "$(timestamp) --- Deleting tenant: $name ---"

  CONTAINER_NAME="files_${name}"

  FILES_VOL="${name}_files_vol"
  CONFIG_VOL="${name}_config_vol"
  DB_VOL="${name}_settings_vol"

  # Remove container
  if $SUDO docker ps -a --format '{{.Names}}' | grep -q -F "${CONTAINER_NAME}"; then
    echo "$(timestamp) Stopping container ${CONTAINER_NAME}..."
    $SUDO docker stop "${CONTAINER_NAME}" >/dev/null 2>&1 || true
    echo "$(timestamp) Removing container ${CONTAINER_NAME}..."
    $SUDO docker rm "${CONTAINER_NAME}" >/dev/null 2>&1 || true
    echo "$(timestamp) Container ${CONTAINER_NAME} deleted."
  else
    echo "$(timestamp) Container ${CONTAINER_NAME} not found. Skipping container removal."
  fi

  # Remove volumes (all 3)
  for v in "${FILES_VOL}" "${CONFIG_VOL}" "${DB_VOL}"; do
    if $SUDO docker volume ls -q | grep -q -F "${v}"; then
      echo "$(timestamp) Removing volume ${v}..."
      $SUDO docker volume rm "${v}" >/dev/null 2>&1 || true
      echo "$(timestamp) Volume ${v} deleted."
    else
      echo "$(timestamp) Volume ${v} not found. Skipping."
    fi
  done

  # Optional: remove tenant dir (kalau kamu masih bikin folder tenants/<name>/... )
  SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  TENANTS_ROOT="${SCRIPT_DIR}/tenants"
  TENANT_DIR="${TENANTS_ROOT}/${name}"

  if [ -d "${TENANT_DIR}" ]; then
    TENANT_DIR_REAL=$(realpath "${TENANT_DIR}" 2>/dev/null || echo "${TENANT_DIR}")
    TENANTS_ROOT_REAL=$(realpath "${TENANTS_ROOT}" 2>/dev/null || echo "${TENANTS_ROOT}")

    if [[ "${TENANT_DIR_REAL}" == "${TENANTS_ROOT_REAL}"* ]]; then
      echo "$(timestamp) Removing tenant directory ${TENANT_DIR_REAL}..."
      rm -rf "${TENANT_DIR_REAL}" || echo "$(timestamp) [WARN] Failed to remove ${TENANT_DIR_REAL}"
      echo "$(timestamp) Tenant directory removed."
    else
      echo "$(timestamp) [WARN] Refusing to remove ${TENANT_DIR_REAL}: not under ${TENANTS_ROOT_REAL}."
    fi
  else
    echo "$(timestamp) Tenant directory ${TENANT_DIR} not found. Skipping directory removal."
  fi

  echo
done

echo "$(timestamp) --- FINISHED ---"
