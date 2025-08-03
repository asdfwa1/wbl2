package main

import (
	"fmt"
	"os"
	"task8/ntp"
)

func main() {
	rt := &ntp.ResponseTime{}

	timeResponse, err := rt.GetTime("")
	if err != nil {
		fmt.Printf("Mistake: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Exact time %s from NTP server %s",
		timeResponse.Time.Format(ntp.TimeFormat),
		timeResponse.Host,
	)
}
