#!/usr/bin/env bash
#
# This script will generate a splitd.yaml file based on environment variables output.

set -e

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
[ ! -z ${SPLITD_AUTH_URL+x} ]		                    && accum=$(echo "${accum}" | yq '.sdk.urls.auth = env(SPLITD_AUTH_URL)')
[ ! -z ${SPLITD_SDK_URL+x} ] 		                    && accum=$(echo "${accum}" | yq '.sdk.urls.sdk = env(SPLITD_SDK_URL)')
[ ! -z ${SPLITD_EVENTS_URL+x} ] 	                    && accum=$(echo "${accum}" | yq '.sdk.urls.events = env(SPLITD_EVENTS_URL)')
[ ! -z ${SPLITD_TELEMETRY_URL+x} ] 	                    && accum=$(echo "${accum}" | yq '.sdk.urls.telemetry = env(SPLITD_TELEMETRY_URL)')
[ ! -z ${SPLITD_STREAMING_URL+x} ] 	                    && accum=$(echo "${accum}" | yq '.sdk.urls.streaming = env(SPLITD_STREAMING_URL)')
[ ! -z ${SPLITD_STREAMING_ENABLED+x} ]                  && accum=$(echo "${accum}" | yq '.sdk.streamingEnabled = env(SPLITD_STREAMING_ENABLED)')
[ ! -z ${SPLITD_LABELS_ENABLED+x} ]                     && accum=$(echo "${accum}" | yq '.sdk.labelsEnabled = env(SPLITD_LABELS_ENABLED)')
[ ! -z ${SPLITD_FEATURE_FLAGS_SPLIT_REFRESH_SECS+x} ]   && accum=$(\
    echo "${accum}" | yq '.sdk.featureFlags.splitRefreshSeconds = env(SPLITD_FEATURE_FLAGS_SPLIT_REFRESH_SECS)')
[ ! -z ${SPLITD_FEATURE_FLAGS_SPLIT_QUEUE_SIZE+x} ]     && accum=$(\
    echo "${accum}" | yq '.sdk.featureFlags.splitNotificationQueueSize = env(SPLITD_FEATURE_FLAGS_SPLIT_QUEUE_SIZE)')
[ ! -z ${SPLITD_FEATURE_FLAGS_SEGMENT_REFRESH_SECS+x} ] && accum=$(\
    echo "${accum}" | yq '.sdk.featureFlags.segmentRefreshSeconds = env(SPLITD_FEATURE_FLAGS_SEGMENT_REFRESH_SECS)')
[ ! -z ${SPLITD_FEATURE_FLAGS_SEGMENT_QUEUE_SIZE+x} ]   && accum=$(\
    echo "${accum}" | yq '.sdk.featureFlags.segmentNotificationQueueSize = env(SPLITD_FEATURE_FLAGS_SEGMENT_QUEUE_SIZE)')
[ ! -z ${SPLITD_FEATURE_FLAGS_SEGMENT_WORKER_COUNT+x} ] && accum=$(\
    echo "${accum}" | yq '.sdk.featureFlags.segmentUpdateWorkers = env(SPLITD_FEATURE_FLAGS_SEGMENT_WORKER_COUNT)')
[ ! -z ${SPLITD_FEATURE_FLAGS_SEGMENT_SYNC_BUFFER+x} ]  && accum=$(\
    echo "${accum}" | yq '.sdk.featureFlags.segmentUpdateQueueSize = env(SPLITD_FEATURE_FLAGS_SEGMENT_SYNC_BUFFER)')
[ ! -z ${SPLITD_IMPRESSIONS_MODE+x} ]                   && accum=$(echo "${accum}" | yq '.sdk.impressions.mode = env(SPLITD_IMPRESSIONS_MODE)')
[ ! -z ${SPLITD_IMPRESSIONS_REFRESH_SECS+x} ]           && accum=$(echo "${accum}" | yq '.sdk.impressions.refreshRateSeconds = env(SPLITD_IMPRESSIONS_REFRESH_SECS)')
[ ! -z ${SPLITD_IMPRESSIONS_QUEUE_SIZE+x} ]             && accum=$(echo "${accum}" | yq '.sdk.impressions.queueSize = env(SPLITD_IMPRESSIONS_QUEUE_SIZE)')
[ ! -z ${SPLITD_IMPRESSIONS_COUNT_REFRESH_SECS+x} ]     && accum=$(\
    echo "${accum}" | yq '.sdk.impressions.countRefreshRateSeconds = env(SPLITD_IMPRESSIONS_COUNT_REFRESH_SECS)')
