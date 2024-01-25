package v1

import (
	"testing"

	"github.com/splitio/go-split-commons/v5/dtos"
	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio/common/lang"
	v1 "github.com/splitio/splitd/splitio/link/protocol/v1"
	proto1Mocks "github.com/splitio/splitd/splitio/link/protocol/v1/mocks"
	serializerMocks "github.com/splitio/splitd/splitio/link/serializer/mocks"
	transferMocks "github.com/splitio/splitd/splitio/link/transfer/mocks"
	"github.com/splitio/splitd/splitio/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClientGetTreatmentNoImpression(t *testing.T) {

	logger := logging.NewLogger(nil)

	rawConnMock := &transferMocks.RawConnMock{}
	rawConnMock.On("SendMessage", []byte("registrationMessage")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte("registrationSuccess"), nil).Once()
	rawConnMock.On("SendMessage", []byte("treatmentMessage")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte("treatmentResult"), nil).Once()

	serializerMock := &serializerMocks.SerializerMock{}
	serializerMock.On("Serialize", proto1Mocks.NewRegisterRPC("some", false)).Return([]byte("registrationMessage"), nil).Once()
	serializerMock.On("Parse", []byte("registrationSuccess"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.ResponseWrapper[v1.RegisterPayload]) = v1.ResponseWrapper[v1.RegisterPayload]{Status: v1.ResultOk}
	}).Once()

	serializerMock.On("Serialize", proto1Mocks.NewTreatmentRPC("key1", "buck1", "feat1", map[string]interface{}{"a": 1}, false)).
		Return([]byte("treatmentMessage"), nil).Once()
	serializerMock.On("Parse", []byte("treatmentResult"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.ResponseWrapper[v1.TreatmentPayload]) = v1.ResponseWrapper[v1.TreatmentPayload]{
			Status:  v1.ResultOk,
			Payload: v1.TreatmentPayload{Treatment: "on"},
		}
	}).Once()
	client, err := New("some", logger, rawConnMock, serializerMock, false)
	assert.NotNil(t, client)
	assert.Nil(t, err)

	res, err := client.Treatment("key1", "buck1", "feat1", map[string]interface{}{"a": 1})
	assert.Nil(t, err)
	assert.Equal(t, "on", res.Treatment)
	assert.Nil(t, res.Impression)
}

func TestClientGetTreatmentWithConfig(t *testing.T) {

	logger := logging.NewLogger(nil)

	rawConnMock := &transferMocks.RawConnMock{}
	rawConnMock.On("SendMessage", []byte("registrationMessage")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte("registrationSuccess"), nil).Once()
	rawConnMock.On("SendMessage", []byte("treatmentWithConfigMessage")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte("treatmentWithConfigResult"), nil).Once()

	serializerMock := &serializerMocks.SerializerMock{}
	serializerMock.On("Serialize", proto1Mocks.NewRegisterRPC("some", false)).Return([]byte("registrationMessage"), nil).Once()
	serializerMock.On("Parse", []byte("registrationSuccess"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.ResponseWrapper[v1.RegisterPayload]) = v1.ResponseWrapper[v1.RegisterPayload]{Status: v1.ResultOk}
	}).Once()

	serializerMock.On("Serialize", proto1Mocks.NewTreatmentRPC("key1", "buck1", "feat1", map[string]interface{}{"a": 1}, true)).
		Return([]byte("treatmentWithConfigMessage"), nil).Once()
	serializerMock.On("Parse", []byte("treatmentWithConfigResult"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.ResponseWrapper[v1.TreatmentPayload]) = v1.ResponseWrapper[v1.TreatmentPayload]{
			Status:  v1.ResultOk,
			Payload: v1.TreatmentPayload{Treatment: "on", Config: lang.Ref(`{"some": 1}`)},
		}
	}).Once()
	client, err := New("some", logger, rawConnMock, serializerMock, false)
	assert.NotNil(t, client)
	assert.Nil(t, err)

	res, err := client.TreatmentWithConfig("key1", "buck1", "feat1", map[string]interface{}{"a": 1})
	assert.Nil(t, err)
	assert.Equal(t, lang.Ref(`{"some": 1}`), res.Config)
	assert.Equal(t, "on", res.Treatment)
	assert.Nil(t, res.Impression)
}

