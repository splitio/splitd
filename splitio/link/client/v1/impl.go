package v1

import (
	"fmt"

	"github.com/splitio/go-split-commons/v4/dtos"
	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio"
	"github.com/splitio/splitd/splitio/link/client/types"
	"github.com/splitio/splitd/splitio/link/protocol"
	protov1 "github.com/splitio/splitd/splitio/link/protocol/v1"
	"github.com/splitio/splitd/splitio/link/serializer"
	"github.com/splitio/splitd/splitio/link/transfer"
	"github.com/splitio/splitd/splitio/sdk"
)

const (
	Control = "control"
)

type Impl struct {
	logger           logging.LoggerInterface
	conn             transfer.RawConn
	serializer       serializer.Interface
	listenerFeedback bool
}

func New(id string, logger logging.LoggerInterface, conn transfer.RawConn, serializer serializer.Interface, listenerFeedback bool) (*Impl, error) {
	i := &Impl{
		logger:           logger,
		conn:             conn,
		serializer:       serializer,
		listenerFeedback: listenerFeedback,
	}

	if err := i.register(id, listenerFeedback); err != nil {
		i.conn.Shutdown()
		return nil, fmt.Errorf("error during client registration: %w", err)
	}

	return i, nil
}

// Treatment implements Interface
func (c *Impl) Treatment(key string, bucketingKey string, feature string, attrs map[string]interface{}) (*types.Result, error) {
	var bkp *string
	if bucketingKey != "" {
		bkp = &bucketingKey
	}
	rpc := protov1.RPC{
		RPCBase: protocol.RPCBase{Version: protocol.V1},
		OpCode:  protov1.OCTreatment,
		Args:    protov1.TreatmentArgs{Key: key, BucketingKey: bkp, Feature: feature, Attributes: attrs}.Encode(),
	}

	resp, err := doRPC[protov1.ResponseWrapper[protov1.TreatmentPayload]](c, &rpc)
	if err != nil {
		return &types.Result{Treatment: Control}, fmt.Errorf("error executing treatment rpc: %w", err)
	}

	if resp.Status != protov1.ResultOk {
		return &types.Result{Treatment: Control}, fmt.Errorf("server responded treatment rpc with error %d", resp.Status)
	}

	var imp *dtos.Impression
	if c.listenerFeedback && resp.Payload.ListenerData != nil {
		imp = &dtos.Impression{
			KeyName:      key,
			FeatureName:  feature,
			Treatment:    resp.Payload.Treatment,
			Time:         resp.Payload.ListenerData.Timestamp,
			ChangeNumber: resp.Payload.ListenerData.ChangeNumber,
			Label:        resp.Payload.ListenerData.Label,
			BucketingKey: bucketingKey,
		}
	}

	return &types.Result{Treatment: resp.Payload.Treatment, Impression: imp}, nil
}

// Treatment implements Interface
func (c *Impl) Treatments(key string, bucketingKey string, features []string, attrs map[string]interface{}) (types.Results, error) {
	var bkp *string
	if bucketingKey != "" {
		bkp = &bucketingKey
	}
	rpc := protov1.RPC{
		RPCBase: protocol.RPCBase{Version: protocol.V1},
		OpCode:  protov1.OCTreatments,
		Args:    protov1.TreatmentsArgs{Key: key, BucketingKey: bkp, Features: features, Attributes: attrs}.Encode(),
	}

	resp, err := doRPC[protov1.ResponseWrapper[protov1.TreatmentsPayload]](c, &rpc)
	if err != nil {
		return nil, fmt.Errorf("error executing treatments rpc: %w", err)
	}

	if resp.Status != protov1.ResultOk {
		return nil, fmt.Errorf("server responded treatments rpc with error %d", resp.Status)
	}

	results := make(types.Results)
	for idx := range features {
		var imp *dtos.Impression
		if c.listenerFeedback && resp.Payload.Results[idx].ListenerData != nil {
			imp = &dtos.Impression{
				KeyName:      key,
				FeatureName:  features[idx],
				Treatment:    resp.Payload.Results[idx].Treatment,
				Time:         resp.Payload.Results[idx].ListenerData.Timestamp,
				ChangeNumber: resp.Payload.Results[idx].ListenerData.ChangeNumber,
				Label:        resp.Payload.Results[idx].ListenerData.Label,
				BucketingKey: bucketingKey,
			}
		}
		results[features[idx]] = types.Result{Treatment: resp.Payload.Results[idx].Treatment, Impression: imp}
	}

	return results, nil
}

