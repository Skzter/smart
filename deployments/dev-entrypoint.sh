#!/usr/bin/env sh

set -euo pipefail

MODULE_NAME="gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart"

# Ensure the same defaults that Taskfile uses when running Air locally.
export VERSION_VAR="${VERSION_VAR:-${MODULE_NAME}/internal/build.Version}"
export OPENAI_KEY_VAR="${OPENAI_KEY_VAR:-${MODULE_NAME}/internal/build.OpenAIKey}"
export VERSION="${VERSION:-dev}"
export APP_CMD="${APP_CMD:-./cmd/suproxy}"
export PATH="/app/bin:/go/bin:${PATH}"

DOPPLER_PROJECT="${DOPPLER_PROJECT:-htwk-projekt}"
DOPPLER_CONFIG="${DOPPLER_CONFIG:-dev}"

# Avoid git safety issues when the repo is mounted from host with different UID.
git config --global --add safe.directory /app || true

start_air() {
  # Mirror the Taskfile behaviour so Air picks up the existing config.
  exec go tool air -c .air.toml
}

# Prefer running through Doppler so secrets are injected just like local Task runs.
if [ -n "${DOPPLER_TOKEN:-}" ]; then
  exec doppler run --project "${DOPPLER_PROJECT}" --config "${DOPPLER_CONFIG}" -- go tool air -c .air.toml
fi

echo "DOPPLER_TOKEN not provided. Starting Air without Doppler; ensure required env vars are set." >&2
start_air

