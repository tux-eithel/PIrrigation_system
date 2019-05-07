package main

import (
	"log"
	"sync"

	"gobot.io/x/gobot"
)

// initRemoteRobots initializes the remote worker.
// It make ah http call (gRPC in future) and way the respose.
func initRemoteRobots() bool {
	return true
}

func workRemoteRobots(robotName string, eventer gobot.Eventer, waitRobots *sync.WaitGroup) {
	commands := eventer.Subscribe()
	var err error
	defer waitRobots.Done()

	for e := range commands {
		switch e.Name {

		case startRemoteRobots: // Here we start remote robots.
			// Try to start the remote robots.
			err = doRemoteWork()
			if err != nil {
				log.Printf("unable to '%s' on robot '%s': %v\nThis schedule will be skipped...", e.Name, robotName, err)
			} else {
				// If everythings goes well we are going to start local robots.
				eventer.Publish(startRelay, struct{}{})
			}


		case stopWorkers: // Here we stop remote robots.
			statusExit, ok := e.Data.(StopSignal)
			if !ok || statusExit == stopAndQuit {
				// TODO: add some remote command to shutdown ?
				eventer.Unsubscribe(commands)
				return
			}

			if statusExit == stopRemote {
				// TODO: try to shutdown remote robots
				err = stopRemoteWork()
				if err != nil {
					eventer.Publish(stopWorkers, stopLocal)
				} else {
					eventer.Publish(stopWorkers, stopAndQuit)
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
