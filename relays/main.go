// Package main for relays manages the relays which controls the valves.
// It's separate from the the "pomp" main which instead controls also the schedulation.
// Basically this program accepts remote call (gRPC) and open/close valves.
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
)

func main() {
	r := raspi.NewAdaptor()

	// relayPlus and relayMinus controls the direcotion of current
	relayPlus := gpio.NewGroveRelayDriver(r, "11")
	relayMinus := gpio.NewGroveRelayDriver(r, "13")

	// First bi-stable valve
	valve1 := gpio.NewGroveRelayDriver(r, "15")

	// Prepare the robot
	r1 := gobot.NewRobot("relays",
		[]gobot.Connection{r},
		[]gobot.Device{valve1, relayPlus, relayMinus},
	)

	// Starts all the robots!
	// We pass "false" as parameter so we can manually stop the robots.
	robots := gobot.Robots{r1}
	err := robots.Start(false)
	if err != nil {
		log.Fatalln("Unable to start robots:", err)
	}

	go func() {

		// Reset the relays.
		// A release is closed when you set "HIGH" the pin.
		relayMinus.On()
		relayPlus.On()
		valve1.On()

		// After 5 seconds, we try to open all the valves.
		// This is made for security reason. If we are unable to close valves,
		// at least water will come out without damage the pomp.
		<-time.After(5 * time.Second)
		fmt.Println("try to open valve")
		valve1.Off()
		<-time.After(500 * time.Millisecond) // We wait a bit.
		fmt.Println("should be open")
		valve1.On()

		// After 5 second we try to close a valve.
		// This part simulate a "irrigation program" where user could
		// choose which part of the garden has to be irrigated.
		<-time.After(5 * time.Second)
		fmt.Println("try to close the valve")
		// Invert the current.
		relayMinus.Off()
		relayPlus.Off()

		// Activate the valve.
		valve1.Off()
		fmt.Println("should be close")
		<-time.After(500 * time.Millisecond) // We wait a bit.
		valve1.On()
	}()

	// Wait the ctrl-c signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	fmt.Println("closing procedure... open all the valve")
	relayMinus.On()
	relayPlus.On()
	valve1.Off()
	fmt.Println("should be open")
	<-time.After(500 * time.Millisecond)
	valve1.On()

	// Stop all the robots
	log.Println("wait all robots closes...")
	err = robots.Stop()
	if err != nil {
		log.Fatalln("Unable to stop robots:", err)
	}
}

// TODO: function to open all the valve (used at the start and at the end)

// TODO: define protobuf schema for comunication with the main raspberry
