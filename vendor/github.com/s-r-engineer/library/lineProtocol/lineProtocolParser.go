package lineProtocol

import (
	"strings"
	"time"

	"github.com/influxdata/line-protocol/v2/lineprotocol"
	"go.uber.org/multierr"
)

func LineProtocolParser(lines string) (l Lines, err error) {
	reader := lineprotocol.NewDecoder(strings.NewReader(lines))
lineLoop:
	for {
		lineToFill := Line{}
		if !reader.Next() {
			break
		}
		measurement, processError := reader.Measurement()
		if processError != nil {
			err = multierr.Append(err, processError)
			continue lineLoop
		}
		lineToFill.Measurement = string(measurement)
		lineToFill.Tags = make(map[string]string)
		lineToFill.Fields = make(map[string]string)
		for {
			k, v, processError := reader.NextTag()
			if processError != nil {
				err = multierr.Append(err, processError)
				continue lineLoop
			}
			if string(k) == "" {
				break
			}
			lineToFill.Tags[string(k)] = string(v)
		}
		for {
			k, v, processError := reader.NextField()
			if processError != nil {
				err = multierr.Append(err, processError)
				continue lineLoop
			}
			if string(k) == "" {
				break
			}
			lineToFill.Fields[string(k)] = v.String()
		}
		currentTime := time.Now()
		t, processError := reader.Time(defaultPrecision, currentTime)
		if processError != nil {
			err = multierr.Append(err, processError)
			continue lineLoop
		}
		if currentTime != t {
			lineToFill.Timestamp = t
		}
		l = append(l, lineToFill)
	}
	// libraryLogging.Dumper(err)
	return l, err
}
