package sdk

import (
	"errors"
	"strings"

	"github.com/splitio/go-split-commons/v4/storage"
	"github.com/splitio/go-toolkit/v5/logging"
)

// MaxEventLength constant to limit the event size
const MaxEventLength = 32768

var ErrEventTooBig = errors.New("The maximum size allowed for the properties is 32kb. Event not queued")
var ErrEmtpyTrafficType = errors.New("Traffic type cannot be empty")

type Validator struct {
	logger logging.LoggerInterface
	splits storage.SplitStorage
}

func (i *Validator) validateTrafficType(trafficType string) (string, error) {
	if len(trafficType) == 0 {
		return "", ErrEmtpyTrafficType
	}

	toLower := strings.ToLower(trafficType)
	if toLower != trafficType {
		i.logger.Warning("Track: traffic type should be all lowercase - converting string to lowercase")
	}

	if !i.splits.TrafficTypeExists(toLower) {
		i.logger.Warning("Track: traffic type " + toLower + " does not have any corresponding feature flags in this environment, " +
			"make sure you’re tracking your events to a valid traffic type defined in the Split user interface")
	}

	return toLower, nil
}

func (i *Validator) validateTrackProperties(properties map[string]interface{}) (map[string]interface{}, int, error) {
	if len(properties) == 0 {
		return nil, 0, nil
	}

	if len(properties) > 300 {
		i.logger.Warning("Track: Event has more than 300 properties. Some of them will be trimmed when processed")
	}

	processed := make(map[string]interface{})
	size := 1024 // Average event size is ~750 bytes. Using 1kbyte as a starting point.
	for name, value := range properties {
		size += len(name)
		switch value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool, nil:
			processed[name] = value
		case string:
			asStr := value.(string)
			size += len(asStr)
			processed[name] = value
		default:
			i.logger.Warning("Property %s is of invalid type. Setting value to nil")
			processed[name] = nil
		}

		if size > MaxEventLength {
			i.logger.Error("The maximum size allowed for the properties is 32kb. Event not queued")
			return nil, size, ErrEventTooBig
		}
	}
	return processed, size, nil
}
