package v1

import (
	"testing"

	"github.com/splitio/go-split-commons/v4/dtos"
	"github.com/splitio/go-toolkit/v5/logging"
	v1 "github.com/splitio/splitd/splitio/link/protocol/v1"
	proto1Mocks "github.com/splitio/splitd/splitio/link/protocol/v1/mocks"
	serializerMocks "github.com/splitio/splitd/splitio/link/serializer/mocks"
	transferMocks "github.com/splitio/splitd/splitio/link/transfer/mocks"
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

	serializerMock.On("Serialize", proto1Mocks.NewTreatmentRPC("key1", "buck1", "feat1", map[string]interface{}{"a": 1})).
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

	serializerMock.On("Serialize", proto1Mocks.NewTreatmentRPC("key1", "buck1", "feat1", map[string]interface{}{"a": 1})).
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

	serializerMock.On("Serialize", proto1Mocks.NewTreatmentsRPC("key1", "buck1", []string{"a", "b", "c"}, map[string]interface{}{"a": 1})).
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

	serializerMock.On("Serialize", proto1Mocks.NewTreatmentsRPC("key1", "buck1", []string{"a", "b", "c"}, map[string]interface{}{"a": 1})).
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

func ref[T any](t T) *T {
	return &t
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
