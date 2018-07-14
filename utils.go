package main

import (
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/drivers/spi"
)

// readsFromMCP uses an eventer to read at specific interval
// the value from the chip.
// It returns two channels: one for listen the events, the other to close this function
func readsFromMCP(mcp *spi.MCP3008Driver, interval time.Duration, threshold int) (<-chan *gobot.Event, chan struct{}) {
	oldValue := -1 // TODO: make this a parameter
	streamValues := gobot.NewEventer()
	streamValues.AddEvent(gpio.Error)
	streamValues.AddEvent(gpio.Data)
	halt := make(chan struct{})
	events := streamValues.Subscribe()

	go func() {
		for {
			newValue, err := mcp.Read(2)
			if err != nil {
				streamValues.Publish(gpio.Error, err)
			} else if !(oldValue-threshold <= newValue && newValue <= oldValue+threshold) { // Send the value only if it differs for the previous one by the threshold
				if oldValue != -1 {
					streamValues.Publish(gpio.Data, newValue)
				}
				oldValue = newValue
			}

			select {
			case <-time.After(interval): // Wait an interval
			case <-halt: // Close the function
				close(events)
				return
			}
		}
	}()

	return events, halt
}
