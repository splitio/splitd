package sdk

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/splitio/go-split-commons/v4/dtos"
	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio/sdk/conf"
	"github.com/splitio/splitd/splitio/sdk/types"
	"github.com/stretchr/testify/assert"
)

func TestInstantiationAndGetTreatmentE2E(t *testing.T) {
	metricsInitCalled := 0
	mockedSplit1 := dtos.SplitDTO{
		Algo:                  2,
		ChangeNumber:          123,
		DefaultTreatment:      "default",
		Killed:                false,
		Name:                  "split",
		Seed:                  1234,
		Status:                "ACTIVE",
		TrafficAllocation:     1,
		TrafficAllocationSeed: -1667452163,
		TrafficTypeName:       "tt1",
		Conditions: []dtos.ConditionDTO{
			{
				ConditionType: "ROLLOUT",
				Label:         "in segment all",
				MatcherGroup: dtos.MatcherGroupDTO{
					Combiner: "AND",
					Matchers: []dtos.MatcherDTO{{MatcherType: "ALL_KEYS"}},
				},
				Partitions: []dtos.PartitionDTO{{Size: 100, Treatment: "on"}},
			},
		},
	}
	mockedSplit2 := dtos.SplitDTO{Name: "split2", Killed: true, Status: "ACTIVE"}
	mockedSplit3 := dtos.SplitDTO{Name: "split3", Killed: true, Status: "INACTIVE"}

	sdkServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/splitChanges", r.URL.Path)

		splitChanges := dtos.SplitChangesDTO{
			Splits: []dtos.SplitDTO{mockedSplit1, mockedSplit2, mockedSplit3},
			Since:  3,
			Till:   3,
		}

		raw, err := json.Marshal(splitChanges)
		assert.Nil(t, err)

		w.Write(raw)
	}))
	defer sdkServer.Close()

	var eventsCalls int32
	eventsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&eventsCalls, 1)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/testImpressions/bulk", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		assert.Nil(t, err)

		var imps []dtos.ImpressionsDTO
		assert.Nil(t, json.Unmarshal(body, &imps))

		assert.Equal(t, "split", imps[0].TestName)
		assert.Equal(t, 1, len(imps[0].KeyImpressions))
		w.WriteHeader(200)
	}))
	defer eventsServer.Close()

	telemetryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/metrics/config":
			metricsInitCalled++
			rBody, _ := ioutil.ReadAll(r.Body)
			var dataInPost dtos.Config
			err := json.Unmarshal(rBody, &dataInPost)
			assert.Nil(t, err)
		}
		fmt.Fprintln(w, "ok")
	}))
	defer telemetryServer.Close()

	sdkConf := conf.DefaultConfig()
	sdkConf.URLs.Events = eventsServer.URL
	sdkConf.URLs.SDK = sdkServer.URL
	sdkConf.URLs.Telemetry = telemetryServer.URL
	sdkConf.StreamingEnabled = false

	logger := logging.NewLogger(nil)
	client, err := New(logger, "someApikey", sdkConf)
	assert.Nil(t, err)

	res, err := client.Treatment(&types.ClientConfig{}, "aaaaaaklmnbv", nil, "split", nil)
	assert.Equal(t, "on", res.Treatment)

	assert.Nil(t, client.Shutdown())
	assert.Equal(t, int32(1), atomic.LoadInt32(&eventsCalls))
}
