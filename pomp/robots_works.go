package main

import (
	"log"
	"sync"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/drivers/spi"
)

// workRelay does the raley work.
func workRelay(robotName string, relay *gpio.RelayDriver, eventer gobot.Eventer, waitRobots *sync.WaitGroup) {
	commands := eventer.Subscribe()
	var err error
	defer waitRobots.Done()

	for e := range commands {
		switch e.Name {

		// Here we start the raley.
		// If all goes well we are going to start the MCP
		case startRelay:
			err = relay.Off()
			if err != nil {
				log.Printf("unable to '%s' on robots '%s': %v\n", e.Name, robotName, err)
			} else {
				log.Println("start relay!")
				eventer.Publish(startMCP, struct{}{})
			}

		// Here we stop the relay.
		case stopWorkers:
			err = relay.On()
			if err != nil {
				log.Printf("unable to '%s' on robots '%s': %v\n", e.Name, robotName, err)
			} else {
				log.Printf("robot '%s' will be '%s'\n", robotName, e.Name)
			}

			if exitNow, ok := e.Data.(bool); !ok || exitNow {
				eventer.Unsubscribe(commands)
				return
			}

		}
	}
}

// workMCP does the MCP work
func workMCP(robotName string, mcp *spi.MCP3008Driver, eventer gobot.Eventer, waitRobots *sync.WaitGroup) {
	commands := eventer.Subscribe()
	var err error
	var stopReadAnalogData chan struct{}
	var analogData <-chan *gobot.Event
	defer waitRobots.Done()

	for e := range commands {
		switch e.Name {

		case startMCP:
			if stopReadAnalogData != nil {
				log.Printf("robot '%s' already started... skip!\n", robotName)
				continue
			}
			log.Println("start mcp!")
			go func() {
				analogData, stopReadAnalogData = readsFromMCP(mcp, 250*time.Millisecond, 100)
				for ae := range analogData {
					switch ae.Name {
					case gpio.Error:
						err = ae.Data.(error)
						log.Printf("robot '%s' unable to read value: %v... for security reason we are going to shut down the system!\n\n", robotName, err)
						eventer.Publish(stopWorkers, true)
						return
					case gpio.Data:
						value := ae.Data.(int)
						log.Printf("robot '%s' seems like there is no water '%d'... we are going to shut down the system!\n", robotName, value)
						eventer.Publish(stopWorkers, true)
						return
					}
				}
			}()

		// Here we are going to close the MCP
		case stopWorkers:
			if stopReadAnalogData != nil {
				stopReadAnalogData <- struct{}{}
				stopReadAnalogData = nil
			}

			log.Printf("robot '%s' will be '%s'\n", robotName, stopWorkers)
			if exitNow, ok := e.Data.(bool); !ok || exitNow {
				eventer.Unsubscribe(commands)
				return
			}
		}

	}
}
