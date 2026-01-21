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

	"github.com/splitio/go-split-commons/v9/dtos"
	"github.com/splitio/go-split-commons/v9/service/api/specs"
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

		splitChanges := dtos.RuleChangesDTO{
			FeatureFlags: dtos.FeatureFlagsDTO{
				Splits: []dtos.SplitDTO{mockedSplit1, mockedSplit2, mockedSplit3},
				Since:  3,
				Till:   3,
			},
		}

		assert.Equal(t, "-1", r.URL.Query().Get("since"))
		assert.Equal(t, specs.FLAG_V1_3, r.URL.Query().Get("s"))

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
	stringConfig := "flag1_config"
	globalTreatment := "global_treatment"
	flag1Treatment := "flag1_treatment"
	sdkConf.FallbackTreatment = dtos.FallbackTreatmentConfig{
		GlobalFallbackTreatment: &dtos.FallbackTreatment{
			Treatment: &globalTreatment,
		},
		ByFlagFallbackTreatment: map[string]dtos.FallbackTreatment{
			"flag1": {
				Treatment: &flag1Treatment,
				Config:    &stringConfig,
			},
		},
	}

	logger := logging.NewLogger(nil)
	client, err := New(logger, "someApikey", sdkConf)
	assert.Nil(t, err)

	res, _ := client.Treatment(&types.ClientConfig{}, "aaaaaaklmnbv", nil, "split", nil)
	assert.Equal(t, "on", res.Treatment)

	assert.Nil(t, client.Shutdown())
	assert.Equal(t, int32(1), atomic.LoadInt32(&eventsCalls))
}

func TestInstantiationAndGetTreatmentE2EWithFallbackTreatment(t *testing.T) {
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

		splitChanges := dtos.RuleChangesDTO{
			FeatureFlags: dtos.FeatureFlagsDTO{
				Splits: []dtos.SplitDTO{mockedSplit1, mockedSplit2, mockedSplit3},
				Since:  3,
				Till:   3,
			},
		}

		assert.Equal(t, "-1", r.URL.Query().Get("since"))
		assert.Equal(t, specs.FLAG_V1_3, r.URL.Query().Get("s"))

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

		assert.Equal(t, "not_exist", imps[0].TestName)
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
	stringConfig := "flag1_config"
	globalTreatment := "global_treatment"
	flag1Treatment := "flag1_treatment"
	sdkConf.FallbackTreatment = dtos.FallbackTreatmentConfig{
		GlobalFallbackTreatment: &dtos.FallbackTreatment{
			Treatment: &globalTreatment,
		},
		ByFlagFallbackTreatment: map[string]dtos.FallbackTreatment{
			"flag1": {
				Treatment: &flag1Treatment,
				Config:    &stringConfig,
			},
		},
	}

	logger := logging.NewLogger(nil)
	client, err := New(logger, "someApikey", sdkConf)
	assert.Nil(t, err)

	opts := dtos.EvaluationOptions{
		Properties: map[string]interface{}{
			"pleassssse": "holaaaaa",
		},
	}

	res1, _ := client.Treatment(&types.ClientConfig{}, "aaaaaaklmnbv", nil, "not_exist", nil, client.WithEvaluationOptions(&opts))
	assert.Equal(t, "global_treatment", res1.Treatment)
	assert.Equal(t, "{\"pleassssse\":\"holaaaaa\"}", res1.Impression.Properties)

	assert.Nil(t, client.Shutdown())
	assert.Equal(t, int32(1), atomic.LoadInt32(&eventsCalls))
}

