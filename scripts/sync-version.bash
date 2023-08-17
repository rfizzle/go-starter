#!/usr/bin/env bash
ROOT_DIR=$(git rev-parse --show-toplevel)

# Get the version from docker.json
PACKAGE_VERSION=$(grep version < "${ROOT_DIR}/package.json" \
  | head -1 \
  | awk -F: '{ print $2 }' \
  | sed 's/[",]//g' \
  | tr -d '[:space:]')
echo "Updating swagger files to use version: ${PACKAGE_VERSION}"

# Now do the replacement in-place (MacOS/Unix compatible)
REPLACE=' version: .*$'
WITH=" version: '${PACKAGE_VERSION}'"
sed -i.tmp "s#${REPLACE}#${WITH}#g" "${ROOT_DIR}/api/config/template.yaml"
sed -i.tmp "s#${REPLACE}#${WITH}#g" "${ROOT_DIR}/api/swagger.yaml"
rm -f "${ROOT_DIR}/api/config/template.yaml.tmp"
rm -f "${ROOT_DIR}/api/swagger.yaml.tmp"