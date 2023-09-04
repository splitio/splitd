#!/usr/bin/env bash

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)

source "${SCRIPT_DIR}/assert.sh"

function testNoApikeyFails {
    unset SPLITD_APIKEY
    export SPLITD_LINK_ADDRESS="some"
    bash "${SCRIPT_DIR}/../entrypoint.sh" &> /dev/null
    assert_eq 1 $? "should fail with a missing apikey"
}

function testNoAddressFails {
    export SPLITD_APIKEY="some"
    unset SPLITD_LINK_ADDRESS
    bash "${SCRIPT_DIR}/../entrypoint.sh" &> /dev/null
    assert_eq 1 $? "should fail with missing link address"
}

function testAllVars {
    # Set variables
    export SPLITD_APIKEY="someApikey"
    export SPLITD_AUTH_URL="someAuthURL"
    export SPLITD_SDK_URL="someSdkURL"
    export SPLITD_EVENTS_URL="someEventsURL"
    export SPLITD_TELEMETRY_URL="someTelemetryURL"
    export SPLITD_STREAMING_URL="someStreamingURL"
    export SPLITD_STREAMING_ENABLED="false"
    export SPLITD_LABELS_ENABLED="false"
    export SPLITD_LINK_TYPE="someLinkType"
    export SPLITD_LINK_ADDRESS="someLinkAddress"
    export SPLITD_LINK_SERIALIZATION="someSerialization"
    export SPLITD_LINK_MAX_CONNS=1
    export SPLITD_LINK_READ_TIMEOUT_MS=2
    export SPLITD_LINK_WRITE_TIMEOUT_MS=3
    export SPLITD_LINK_ACCEPT_TIMEOUT_MS=4
    export SPLITD_LOG_LEVEL="WARNING"
    export SPLITD_LOG_OUTPUT="/dev/stderr"

    export SPLITD_FEATURE_FLAGS_SPLIT_REFRESH_SECS="1"
    export SPLITD_FEATURE_FLAGS_SPLIT_QUEUE_SIZE="2"
    export SPLITD_FEATURE_FLAGS_SEGMENT_REFRESH_SECS="3"
    export SPLITD_FEATURE_FLAGS_SEGMENT_QUEUE_SIZE="4"
    export SPLITD_FEATURE_FLAGS_SEGMENT_WORKER_COUNT="5"
    export SPLITD_FEATURE_FLAGS_SEGMENT_SYNC_BUFFER="6"
    export SPLITD_IMPRESSIONS_MODE="anotherMode"
    export SPLITD_IMPRESSIONS_REFRESH_SECS="7"
    export SPLITD_IMPRESSIONS_QUEUE_SIZE="8"
    export SPLITD_IMPRESSIONS_COUNT_REFRESH_SECS="9"
    export SPLITD_IMPRESSIONS_OBSERVER_SIZE="10"
    export SPLITD_EVENTS_REFRESH_SECS="11"
    export SPLITD_EVENTS_QUEUE_SIZE="12"


    # Exec entrypoint
    [ -f "./testcfg" ] && rm ./testcfg
    export SPLITD_CFG_OUTPUT="./testcfg"
    export SPLITD_EXEC="${SCRIPT_DIR}/../../splitd"
    export TPL_FILE="${SCRIPT_DIR}/../../splitd.yaml.tpl"
    conf_json=$(bash "${SCRIPT_DIR}/../entrypoint.sh" -outputConfig | awk '/^Config:/ {print $2}')

    # Validate config output
    assert_eq '"somexxxxxxx"' $(echo "$conf_json" | jq '.SDK.Apikey') "incorrect apikey"
    assert_eq '"someAuthURL"' $(echo "$conf_json" | jq '.SDK.URLs.Auth') "incorrect auth url"
    assert_eq '"someSdkURL"' $(echo "$conf_json" | jq '.SDK.URLs.SDK') "incorrect sdk url"
    assert_eq '"someEventsURL"' $(echo "$conf_json" | jq '.SDK.URLs.Events') "incorrect events url"
    assert_eq '"someTelemetryURL"' $(echo "$conf_json" | jq '.SDK.URLs.Telemetry') "incorrect telemetry url"
    assert_eq '"someStreamingURL"' $(echo "$conf_json" | jq '.SDK.URLs.Streaming') "incorrect streaming url"
    assert_eq "false" $(echo "$conf_json" | jq '.SDK.StreamingEnabled') "streaming should be enabled"
    assert_eq "false" $(echo "$conf_json" | jq '.SDK.LabelsEnabled') "labels should be enabled"
    assert_eq "1" $(echo "$conf_json" | jq '.SDK.FeatureFlags.SplitRefreshRateSeconds') "incorrect split refresh rate"
    assert_eq "2" $(echo "$conf_json" | jq '.SDK.FeatureFlags.SplitNotificationQueueSize') "incorrect split queue size"
    assert_eq "3" $(echo "$conf_json" | jq '.SDK.FeatureFlags.SegmentRefreshRateSeconds') "incorrect segment refresh rate"
    assert_eq "4" $(echo "$conf_json" | jq '.SDK.FeatureFlags.SegmentNotificationQueueSize') "incorrect segment queue size"
    assert_eq "5" $(echo "$conf_json" | jq '.SDK.FeatureFlags.SegmentWorkerCount') "incorrect segment worker count"
    assert_eq "6" $(echo "$conf_json" | jq '.SDK.FeatureFlags.SegmentWorkerBufferSize') "incorrect segment sync buffer"
    assert_eq '"anotherMode"' $(echo "$conf_json" | jq '.SDK.Impressions.Mode') "incorrect impressions mode"
    assert_eq "7" $(echo "$conf_json" | jq '.SDK.Impressions.RefreshRateSeconds') "incorrect impressions refresh rate"
    assert_eq "8" $(echo "$conf_json" | jq '.SDK.Impressions.QueueSize') "incorrect impressions impressions queue size"
    assert_eq "9" $(echo "$conf_json" | jq '.SDK.Impressions.CountRefreshRateSeconds') "incorrect impressions count refresh rate"
    assert_eq "10" $(echo "$conf_json" | jq '.SDK.Impressions.ObserverSize') "incorrect impressions observer size"
    assert_eq "11" $(echo "$conf_json" | jq '.SDK.Events.RefreshRateSeconds') "incorrect events refresh rate"
    assert_eq "12" $(echo "$conf_json" | jq '.SDK.Events.QueueSize') "incorrect events queue size"

    # ---

    assert_eq '"someLinkType"' $(echo "$conf_json" | jq '.Link.Type') "incorrect link type"
    assert_eq '"someLinkAddress"' $(echo "$conf_json" | jq '.Link.Address') "incorrect link address"
    assert_eq '"someSerialization"' $(echo "$conf_json" | jq '.Link.Serialization') "incorrect serialization"
    assert_eq "1" $(echo "$conf_json" | jq '.Link.MaxSimultaneousConns') "incorrect max simultaneous conns"
    assert_eq "2" $(echo "$conf_json" | jq '.Link.ReadTimeoutMS') "incorrect read timeout"
    assert_eq "3" $(echo "$conf_json" | jq '.Link.WriteTimeoutMS') "incorrect write timeout"
    assert_eq "4" $(echo "$conf_json" | jq '.Link.AcceptTimeoutMS') "incorrect accept timeout"

    # ---

    assert_eq '"WARNING"' $(echo "$conf_json" | jq '.Logger.Level') "incorrect log level"
    assert_eq '"/dev/stderr"' $(echo "$conf_json" | jq '.Logger.Output') "incorrect log output"

}

testNoApikeyFails && testNoAddressFails && testAllVars && echo "entrypoint tests success."