func TestTrack(t *testing.T) {

	logger := logging.NewLogger(nil)

	rawConnMock := &transferMocks.RawConnMock{}
	rawConnMock.On("SendMessage", []byte("registrationMessage")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte("registrationSuccess"), nil).Once()
	rawConnMock.On("SendMessage", []byte("trackMessage")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte("trackResult"), nil).Once()

	serializerMock := &serializerMocks.SerializerMock{}
	serializerMock.On("Serialize", proto1Mocks.NewRegisterRPC("some", false)).Return([]byte("registrationMessage"), nil).Once()
	serializerMock.On("Parse", []byte("registrationSuccess"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.ResponseWrapper[v1.RegisterPayload]) = v1.ResponseWrapper[v1.RegisterPayload]{Status: v1.ResultOk}
	}).Once()

	serializerMock.On("Serialize", proto1Mocks.NewTrackRPC("key1", "user", "checkin", lang.Ref(2.74), map[string]interface{}{"p1": 123})).
		Return([]byte("trackMessage"), nil).Once()
	serializerMock.On("Parse", []byte("trackResult"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.ResponseWrapper[v1.TrackPayload]) = *proto1Mocks.NewTrackResp(true)
	}).Once()
	client, err := New("some", logger, rawConnMock, serializerMock, false)
	assert.NotNil(t, client)
	assert.Nil(t, err)

	err = client.Track("key1", "user", "checkin", lang.Ref(2.74), map[string]interface{}{"p1": 123})
	assert.Nil(t, err)
}

func TestClientGetTreatmentWithImpression(t *testing.T) {

	logger := logging.NewLogger(nil)

	rawConnMock := &transferMocks.RawConnMock{}
	rawConnMock.On("SendMessage", []byte("registrationMessage")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte("registrationSuccess"), nil).Once()
	rawConnMock.On("SendMessage", []byte("treatmentMessage")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte("treatmentResult"), nil).Once()

	serializerMock := &serializerMocks.SerializerMock{}
	serializerMock.On("Serialize", proto1Mocks.NewRegisterRPC("some", true)).Return([]byte("registrationMessage"), nil).Once()
	serializerMock.On("Parse", []byte("registrationSuccess"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.ResponseWrapper[v1.RegisterPayload]) = v1.ResponseWrapper[v1.RegisterPayload]{Status: v1.ResultOk}
	}).Once()

	serializerMock.On("Serialize", proto1Mocks.NewTreatmentRPC("key1", "buck1", "feat1", map[string]interface{}{"a": 1}, false)).
		Return([]byte("treatmentMessage"), nil).Once()
	serializerMock.On("Parse", []byte("treatmentResult"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.ResponseWrapper[v1.TreatmentPayload]) = v1.ResponseWrapper[v1.TreatmentPayload]{
			Status: v1.ResultOk,
			Payload: v1.TreatmentPayload{
				Treatment:    "on",
				ListenerData: &v1.ListenerExtraData{Label: "l1", Timestamp: 123, ChangeNumber: 1234},
			},
		}
	}).Once()
	client, err := New("some", logger, rawConnMock, serializerMock, true)
	assert.NotNil(t, client)
	assert.Nil(t, err)

	res, err := client.Treatment("key1", "buck1", "feat1", map[string]interface{}{"a": 1})
	assert.Nil(t, err)
	assert.Equal(t, "on", res.Treatment)
	validateImpression(t, &dtos.Impression{
		KeyName:      "key1",
		BucketingKey: "buck1",
		FeatureName:  "feat1",
		Treatment:    "on",
		Label:        "l1",
		ChangeNumber: 1234,
		Time:         123,
	}, res.Impression)

}

