package main

import (
	"fmt"
	"time"

	"github.com/markusmobius/go-dateparser"
	splunk "github.com/pvik/go-splunk-rest"
)

func stringValFromInterface(val interface{}) string {
	return fmt.Sprintf("%s", val)
}

func parseTimeString(timeStr string) (time.Time, bool, error) {
	if timeStr == "now" {
		return time.Now(), true, nil
	} else {
		var err error
		validTime, err := time.Parse(splunk.TIME_FORMAT, timeStr)
		if err != nil {
			// Check to see if string can be parsed as a relative time
			dt, err := dateparser.Parse(
				&dateparser.Configuration{
					CurrentTime: time.Now(),
				},
				timeStr,
			)
			if err != nil {
				return time.Now(), false, fmt.Errorf("invalid time string (%s): %s", timeStr, err)
			}

			return dt.Time, true, nil
		}

		return validTime, true, nil
	}
}
