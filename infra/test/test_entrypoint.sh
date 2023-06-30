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
  
    # Exec entrypoint
    [ -f "./testcfg" ] && rm ./testcfg
    export SPLITD_CFG_OUTPUT="./testcfg" 
    export SPLITD_EXEC="${SCRIPT_DIR}/../../splitd"
    export TPL_FILE="${SCRIPT_DIR}/../../splitd.yaml.tpl"
    conf_json=$(bash "${SCRIPT_DIR}/../entrypoint.sh" -outputConfig | awk '/^Config:/ {print $2}')

    # Validate config output
    assert_eq "\"somexxxxxxx\"" $(echo "$conf_json" | jq '.SDK.Apikey') "invalid apikey"
    assert_eq "\"someAuthURL\"" $(echo "$conf_json" | jq '.SDK.URLs.Auth') "invalid auth url"
    assert_eq "\"someSdkURL\"" $(echo "$conf_json" | jq '.SDK.URLs.SDK') "invalid sdk url"
    assert_eq "\"someEventsURL\"" $(echo "$conf_json" | jq '.SDK.URLs.Events') "invalid events url"
    assert_eq "\"someTelemetryURL\"" $(echo "$conf_json" | jq '.SDK.URLs.Telemetry') "invalid telemetry url"
    assert_eq "\"someStreamingURL\"" $(echo "$conf_json" | jq '.SDK.URLs.Streaming') "invalid streaming url"
    assert_eq "false" $(echo "$conf_json" | jq '.SDK.StreamingEnabled') "streaming should be enabled"
    assert_eq "false" $(echo "$conf_json" | jq '.SDK.LabelsEnabled') "labels should be enabled"
    assert_eq "\"someLinkType\"" $(echo "$conf_json" | jq '.Link.Type') "invalid link type"
    assert_eq "\"someLinkAddress\"" $(echo "$conf_json" | jq '.Link.Address') "invalid link address"
    assert_eq "\"someSerialization\"" $(echo "$conf_json" | jq '.Link.Serialization') "invalid serialization"
    assert_eq "1" $(echo "$conf_json" | jq '.Link.MaxSimultaneousConns') "invalid max simultaneous conns"
    assert_eq "2" $(echo "$conf_json" | jq '.Link.ReadTimeoutMS') "invalid read timeout"
    assert_eq "3" $(echo "$conf_json" | jq '.Link.WriteTimeoutMS') "invalid write timeout"
    assert_eq "4" $(echo "$conf_json" | jq '.Link.AcceptTimeoutMS') "invalid accept timeout"
}


testNoApikeyFails && testNoAddressFails && testAllVars && echo "entrypoint tests success."
