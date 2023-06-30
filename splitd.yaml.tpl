# vim:ft=yaml
logging:
  level: "ERROR"
sdk:
  apikey: "YOUR_API_KEY"
  urls:
    auth: "https://auth.split.io"
    sdk: "https://sdk.split.io/api"
    events: "https://events.split.io/api"
    streaming: "https://streaming.split.io/sse"
    telemetry: "https://telemetry.split.io/api/v1"

link:
  type: "unix-seqpacket"
  address: "/var/run/splitd.sock"
  serialization: "msgpack"