func TestClientGetTreatmentsNoImpression(t *testing.T) {

	logger := logging.NewLogger(nil)

	rawConnMock := &transferMocks.RawConnMock{}
	rawConnMock.On("SendMessage", []byte("registrationMessage")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte("registrationSuccess"), nil).Once()
	rawConnMock.On("SendMessage", []byte("treatmentsMessage")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte("treatmentsResult"), nil).Once()

	serializerMock := &serializerMocks.SerializerMock{}
	serializerMock.On("Serialize", proto1Mocks.NewRegisterRPC("some", false)).Return([]byte("registrationMessage"), nil).Once()
	serializerMock.On("Parse", []byte("registrationSuccess"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.ResponseWrapper[v1.RegisterPayload]) = v1.ResponseWrapper[v1.RegisterPayload]{Status: v1.ResultOk}
	}).Once()

	serializerMock.On("Serialize", proto1Mocks.NewTreatmentsRPC("key1", "buck1", []string{"a", "b", "c"}, map[string]interface{}{"a": 1}, false)).
		Return([]byte("treatmentsMessage"), nil).Once()
	serializerMock.On("Parse", []byte("treatmentsResult"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.ResponseWrapper[v1.TreatmentsPayload]) = v1.ResponseWrapper[v1.TreatmentsPayload]{
			Status:  v1.ResultOk,
			Payload: v1.TreatmentsPayload{Results: []v1.TreatmentPayload{{Treatment: "on"}, {Treatment: "off"}, {Treatment: "na"}}}}
	}).Once()
	client, err := New("some", logger, rawConnMock, serializerMock, false)
	assert.NotNil(t, client)
	assert.Nil(t, err)

	res, err := client.Treatments("key1", "buck1", []string{"a", "b", "c"}, map[string]interface{}{"a": 1})
	assert.Nil(t, err)
	assert.Equal(t, "on", res["a"].Treatment)
	assert.Nil(t, res["a"].Impression)
	assert.Equal(t, "off", res["b"].Treatment)
	assert.Nil(t, res["b"].Impression)
	assert.Equal(t, "na", res["c"].Treatment)
	assert.Nil(t, res["c"].Impression)

}

func TestClientGetTreatmentsWithConfig(t *testing.T) {

	logger := logging.NewLogger(nil)

	rawConnMock := &transferMocks.RawConnMock{}
	rawConnMock.On("SendMessage", []byte("registrationMessage")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte("registrationSuccess"), nil).Once()
	rawConnMock.On("SendMessage", []byte("treatmentsWithConfigMessage")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte("treatmentsWithConfigResult"), nil).Once()

	serializerMock := &serializerMocks.SerializerMock{}
	serializerMock.On("Serialize", proto1Mocks.NewRegisterRPC("some", false)).Return([]byte("registrationMessage"), nil).Once()
	serializerMock.On("Parse", []byte("registrationSuccess"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.ResponseWrapper[v1.RegisterPayload]) = v1.ResponseWrapper[v1.RegisterPayload]{Status: v1.ResultOk}
	}).Once()

	serializerMock.On("Serialize", proto1Mocks.NewTreatmentsRPC("key1", "buck1", []string{"a", "b", "c"}, map[string]interface{}{"a": 1}, true)).
		Return([]byte("treatmentsWithConfigMessage"), nil).Once()
	serializerMock.On("Parse", []byte("treatmentsWithConfigResult"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.ResponseWrapper[v1.TreatmentsPayload]) = v1.ResponseWrapper[v1.TreatmentsPayload]{
			Status: v1.ResultOk,
			Payload: v1.TreatmentsPayload{Results: []v1.TreatmentPayload{
				{Treatment: "on", Config: lang.Ref(`{"some": 2}`)}, {Treatment: "off"}, {Treatment: "na"}}}}
	}).Once()
	client, err := New("some", logger, rawConnMock, serializerMock, false)
	assert.NotNil(t, client)
	assert.Nil(t, err)

	res, err := client.TreatmentsWithConfig("key1", "buck1", []string{"a", "b", "c"}, map[string]interface{}{"a": 1})
	assert.Nil(t, err)
	assert.Equal(t, "on", res["a"].Treatment)
	assert.Nil(t, res["a"].Impression)
	assert.Equal(t, lang.Ref(`{"some": 2}`), res["a"].Config)
	assert.Equal(t, "off", res["b"].Treatment)
	assert.Nil(t, res["b"].Impression)
	assert.Nil(t, res["b"].Config)
	assert.Equal(t, "na", res["c"].Treatment)
	assert.Nil(t, res["c"].Config)
	assert.Nil(t, res["c"].Impression)

}

