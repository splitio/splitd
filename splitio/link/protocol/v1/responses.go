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

type TreatmentsWithFeaturePayload struct {
	Results map[string]TreatmentPayload `msgpack:"r"`
}

type TreatmentPayload struct {
	Treatment    string             `msgpack:"t"`
	Config       *string            `msgpack:"c,omitempty"`
	ListenerData *ListenerExtraData `msgpack:"l,omitempty"`
}

type TreatmentsPayload struct {
	Results []TreatmentPayload `msgpack:"r"`
}

type TrackPayload struct {
	Success bool `msgpack:"s"`
}

type SplitNamesPayload struct {
	Names []string `msgpack:"n"`
}

type SplitPayload struct {
	Name                string            `msgpack:"n"`
	TrafficType         string            `msgpack:"t"`
	Killed              bool              `msgpack:"k"`
	Treatments          []string          `msgpack:"s"`
	ChangeNumber        int64             `msgpack:"c"`
	Configs             map[string]string `msgpack:"f"`
	DefaultTreatment    string            `msgpack:"d"`
	Sets                []string          `msgpack:"e"`
	ImpressionsDisabled bool              `msgpack:"i"`
}

type SplitsPayload struct {
	Splits []SplitPayload `msgpack:"s"`
}

type ListenerExtraData struct {
	Label        string `msgpack:"l"`
	Timestamp    int64  `msgpack:"m"`
	ChangeNumber int64  `msgpack:"c"`
}

type validPayloadsConstraint interface {
	TreatmentPayload |
		TreatmentsPayload |
		TrackPayload |
		SplitNamesPayload |
		SplitPayload |
		SplitsPayload |
		RegisterPayload |
		TreatmentsWithFeaturePayload
}
