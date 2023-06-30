#!/usr/bin/env bash
#
# This script will generate a splitd.yaml file based on environment variables output.

set -e
set -o xtrace

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)
TPL_FILE="${TPL_FILE:-${SCRIPT_DIR}/splitd.yaml.tpl}"
SPLITD_CFG_OUTPUT="${SPLITD_CFG_OUTPUT:-/etc/splitd.yaml}"
SPLITD_EXEC="${SPLITD_EXEC:-/opt/splitd/splitd}"

# Validate mandatory arguments and initialize the template with those values
[ -z ${SPLITD_APIKEY+x} ] && echo "SPLITD_APIKEY env var is mandatory." && exit 1
[ -z ${SPLITD_LINK_ADDRESS+x} ] && echo "SPLITD_LINK_ADDRESS env var is mandatory." && exit 1
accum=$(yq '.sdk.apikey = env(SPLITD_APIKEY) | .link.address = env(SPLITD_LINK_ADDRESS)' "${TPL_FILE}")

# Generate a new yaml file by substituting values on the template with user-provided env-vars
# @{
# sdk configs
[ ! -z ${SPLITD_AUTH_URL+x} ]		    && accum=$(echo "${accum}" | yq '.sdk.urls.auth = env(SPLITD_AUTH_URL)')
[ ! -z ${SPLITD_SDK_URL+x} ] 		    && accum=$(echo "${accum}" | yq '.sdk.urls.sdk = env(SPLITD_SDK_URL)')
[ ! -z ${SPLITD_EVENTS_URL+x} ] 	    && accum=$(echo "${accum}" | yq '.sdk.urls.events = env(SPLITD_EVENTS_URL)')
[ ! -z ${SPLITD_TELEMETRY_URL+x} ] 	    && accum=$(echo "${accum}" | yq '.sdk.urls.telemetry = env(SPLITD_TELEMETRY_URL)')
[ ! -z ${SPLITD_STREAMING_URL+x} ] 	    && accum=$(echo "${accum}" | yq '.sdk.urls.streaming = env(SPLITD_STREAMING_URL)')
[ ! -z ${SPLITD_STREAMING_ENABLED+x} ]  && accum=$(echo "${accum}" | yq '.sdk.streamingEnabled = env(SPLITD_STREAMING_ENABLED)')
[ ! -z ${SPLITD_LABELS_ENABLED+x} ]     && accum=$(echo "${accum}" | yq '.sdk.labelsEnabled = env(SPLITD_LABELS_ENABLED)')
# link configs
[ ! -z ${SPLITD_LINK_TYPE+x} ] 		        && accum=$(echo "${accum}" | yq '.link.type = env(SPLITD_LINK_TYPE)')
[ ! -z ${SPLITD_LINK_SERIALIZATION+x} ]     && accum=$(echo "${accum}" | yq '.link.serialization = env(SPLITD_LINK_SERIALIZATION)')
[ ! -z ${SPLITD_LINK_MAX_CONNS+x} ]         && accum=$(echo "${accum}" | yq '.link.maxSimultaneousConns = env(SPLITD_LINK_MAX_CONNS)')
[ ! -z ${SPLITD_LINK_READ_TIMEOUT_MS+x} ]   && accum=$(echo "${accum}" | yq '.link.readTimeoutMS = env(SPLITD_LINK_READ_TIMEOUT_MS)')
[ ! -z ${SPLITD_LINK_WRITE_TIMEOUT_MS+x} ]  && accum=$(echo "${accum}" | yq '.link.writeTimeoutMS = env(SPLITD_LINK_WRITE_TIMEOUT_MS)')
[ ! -z ${SPLITD_LINK_ACCEPT_TIMEOUT_MS+x} ] && accum=$(echo "${accum}" | yq '.link.acceptTimeoutMS = env(SPLITD_LINK_ACCEPT_TIMEOUT_MS)')
# @}
# Output final config and start daemon
echo "${accum}" > ${SPLITD_CFG_OUTPUT}
exec env SPLITD_CONF_FILE="${SPLITD_CFG_OUTPUT}" "${SPLITD_EXEC}" $@