[ ! -z ${SPLITD_IMPRESSIONS_OBSERVER_SIZE+x} ]          && accum=$(echo "${accum}" | yq '.sdk.impressions.observerSize = env(SPLITD_IMPRESSIONS_OBSERVER_SIZE)')
[ ! -z ${SPLITD_EVENTS_REFRESH_SECS+x} ]                && accum=$(echo "${accum}" | yq '.sdk.events.refreshRateSeconds = env(SPLITD_EVENTS_REFRESH_SECS)')
[ ! -z ${SPLITD_EVENTS_QUEUE_SIZE+x} ]                  && accum=$(echo "${accum}" | yq '.sdk.events.queueSize = env(SPLITD_EVENTS_QUEUE_SIZE)')

# link configs
[ ! -z ${SPLITD_LINK_TYPE+x} ] 		        && accum=$(echo "${accum}" | yq '.link.type = env(SPLITD_LINK_TYPE)')
[ ! -z ${SPLITD_LINK_SERIALIZATION+x} ]     && accum=$(echo "${accum}" | yq '.link.serialization = env(SPLITD_LINK_SERIALIZATION)')
[ ! -z ${SPLITD_LINK_MAX_CONNS+x} ]         && accum=$(echo "${accum}" | yq '.link.maxSimultaneousConns = env(SPLITD_LINK_MAX_CONNS)')
[ ! -z ${SPLITD_LINK_READ_TIMEOUT_MS+x} ]   && accum=$(echo "${accum}" | yq '.link.readTimeoutMS = env(SPLITD_LINK_READ_TIMEOUT_MS)')
[ ! -z ${SPLITD_LINK_WRITE_TIMEOUT_MS+x} ]  && accum=$(echo "${accum}" | yq '.link.writeTimeoutMS = env(SPLITD_LINK_WRITE_TIMEOUT_MS)')
[ ! -z ${SPLITD_LINK_ACCEPT_TIMEOUT_MS+x} ] && accum=$(echo "${accum}" | yq '.link.acceptTimeoutMS = env(SPLITD_LINK_ACCEPT_TIMEOUT_MS)')
[ ! -z ${SPLITD_LINK_BUFFER_SIZE+x} ]       && accum=$(echo "${accum}" | yq '.link.bufferSize = env(SPLITD_LINK_BUFFER_SIZE)')

# logger configs
[ ! -z ${SPLITD_LOG_LEVEL+x} ]  && accum=$(echo "${accum}" | yq '.logging.level = env(SPLITD_LOG_LEVEL)')
[ ! -z ${SPLITD_LOG_OUTPUT+x} ] && accum=$(echo "${accum}" | yq '.logging.output = env(SPLITD_LOG_OUTPUT)')

# profiling configs
[ ! -z ${SPLITD_PROFILING_ENABLE+x} ]  && accum=$(echo "${accum}" | yq '.debug.profiling.enable = env(SPLITD_PROFILING_ENABLE)')
[ ! -z ${SPLITD_PROFILING_HOST+x} ]  && accum=$(echo "${accum}" | yq '.debug.profiling.host = env(SPLITD_PROFILING_HOST)')
[ ! -z ${SPLITD_PROFILING_PORT+x} ]  && accum=$(echo "${accum}" | yq '.debug.profiling.port = env(SPLITD_PROFILING_PORT)')
# @}

# Ensure that the socket-file is read-writable by anyone
umask 000

# Output final config and start daemon
echo "${accum}" > ${SPLITD_CFG_OUTPUT}
exec env SPLITD_CONF_FILE="${SPLITD_CFG_OUTPUT}" "${SPLITD_EXEC}" $@
