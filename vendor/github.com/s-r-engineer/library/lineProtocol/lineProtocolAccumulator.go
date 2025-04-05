package lineProtocol

import (
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/influxdata/line-protocol/v2/lineprotocol"
	libraryErrors "github.com/s-r-engineer/library/errors"
	librarySync "github.com/s-r-engineer/library/sync"
)

func NewAccumulator() Accumulator {
	lock1, unlock1 := librarySync.GetMutex()
	return Accumulator{
		accumulatedData: make([]byte, 0),
		lock:            lock1,
		unlock:          unlock1,
	}
}

func (a *Accumulator) GetBytes() []byte {
	return a.accumulatedData
}

func (a *Accumulator) AddLine(measurement string, fields map[string]any, tags map[string]string, timestamp time.Time) error {
	errWrapperErr, errWrapperString := libraryErrors.PartWrapErrorOrString("add line error")
	if len(fields) == 0 {
		return errWrapperString("zero fields")
	}
	if measurement == "" {
		return errWrapperString("add line error -> empty measurement")
	}
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	var enc lineprotocol.Encoder
	enc.SetPrecision(defaultPrecision)
	if enc.Err() != nil {
		return errWrapperErr(enc.Err())
	}
	enc.StartLine(measurement)
	if enc.Err() != nil {
		return errWrapperErr(enc.Err())
	}
	keys := make([]string, 0, len(tags))
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		enc.AddTag(key, tags[key])
		if enc.Err() != nil {
			return errWrapperErr(enc.Err())
		}
	}
	for key, value := range fields {
		newValue, ok := lineprotocol.NewValue(value)
		if !ok {
			return errWrapperString(fmt.Sprintf("wrong data type. Must be int64, uint64, float64, bool, string, []byte but got %s", reflect.TypeOf(value)))
		}
		enc.AddField(key, newValue)
		if enc.Err() != nil {
			return errWrapperErr(enc.Err())
		}
	}
	enc.EndLine(timestamp)
	if enc.Err() != nil {
		return errWrapperErr(enc.Err())
	}
	a.lock()
	defer a.unlock()
	a.accumulatedData = append(a.accumulatedData, enc.Bytes()...)
	return nil
}
