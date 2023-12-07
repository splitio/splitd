# vi:ft=yaml
logging:
    level: error
    output: /dev/stdout
    rotationMaxFiles: null
    rotationMaxBytesPerFile: null
sdk:
    apikey: <server-side-apitoken>
    labelsEnabled: true
    streamingEnabled: true
    urls:
        auth: https://auth.split.io
        sdk: https://sdk.split.io/api
        events: https://events.split.io/api
        streaming: https://streaming.split.io/sse
        telemetry: https://telemetry.split.io/api/v1
    featureFlags:
        splitNotificationQueueSize: 5000
        splitRefreshSeconds: 30
        segmentNotificationQueueSize: 5000
        segmentRefreshSeconds: 60
        segmentUpdateWorkers: 20
        segmentUpdateQueueSize: 500
    impressions:
        mode: optimized
        refreshRateSeconds: 1800
        countRefreshRateSeconds: 3600
        queueSize: 8192
        observerSize: 500000
    events:
        refreshRateSeconds: 60
        queueSize: 8192
link:
    type: unix-seqpacket
    address: /var/run/splitd.sock
    maxSimultaneousConns: 32
    readTimeoutMS: 1000
    writeTimeoutMS: 1000
    acceptTimeoutMS: 1000
    serialization: msgpack
    bufferSize: 1024
    protocol: v1
debug:
    profiling:
        enable: false
        host: localhost
        port: 8888

