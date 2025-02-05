package controllers

type SplitViewDTO struct {
	Name                string            `json:"name"`
	TrafficType         string            `json:"trafficType"`
	Killed              bool              `json:"killed"`
	Treatments          []string          `json:"treatments"`
	ChangeNumber        int64             `json:"changeNumber"`
	Configs             map[string]string `json:"configs"`
	DefaultTreatment    string            `json:"defaultTreatment"`
	Sets                []string          `json:"sets"`
	ImpressionsDisabled bool              `json:"impressionsDisabled"`
}
