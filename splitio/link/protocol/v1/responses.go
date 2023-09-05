package v1

type Result byte

const (
	ResultOk            Result = 0x01
	ResultInternalError Result = 0x10
)

type ResponseWrapper[T validPayloadsConstraint] struct {
	Status  Result `msgpack:"s"`
	Payload T      `msgpack:"p"`
}

type RegisterPayload struct{}

type TreatmentPayload struct {
	Treatment    string             `msgpack:"t"`
	ListenerData *ListenerExtraData `msgpack:"l,omitempty"`
}

type TreatmentsPayload struct {
	Results []TreatmentPayload `msgpack:"r"`
}

type TreatmentWithConfigPayload struct {
	Treatment    string             `msgpack:"t"`
	Config       string             `msgpack:"c"`
	ListenerData *ListenerExtraData `msgpack:"l,omitempty"`
}

type TreatmentsWithConfigPayload struct {
	Results []TreatmentWithConfigPayload `msgpack:"r"`
}

type TrackPayload struct {
    Success bool `msgpack:"s"`
}

type ListenerExtraData struct {
	Label        string `msgpack:"l"`
	Timestamp    int64  `msgpack:"m"`
	ChangeNumber int64  `msgpack:"c"`
}

type validPayloadsConstraint interface {
	TreatmentPayload |
		TreatmentsPayload |
		TreatmentWithConfigPayload |
		TreatmentsWithConfigPayload |
		TrackPayload |
		RegisterPayload
}