// TreatmentWithConfig implements types.ClientInterface
func (c *Impl) TreatmentWithConfig(key string, bucketingKey string, feature string, attrs map[string]interface{}) (*types.Result, error) {
	var bkp *string
	if bucketingKey != "" {
		bkp = &bucketingKey
	}
	rpc := protov1.RPC{
		RPCBase: protocol.RPCBase{Version: protocol.V1},
		OpCode:  protov1.OCTreatmentWithConfig,
		Args:    protov1.TreatmentArgs{Key: key, BucketingKey: bkp, Feature: feature, Attributes: attrs}.Encode(),
	}

	resp, err := doRPC[protov1.ResponseWrapper[protov1.TreatmentPayload]](c, &rpc)
	if err != nil {
		return &types.Result{Treatment: Control}, fmt.Errorf("error executing treatment rpc: %w", err)
	}

	if resp.Status != protov1.ResultOk {
		return &types.Result{Treatment: Control}, fmt.Errorf("server responded treatment rpc with error %d", resp.Status)
	}

	var imp *dtos.Impression
	if c.listenerFeedback && resp.Payload.ListenerData != nil {
		imp = &dtos.Impression{
			KeyName:      key,
			FeatureName:  feature,
			Treatment:    resp.Payload.Treatment,
			Time:         resp.Payload.ListenerData.Timestamp,
			ChangeNumber: resp.Payload.ListenerData.ChangeNumber,
			Label:        resp.Payload.ListenerData.Label,
			BucketingKey: bucketingKey,
		}
	}

	return &types.Result{
		Treatment:  resp.Payload.Treatment,
		Impression: imp,
		Config:     resp.Payload.Config,
	}, nil
}

// TreatmentsWithConfig implements types.ClientInterface
func (c *Impl) TreatmentsWithConfig(key string, bucketingKey string, features []string, attrs map[string]interface{}) (types.Results, error) {
	var bkp *string
	if bucketingKey != "" {
		bkp = &bucketingKey
	}
	rpc := protov1.RPC{
		RPCBase: protocol.RPCBase{Version: protocol.V1},
		OpCode:  protov1.OCTreatmentsWithConfig,
		Args:    protov1.TreatmentsArgs{Key: key, BucketingKey: bkp, Features: features, Attributes: attrs}.Encode(),
	}

	resp, err := doRPC[protov1.ResponseWrapper[protov1.TreatmentsPayload]](c, &rpc)
	if err != nil {
		return nil, fmt.Errorf("error executing treatments rpc: %w", err)
	}

	if resp.Status != protov1.ResultOk {
		return nil, fmt.Errorf("server responded treatments rpc with error %d", resp.Status)
	}

	results := make(types.Results)
	for idx := range features {
		var imp *dtos.Impression
		if c.listenerFeedback && resp.Payload.Results[idx].ListenerData != nil {
			imp = &dtos.Impression{
				KeyName:      key,
				FeatureName:  features[idx],
				Treatment:    resp.Payload.Results[idx].Treatment,
				Time:         resp.Payload.Results[idx].ListenerData.Timestamp,
				ChangeNumber: resp.Payload.Results[idx].ListenerData.ChangeNumber,
				Label:        resp.Payload.Results[idx].ListenerData.Label,
				BucketingKey: bucketingKey,
			}
		}
		results[features[idx]] = types.Result{
			Treatment:  resp.Payload.Results[idx].Treatment,
			Config:     resp.Payload.Results[idx].Config,
			Impression: imp,
		}
	}

	return results, nil
}

