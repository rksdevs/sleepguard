#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../web/pwa"
npm ci
npm run build
echo "PWA built to web/pwa/dist — configure nginx root to this path"
