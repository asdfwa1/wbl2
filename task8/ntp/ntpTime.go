package ntp

import (
	"fmt"
	"github.com/beevik/ntp"
	"time"
)

const TimeFormat = "2006-01-02 15:04:05"

var validHost = "0.beevik-ntp.pool.ntp.org"

type ResponseTime struct {
	Time time.Time
	Host string
}

func (rt *ResponseTime) GetTime(host string) (*ResponseTime, error) {
	if host == "" {
		host = validHost
	}

	ntpTime, err := ntp.Time(host)
	if err != nil {
		return nil, fmt.Errorf("error receiving NTP time: %w", err)
	}

	rt.Time = ntpTime
	rt.Host = host

	return rt, nil
}