func TestClientGetTreatmentsWithImpression(t *testing.T) {

	logger := logging.NewLogger(nil)

	rawConnMock := &transferMocks.RawConnMock{}
	rawConnMock.On("SendMessage", []byte("registrationMessage")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte("registrationSuccess"), nil).Once()
	rawConnMock.On("SendMessage", []byte("treatmentsMessage")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte("treatmentsResult"), nil).Once()

	serializerMock := &serializerMocks.SerializerMock{}
	serializerMock.On("Serialize", proto1Mocks.NewRegisterRPC("some", true)).Return([]byte("registrationMessage"), nil).Once()
	serializerMock.On("Parse", []byte("registrationSuccess"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.ResponseWrapper[v1.RegisterPayload]) = v1.ResponseWrapper[v1.RegisterPayload]{Status: v1.ResultOk}
	}).Once()

	serializerMock.On("Serialize", proto1Mocks.NewTreatmentsRPC("key1", "buck1", []string{"a", "b", "c"}, map[string]interface{}{"a": 1}, false)).
		Return([]byte("treatmentsMessage"), nil).Once()
	serializerMock.On("Parse", []byte("treatmentsResult"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.ResponseWrapper[v1.TreatmentsPayload]) = v1.ResponseWrapper[v1.TreatmentsPayload]{
			Status: v1.ResultOk,
			Payload: v1.TreatmentsPayload{Results: []v1.TreatmentPayload{
				{Treatment: "on", ListenerData: &v1.ListenerExtraData{Label: "l1", Timestamp: 1, ChangeNumber: 5}},
				{Treatment: "off", ListenerData: &v1.ListenerExtraData{Label: "l2", Timestamp: 2, ChangeNumber: 6}},
				{Treatment: "na", ListenerData: &v1.ListenerExtraData{Label: "l3", Timestamp: 3, ChangeNumber: 7}},
			}}}
	}).Once()
	client, err := New("some", logger, rawConnMock, serializerMock, true)
	assert.NotNil(t, client)
	assert.Nil(t, err)

	res, err := client.Treatments("key1", "buck1", []string{"a", "b", "c"}, map[string]interface{}{"a": 1})
	assert.Nil(t, err)
	assert.Equal(t, "on", res["a"].Treatment)
	assert.Equal(t, "off", res["b"].Treatment)
	assert.Equal(t, "na", res["c"].Treatment)

	validateImpression(t, &dtos.Impression{
		KeyName:      "key1",
		BucketingKey: "buck1",
		FeatureName:  "a",
		Treatment:    "on",
		Label:        "l1",
		ChangeNumber: 5,
		Time:         1,
	}, res["a"].Impression)
	validateImpression(t, &dtos.Impression{
		KeyName:      "key1",
		BucketingKey: "buck1",
		FeatureName:  "b",
		Treatment:    "off",
		Label:        "l2",
		ChangeNumber: 6,
		Time:         2,
	}, res["b"].Impression)
	validateImpression(t, &dtos.Impression{
		KeyName:      "key1",
		BucketingKey: "buck1",
		FeatureName:  "c",
		Treatment:    "na",
		Label:        "l3",
		ChangeNumber: 7,
		Time:         3,
	}, res["c"].Impression)

}

func TestClientSplitNames(t *testing.T) {

	logger := logging.NewLogger(nil)

	rawConnMock := &transferMocks.RawConnMock{}
	rawConnMock.On("SendMessage", []byte("registrationMessage")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte("registrationSuccess"), nil).Once()
	rawConnMock.On("SendMessage", []byte("splitNamesMessage")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte("splitNamesResult"), nil).Once()

	serializerMock := &serializerMocks.SerializerMock{}
	serializerMock.On("Serialize", proto1Mocks.NewRegisterRPC("some", false)).Return([]byte("registrationMessage"), nil).Once()
	serializerMock.On("Parse", []byte("registrationSuccess"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.ResponseWrapper[v1.RegisterPayload]) = v1.ResponseWrapper[v1.RegisterPayload]{Status: v1.ResultOk}
	}).Once()

	serializerMock.On("Serialize", proto1Mocks.NewSplitNamesRPC()).
		Return([]byte("splitNamesMessage"), nil).Once()
	serializerMock.On("Parse", []byte("splitNamesResult"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.ResponseWrapper[v1.SplitNamesPayload]) = v1.ResponseWrapper[v1.SplitNamesPayload]{
			Status:  v1.ResultOk,
			Payload: v1.SplitNamesPayload{Names: []string{"s1", "s2"}},
		}
	}).Once()
	client, err := New("some", logger, rawConnMock, serializerMock, false)
	assert.NotNil(t, client)
	assert.Nil(t, err)

	res, err := client.SplitNames()
	assert.Nil(t, err)
	assert.Equal(t, []string{"s1", "s2"}, res)
}

