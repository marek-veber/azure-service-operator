#!/bin/bash
# Local test build of the Stolostron ASO controller image.
# Override IMG_REG / IMG_REPO / IMG_TAG on the command line as needed, e.g.:
#   IMG_TAG=v2.19.0-hcpclusters.1 ./stolostron/test-build.sh
set -euo pipefail
cd "$(dirname "$0")"
make docker-build \
  IMG_REG="${IMG_REG:-quay.io}" \
  IMG_REPO="${IMG_REPO:-capz}" \
  IMG_TAG="${IMG_TAG:-test-build}"
