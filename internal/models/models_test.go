// Package models provides entity of metric.

package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCounterConstructor(t *testing.T) {
	c := CounterConstructor(int64(25))
	assert.Equal(t, int64(25), *c.Delta)
}

func TestGaugeConstructor(t *testing.T) {
	g := GaugeConstructor(float64(50), "temp")
	assert.Equal(t, float64(50), *g.Value)
	assert.Equal(t, "temp", g.ID)
}
