package v1

type Result byte

const (
	ResultOk            Result = 0x01
	ResultInternalError Result = 0x10
)

type ResponseWrapper[T validPayloadsConstraint] struct {
	Status  Result
	Payload T
}

type RegisterPayload struct{}

type TreatmentPayload struct {
	Treatment string
}

type TreatmentsPayload struct {
	Treatments map[string]string
}

type TreatmentWithConfigPayload struct {
	Treatment string
	Config    string
}

type TreatmentsWithConfigPayload struct {
	Results map[string]struct {
		Treatment string
		Config    string
	}
}

type validPayloadsConstraint interface {
	TreatmentPayload |
		TreatmentsPayload |
		TreatmentWithConfigPayload |
		TreatmentsWithConfigPayload |
		RegisterPayload
}
