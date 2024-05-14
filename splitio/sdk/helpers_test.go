package sdk

import (
	"testing"

	"github.com/splitio/go-split-commons/v6/flagsets"
	sdkConf "github.com/splitio/splitd/splitio/sdk/conf"
	"github.com/stretchr/testify/assert"
)

func TestSetupImpressionsComponents(t *testing.T) {

	sdkCfg := sdkConf.DefaultConfig()
	storages := setupStorages(sdkCfg, flagsets.FlagSetFilter{})

	ic, err := setupImpressionsComponents(&sdkCfg.Impressions, storages.telemetry)
	assert.Nil(t, err)
	assert.NotNil(t, ic.counter)

	sdkCfg.Impressions.Mode = "debug"
	ic, err = setupImpressionsComponents(&sdkCfg.Impressions, storages.telemetry)
	assert.Nil(t, err)
	assert.Nil(t, ic.counter)
}

func TestNoOpTask(t *testing.T) {
	var task NoOpTask
	assert.Equal(t, false, task.IsRunning())
	task.Start()
	assert.Nil(t, task.Stop(true))
}
