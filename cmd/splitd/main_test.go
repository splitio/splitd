package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetupprofiler(t *testing.T) {

	assert.Nil(t, setupProfiler())

	os.Setenv("SPLITD_PROFILING", "true")
	assert.NotNil(t, setupProfiler())
	os.Setenv("SPLITD_PROFILING", "1")
	assert.NotNil(t, setupProfiler())
	os.Setenv("SPLITD_PROFILING", "on")
	assert.NotNil(t, setupProfiler())
	os.Setenv("SPLITD_PROFILING", "On")
	assert.NotNil(t, setupProfiler())
	os.Setenv("SPLITD_PROFILING", "EnabLed")
	assert.NotNil(t, setupProfiler())

}
