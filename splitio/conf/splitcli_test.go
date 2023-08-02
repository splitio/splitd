package conf

import (
	"os"
	"testing"

	"github.com/splitio/splitd/splitio/link"
	"github.com/stretchr/testify/assert"
)

func TestCliConfig(t *testing.T) {
	os.Args = []string{
		os.Args[0], // keep program name
		"-log-level=someLevel",
		"-conn-type=someConnType",
		"-conn-address=someAddr",
		"-buffer-size=123",
		"-method=someMethod",
		"-key=someKey",
		"-bucketing-key=someBucketing",
		"-feature=someFeature",
		"-features=someFeature1,someFeature2",
		"-traffic-type=someTrafficType",
		"-event-type=someEventType",
		"-value=0.123",
		`-attributes={"some": "attribute"}`,
	}

	parsed, err := ParseCliArgs()
	assert.Nil(t, err)
	assert.Equal(t, "someLevel", parsed.LogLevel)
	assert.Equal(t, "someConnType", parsed.ConnType)
	assert.Equal(t, "someAddr", parsed.ConnAddr)
	assert.Equal(t, 123, parsed.BufSize)
	assert.Equal(t, "someMethod", parsed.Method)
	assert.Equal(t, "someKey", parsed.Key)
	assert.Equal(t, "someBucketing", parsed.BucketingKey)
	assert.Equal(t, "someFeature", parsed.Feature)
	assert.Equal(t, []string{"someFeature1", "someFeature2"}, parsed.Features)
	assert.Equal(t, "someTrafficType", parsed.TrafficType)
	assert.Equal(t, "someEventType", parsed.EventType)
	assert.Equal(t, 0.123, parsed.EventVal)
	assert.Equal(t, map[string]interface{}{"some": "attribute"}, parsed.Attributes)

	// test bad buffer size
	os.Args = []string{os.Args[0], "-buffer-size=sarasa"}
	_, err = ParseCliArgs()
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "buffer-size")

	// test bad event value
	os.Args = []string{os.Args[0], "-value=sarasa"}
	_, err = ParseCliArgs()
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "event value")

	// test bad attributes
	os.Args = []string{os.Args[0], "-attributes=123"}
	_, err = ParseCliArgs()
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "attributes")

}

func TestLinkOptions(t *testing.T) {
    // test defaults
    os.Args = os.Args[:1]
	parsed, err := ParseCliArgs()
	assert.Nil(t, err)
	lo, err := parsed.LinkOpts()
	assert.Nil(t, err)
	assert.Equal(t, link.DefaultConsumerOptions(), *lo)

    // test bad protocol
    os.Args = []string{os.Args[0], "-protocol=sarasa"}
	parsed, err = ParseCliArgs()
	assert.Nil(t, err)
	lo, err = parsed.LinkOpts()
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "protocol")

    // test bad conn type
    os.Args = []string{os.Args[0], "-conn-type=sarasa"}
	parsed, err = ParseCliArgs()
	assert.Nil(t, err)
	lo, err = parsed.LinkOpts()
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "connection type")

    // test bad serialization
    os.Args = []string{os.Args[0], "-serialization=pinpin"}
	parsed, err = ParseCliArgs()
	assert.Nil(t, err)
	lo, err = parsed.LinkOpts()
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "serialization")

}