func TestClientSplits(t *testing.T) {

	logger := logging.NewLogger(nil)

	rawConnMock := &transferMocks.RawConnMock{}
	rawConnMock.On("SendMessage", []byte("registrationMessage")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte("registrationSuccess"), nil).Once()
	rawConnMock.On("SendMessage", []byte("splitsMessage")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte("splitsResult"), nil).Once()

	serializerMock := &serializerMocks.SerializerMock{}
	serializerMock.On("Serialize", proto1Mocks.NewRegisterRPC("some", false)).Return([]byte("registrationMessage"), nil).Once()
	serializerMock.On("Parse", []byte("registrationSuccess"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.ResponseWrapper[v1.RegisterPayload]) = v1.ResponseWrapper[v1.RegisterPayload]{Status: v1.ResultOk}
	}).Once()

	serializerMock.On("Serialize", proto1Mocks.NewSplitsRPC()).
		Return([]byte("splitsMessage"), nil).Once()
	serializerMock.On("Parse", []byte("splitsResult"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.ResponseWrapper[v1.SplitsPayload]) = v1.ResponseWrapper[v1.SplitsPayload]{
			Status: v1.ResultOk,
			Payload: v1.SplitsPayload{Splits: []v1.SplitPayload{
				{Name: "s1", TrafficType: "tt1", Killed: true, Treatments: []string{"on", "off"}, ChangeNumber: 1, Configs: map[string]string{"on": "a"}},
				{Name: "s2", TrafficType: "tt1", Killed: true, Treatments: []string{"on", "off"}, ChangeNumber: 2, Configs: map[string]string{"on": "a"}},
			}},
		}
	}).Once()

	client, err := New("some", logger, rawConnMock, serializerMock, false)
	assert.NotNil(t, client)
	assert.Nil(t, err)

	res, err := client.Splits()
	assert.Nil(t, err)
	assert.Equal(t, []sdk.SplitView{
		{Name: "s1", TrafficType: "tt1", Killed: true, Treatments: []string{"on", "off"}, ChangeNumber: 1, Configs: map[string]string{"on": "a"}},
		{Name: "s2", TrafficType: "tt1", Killed: true, Treatments: []string{"on", "off"}, ChangeNumber: 2, Configs: map[string]string{"on": "a"}},
	}, res)
}

func TestClientSplit(t *testing.T) {

	logger := logging.NewLogger(nil)

	rawConnMock := &transferMocks.RawConnMock{}
	rawConnMock.On("SendMessage", []byte("registrationMessage")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte("registrationSuccess"), nil).Once()
	rawConnMock.On("SendMessage", []byte("splitMessage")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte("splitResult"), nil).Once()

	serializerMock := &serializerMocks.SerializerMock{}
	serializerMock.On("Serialize", proto1Mocks.NewRegisterRPC("some", false)).Return([]byte("registrationMessage"), nil).Once()
	serializerMock.On("Parse", []byte("registrationSuccess"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.ResponseWrapper[v1.RegisterPayload]) = v1.ResponseWrapper[v1.RegisterPayload]{Status: v1.ResultOk}
	}).Once()

	serializerMock.On("Serialize", proto1Mocks.NewSplitRPC("s1")).
		Return([]byte("splitMessage"), nil).Once()
	serializerMock.On("Parse", []byte("splitResult"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.ResponseWrapper[v1.SplitPayload]) = v1.ResponseWrapper[v1.SplitPayload]{
			Status: v1.ResultOk, Payload: v1.SplitPayload{
				Name:         "s1",
				TrafficType:  "tt1",
				Killed:       true,
				Treatments:   []string{"on", "off"},
				ChangeNumber: 1,
				Configs:      map[string]string{"on": "a"},
			}}
	}).Once()

	client, err := New("some", logger, rawConnMock, serializerMock, false)
	assert.NotNil(t, client)
	assert.Nil(t, err)

	res, err := client.Split("s1")
	assert.Nil(t, err)
	assert.Equal(t, &sdk.SplitView{
		Name:         "s1",
		TrafficType:  "tt1",
		Killed:       true,
		Treatments:   []string{"on", "off"},
		ChangeNumber: 1,
		Configs:      map[string]string{"on": "a"},
	}, res)
}

func validateImpression(t *testing.T, expected *dtos.Impression, actual *dtos.Impression) {
	t.Helper()
	assert.Equal(t, expected.BucketingKey, actual.BucketingKey)
	assert.Equal(t, expected.ChangeNumber, actual.ChangeNumber)
	assert.Equal(t, expected.FeatureName, actual.FeatureName)
	assert.Equal(t, expected.KeyName, actual.KeyName)
	assert.Equal(t, expected.Label, actual.Label)
	assert.Equal(t, expected.Time, actual.Time)
	assert.Equal(t, expected.Treatment, actual.Treatment)
	assert.Equal(t, expected.Label, actual.Label)

}