func TestInstantiationAndGetTreatmentE2EWithPrerequistesNotAchive(t *testing.T) {
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
		Prerequisites: []dtos.Prerequisite{
			{
				FeatureFlagName: "ff1",
				Treatments: []string{
					"off",
					"v1",
				},
			},
		},
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

	mockedSplit2 := dtos.SplitDTO{
		Algo:                  2,
		ChangeNumber:          123,
		DefaultTreatment:      "off",
		Killed:                false,
		Name:                  "ff1",
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

	sdkServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/splitChanges", r.URL.Path)

		splitChanges := dtos.RuleChangesDTO{
			FeatureFlags: dtos.FeatureFlagsDTO{
				Splits: []dtos.SplitDTO{mockedSplit1, mockedSplit2},
				Since:  3,
				Till:   3,
			},
		}

		assert.Equal(t, "-1", r.URL.Query().Get("since"))
		assert.Equal(t, specs.FLAG_V1_3, r.URL.Query().Get("s"))

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

	res1, _ := client.Treatment(&types.ClientConfig{}, "aaaaaaklmnbv", nil, "split", nil)
	assert.Equal(t, "default", res1.Treatment)

	assert.Nil(t, client.Shutdown())
	assert.Equal(t, int32(1), atomic.LoadInt32(&eventsCalls))
}

func TestInstantiationAndGetTreatmentE2EWithPrerequistesAchive(t *testing.T) {
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
		Prerequisites: []dtos.Prerequisite{
			{
				FeatureFlagName: "ff1",
				Treatments: []string{
					"on",
					"v1",
				},
			},
		},
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

	mockedSplit2 := dtos.SplitDTO{
		Algo:                  2,
		ChangeNumber:          123,
		DefaultTreatment:      "off",
		Killed:                false,
		Name:                  "ff1",
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

	sdkServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/splitChanges", r.URL.Path)

		splitChanges := dtos.RuleChangesDTO{
			FeatureFlags: dtos.FeatureFlagsDTO{
				Splits: []dtos.SplitDTO{mockedSplit1, mockedSplit2},
				Since:  3,
				Till:   3,
			},
		}

		assert.Equal(t, "-1", r.URL.Query().Get("since"))
		assert.Equal(t, specs.FLAG_V1_3, r.URL.Query().Get("s"))

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

	res1, _ := client.Treatment(&types.ClientConfig{}, "aaaaaaklmnbv", nil, "split", nil)
	assert.Equal(t, "on", res1.Treatment)

	assert.Nil(t, client.Shutdown())
	assert.Equal(t, int32(1), atomic.LoadInt32(&eventsCalls))
}

func TestInstantiationAndGetTreatmentE2EWithRBS(t *testing.T) {
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
				Label:         "default rule",
				MatcherGroup: dtos.MatcherGroupDTO{
					Combiner: "AND",
					Matchers: []dtos.MatcherDTO{
						{
							KeySelector: &dtos.KeySelectorDTO{
								TrafficType: "user",
							},
							MatcherType: "IN_RULE_BASED_SEGMENT",
							UserDefinedSegment: &dtos.UserDefinedSegmentMatcherDataDTO{
								SegmentName: "rbsegment1",
							},
							Negate: false,
						},
					},
				},
				Partitions: []dtos.PartitionDTO{
					{
						Size:      100,
						Treatment: "on",
					},
					{
						Size:      0,
						Treatment: "off",
					},
				},
			},
		},
	}

	semver := "3.4.5"
	attribute := "version"

	rbsegment1 := dtos.RuleBasedSegmentDTO{
		Name:   "rbsegment1",
		Status: "ACTIVE",
		Conditions: []dtos.RuleBasedConditionDTO{
			{
				MatcherGroup: dtos.MatcherGroupDTO{
					Combiner: "AND",
					Matchers: []dtos.MatcherDTO{
						{
							KeySelector: &dtos.KeySelectorDTO{
								TrafficType: "user",
								Attribute:   &attribute,
							},
							MatcherType: "EQUAL_TO_SEMVER",
							String:      &semver,
							Whitelist:   nil,
							Negate:      false,
						},
					},
				},
			},
		},
		TrafficTypeName: "user",
	}

	sdkServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/splitChanges", r.URL.Path)

		splitChanges := dtos.RuleChangesDTO{
			FeatureFlags: dtos.FeatureFlagsDTO{
				Splits: []dtos.SplitDTO{mockedSplit1},
				Since:  3,
				Till:   3,
			},
			RuleBasedSegments: dtos.RuleBasedSegmentsDTO{
				RuleBasedSegments: []dtos.RuleBasedSegmentDTO{rbsegment1},
				Since:             3,
				Till:              3,
			},
		}

		assert.Equal(t, "-1", r.URL.Query().Get("since"))
		assert.Equal(t, specs.FLAG_V1_3, r.URL.Query().Get("s"))

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
	attributes := make(map[string]interface{})
	attributes["version"] = "3.4.5"

	res1, _ := client.Treatment(&types.ClientConfig{}, "aaaaaaklmnbv", nil, "split", attributes)
	assert.Equal(t, "on", res1.Treatment)

	assert.Nil(t, client.Shutdown())
	assert.Equal(t, int32(1), atomic.LoadInt32(&eventsCalls))
}
