package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/drivers/spi"
	"gobot.io/x/gobot/platforms/raspi"
)

const (
	startRelay = "START_RELAY"
	startMCP   = "START_MCP"

	// stopWorkers events accept booleans.
	// If is true, then it will stop and exit the worker
	stopWorkers = "STOP_WORKERS"
)

func main() {

	// Create a generic gobot.Eventer.
	// This eventer is useful to send events between workers.
	genericEventer := gobot.NewEventer()
	genericEventer.AddEvent(startRelay)
	genericEventer.AddEvent(startMCP)
	genericEventer.AddEvent(stopWorkers)

	// Instance the time scheduler
	scheduler := newWaterTimeManager()

	// The quit channel closes all the workers.
	waitRobots := &sync.WaitGroup{}

	waitRobots.Add(1)
	go consumerSchedule(scheduler, genericEventer, waitRobots)

	// Create the reaspberry.
	r := raspi.NewAdaptor()

	// Create the relay/led.
	// It's functions are On/Off/Toggle
	//relay := gpio.NewRelayDriver(r, "7")
	relay := gpio.NewLedDriver(r, "35")
	robotRelay := gobot.NewRobot("Relay Pompa",
		[]gobot.Connection{r},
		[]gobot.Device{relay},
	)
	waitRobots.Add(1)
	go workRelay(robotRelay.Name, relay, genericEventer, waitRobots)

	// Create the MCP driver.
	// This driver is useful to read some analogic.
	mcp := spi.NewMCP3008Driver(r, spi.WithSpeed(1350000))
	robotAcqua := gobot.NewRobot("Sensore Acqua",
		[]gobot.Connection{r},
		[]gobot.Device{mcp},
	)
	//mcp.interval = 200 * time.Millisecond
	waitRobots.Add(1)
	go workMCP(robotAcqua.Name, mcp, genericEventer, waitRobots)

	// Starts all the robots!
	// We pass "false" as parameter so we can manually stop the robots.
	robots := gobot.Robots{robotAcqua, robotRelay}
	err := robots.Start(false)
	if err != nil {
		log.Fatalln("Unable to start robots:", err)
	}

	// Function to read data.
	// In the future this will be an http handler.
	go func() {
		buff := bufio.NewReader(os.Stdin)
		for {
			fmt.Printf("Inserisci data inizio e fine separate da ' - ': ")
			text, _ := buff.ReadString('\n')
			t, err := newWaterTime(text[:len(text)-1])
			if err != nil {
				fmt.Printf("unable to parse time: %v, skip...\n", err)
				continue
			}
			_, err = scheduler.Append(t)
			if err != nil {
				fmt.Printf("this time collide with other times: %v\n", err)
				continue
			}
		}
	}()

	// Wait the ctrl-c signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	genericEventer.Publish(stopWorkers, true)

	// Stop all the robots
	log.Println("wait all robots closes...")
	waitRobots.Wait()
	err = robots.Stop()
	if err != nil {
		log.Fatalln("Unable to stop robots:", err)
	}

}
