package types

type ClientConfig struct {
	ReturnImpressionData bool
	Metadata             ClientMetadata
}

type ClientMetadata struct {
	ID         string
	SdkVersion string
}
