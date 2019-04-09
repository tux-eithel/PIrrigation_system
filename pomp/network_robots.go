package main

import (
	"log"
	"sync"

	"gobot.io/x/gobot"
)

func initRemoteRobots() bool {
	return true
}

func workRemoteRobots(robotName string, eventer gobot.Eventer, waitRobots *sync.WaitGroup) {
	commands := eventer.Subscribe()
	var err error
	defer waitRobots.Done()

	for e := range commands {
		switch e.Name {

		case startRemoteRobots:

			// Try to start the remote robots.
			err = doRemoteWork()
			if err != nil {
				log.Printf("unable to '%s' on robot '%s': %v\nThis schedule will be skipped...", e.Name, robotName, err)
			} else {
				// If everythings goes well we are going to start local robots.
				eventer.Publish(startRelay, struct{}{})
			}

		// Here we stop remote robots.
		case stopWorkers:

			statusExit, ok := e.Data.(StopSignal)
			if !ok || statusExit == stopQuit {
				eventer.Unsubscribe(commands)
				return
			}

			if statusExit == stopRemote {
				// TODO: try to shutdown remote robots
				err = stopRemoteWork()
				if err != nil {
					eventer.Publish(stopWorkers, stopLocal)
				} else {
					eventer.Publish(stopWorkers, stopQuit)
				}

			}

		}
	}

}

func doRemoteWork() error {
	// TODO: call using gRPC, the remote robots and wait the response
	return nil
}

func stopRemoteWork() error {
	// TODO: call using gRPC, the remote robots, and wait the response
	return nil
}