// Track implements types.ClientInterface
func (c *Impl) Track(key string, trafficType string, eventType string, value *float64, properties map[string]interface{}) error {

	rpc := protov1.RPC{
		RPCBase: protocol.RPCBase{Version: protocol.V1},
		OpCode:  protov1.OCTrack,
		Args:    protov1.TrackArgs{Key: key, TrafficType: trafficType, EventType: eventType, Value: value, Properties: properties}.Encode(),
	}

	resp, err := doRPC[protov1.ResponseWrapper[protov1.TrackPayload]](c, &rpc)
	if err != nil {
		return fmt.Errorf("error executing track rpc: %w", err)
	}

	if resp.Status != protov1.ResultOk {
		return fmt.Errorf("server responded track rpc with error %d", resp.Status)
	}

	return nil
}

func (c *Impl) SplitNames() ([]string, error) {
	rpc := protov1.RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: protov1.OCSplitNames}
	resp, err := doRPC[protov1.ResponseWrapper[protov1.SplitNamesPayload]](c, &rpc)
	if err != nil {
		return nil, fmt.Errorf("error executing split-names rpc: %w", err)
	}

	if resp.Status != protov1.ResultOk {
		return nil, fmt.Errorf("server responded split-names rpc with error %d", resp.Status)
	}

	return resp.Payload.Names, nil
}

func (c *Impl) Split(name string) (*sdk.SplitView, error) { // TODO(mredolatti): use a local dto instead of package sdk's
	rpc := protov1.RPC{
		RPCBase: protocol.RPCBase{Version: protocol.V1},
		OpCode:  protov1.OCSplit,
		Args:    protov1.SplitArgs{Name: name}.Encode(),
	}

	resp, err := doRPC[protov1.ResponseWrapper[protov1.SplitPayload]](c, &rpc)
	if err != nil {
		return nil, fmt.Errorf("error executing split rpc: %w", err)
	}

	if resp.Status != protov1.ResultOk {
		return nil, fmt.Errorf("server responded split rpc with error %d", resp.Status)
	}

	p := sdk.SplitView(resp.Payload)
	return &p, nil
}

func (c *Impl) Splits() ([]sdk.SplitView, error) {
	rpc := protov1.RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: protov1.OCSplits}
	resp, err := doRPC[protov1.ResponseWrapper[protov1.SplitsPayload]](c, &rpc)
	if err != nil {
		return nil, fmt.Errorf("error executing splits rpc: %w", err)
	}

	if resp.Status != protov1.ResultOk {
		return nil, fmt.Errorf("server responded splits rpc with error %d", resp.Status)
	}

	views := make([]sdk.SplitView, 0, len(resp.Payload.Splits))
	for _, v := range resp.Payload.Splits {
		views = append(views, sdk.SplitView(v))
	}

	return views, nil
}

func (c *Impl) register(id string, impressionsFeedback bool) error {
	var flags protov1.RegisterFlags
	if impressionsFeedback {
		flags |= 1 << protov1.RegisterFlagReturnImpressionData
	}
	rpc := protov1.RPC{
		RPCBase: protocol.RPCBase{Version: protocol.V1},
		OpCode:  protov1.OCRegister,
		Args:    protov1.RegisterArgs{ID: id, SDKVersion: fmt.Sprintf("splitd-%s", splitio.Version), Flags: flags}.Encode(),
	}

	resp, err := doRPC[protov1.ResponseWrapper[protov1.RegisterPayload]](c, &rpc)
	if err != nil {
		return fmt.Errorf("error executing register rpc: %w", err)
	}

	if resp.Status != protov1.ResultOk {
		return fmt.Errorf("server responded register rpc with error %d", resp.Status)
	}

	return nil
}

func doRPC[T any](c *Impl, rpc *protov1.RPC) (*T, error) {
	serialized, err := c.serializer.Serialize(rpc)
	if err != nil {
		return nil, fmt.Errorf("error serializing rpc: %w", err)
	}

	err = c.conn.SendMessage(serialized)
	if err != nil {
		return nil, fmt.Errorf("error sending message to split daemon: %w", err)
	}

	resp, err := c.conn.ReceiveMessage()
	if err != nil {
		return nil, fmt.Errorf("error reading response from daemon: %w", err)
	}

	var response T
	err = c.serializer.Parse(resp, &response)
	if err != nil {
		return nil, fmt.Errorf("error de-serializing server response: %w", err)
	}

	return &response, nil
}

func (c *Impl) Shutdown() error {
	return c.conn.Shutdown()
}

var _ types.ClientInterface = (*Impl)(nil)
