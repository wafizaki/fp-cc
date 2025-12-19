#!/bin/bash

# Improved remove_tenant.sh
# - Fixes argument check bug
# - Adds timestamped, colored and more readable logs
# - Detects whether sudo is required
# - Gracefully handles missing containers/volumes

set -u

# Detect if sudo is needed
if [ "$(id -u)" -ne 0 ]; then
    SUDO="sudo"
else
    SUDO=""
fi

timestamp() {
    date '+%Y-%m-%d %H:%M:%S'
}

echo "$(timestamp) --- REMOVING ALL TENANT CONTAINERS AND VOLUMES ---"
echo

if [ $# -eq 0 ]; then
    echo "[ERROR] Please provide one or more tenant names."
    echo "Usage: $0 tenant_a tenant_b"
    exit 1
fi

for name in "$@"; do
    echo "----------------------------------------------------------------"
    echo "$(timestamp) --- Deleting tenant: $name ---"

    CONTAINER_NAME="files_$name"
    VOLUME_NAME="${name}_settings_vol"

    # Remove container if exists
    if $SUDO docker ps -a --format '{{.Names}}' | grep -q -F "${CONTAINER_NAME}"; then
        echo "$(timestamp) Stopping container ${CONTAINER_NAME}..."
        $SUDO docker stop "${CONTAINER_NAME}" >/dev/null 2>&1 || true
        echo "$(timestamp) Removing container ${CONTAINER_NAME}..."
        $SUDO docker rm "${CONTAINER_NAME}" >/dev/null 2>&1 || true
        echo "$(timestamp) Container ${CONTAINER_NAME} deleted."
    else
        echo "$(timestamp) Container ${CONTAINER_NAME} not found. Skipping container removal."
    fi

    # Remove volume if exists
    if $SUDO docker volume ls -q | grep -q -F "${VOLUME_NAME}"; then
        echo "$(timestamp) Removing volume ${VOLUME_NAME}..."
        $SUDO docker volume rm "${VOLUME_NAME}" >/dev/null 2>&1 || true
        echo "$(timestamp) Volume ${VOLUME_NAME} deleted."
    else
        echo "$(timestamp) Volume ${VOLUME_NAME} not found. Skipping volume removal."
    fi

    # Remove tenant directory under the repository's tenants/ directory (safe check)
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    TENANTS_ROOT="${SCRIPT_DIR}/tenants"
    TENANT_DIR="${TENANTS_ROOT}/${name}"

    if [ -d "${TENANT_DIR}" ]; then
        # Resolve to real paths to avoid accidental outside deletions
        TENANT_DIR_REAL=$(realpath "${TENANT_DIR}") || TENANT_DIR_REAL="${TENANT_DIR}"
        TENANTS_ROOT_REAL=$(realpath "${TENANTS_ROOT}") || TENANTS_ROOT_REAL="${TENANTS_ROOT}"

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
