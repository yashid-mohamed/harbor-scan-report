#!/bin/sh -l

# Harbor instance
export HARBOR_HOST=${1}
export HARBOR_PROTO=${8}
export HARBOR_PORT=${9}

# Harbor Access
export HARBOR_ROBOT=${2}
export HARBOR_TOKEN=${3}

# Harbor image
export IMAGE=${4}
export DIGEST=${14}
export MAX_ALLOWED_SEVERITY=${5}

# GitHub Settings (if set comment will be written)
export GITHUB_TOKEN=${7}
export GITHUB_URL=${6}

# Comment customization
export COMMENT_TITLE=${10}
export COMMENT_MODE=${11}

# Timing options
export TIMEOUT=${12}
export CHECK_INTERVAL=${13}

# Report
export REPORT_SORT_BY=${15}
export REPORT_ONLY_FIXABLE=${16}
export SARIF_OUTPUT_PATH=${17}

# Debug environment
echo "Current directory: $(pwd)"
echo "GITHUB_WORKSPACE: ${GITHUB_WORKSPACE}"
echo "SARIF_OUTPUT_PATH: ${SARIF_OUTPUT_PATH}"

# Run it!
/hsr

# If SARIF_OUTPUT_PATH is provided, verify the file exists and display its location
if [ -n "${SARIF_OUTPUT_PATH}" ]; then
  echo "Checking for SARIF file at ${SARIF_OUTPUT_PATH}"
  if [ -f "${SARIF_OUTPUT_PATH}" ]; then
    echo "✅ SARIF file found at ${SARIF_OUTPUT_PATH}"
    ls -la "${SARIF_OUTPUT_PATH}"
    echo "First 300 bytes of SARIF file:"
    head -c 300 "${SARIF_OUTPUT_PATH}"
    echo ""
  else
    echo "❌ SARIF file not found at ${SARIF_OUTPUT_PATH}"
  fi
fi