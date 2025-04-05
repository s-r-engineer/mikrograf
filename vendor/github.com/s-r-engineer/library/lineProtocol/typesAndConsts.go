package lineProtocol

import (
	"time"

	"github.com/influxdata/line-protocol/v2/lineprotocol"
)

const defaultPrecision = lineprotocol.Nanosecond

type Accumulator struct {
	accumulatedData []byte
	lock, unlock    func()
}

type Lines []Line

type Line struct {
	Measurement  string
	Tags, Fields map[string]string
	Timestamp    time.Time
}
