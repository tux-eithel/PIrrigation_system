package main

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"gobot.io/x/gobot"
)

const (
	parseTimeConst = "2006-01-02 15:04:05"
)

// waterTime is a strucut to keep start and end time.
type waterTime struct {
	start time.Time
	end   time.Time
}

// newWaterTime returns a new waterTime parsing a string.
// It may return an error if parsing goes bad.
func newWaterTime(row string) (*waterTime, error) {
	times := strings.Split(row, " - ")
	if len(times) != 2 {
		return nil, fmt.Errorf("too feew or too much elements")
	}
	loc, _ := time.LoadLocation("Europe/Berlin")
	e := &waterTime{}
	timeStart, err := time.ParseInLocation(parseTimeConst, times[0], loc)
	if err != nil {
		return nil, fmt.Errorf("unable to parse start date: %v", err)
	}
	e.start = timeStart

	timeEnd, err := time.ParseInLocation(parseTimeConst, times[1], loc)
	if err != nil {
		return nil, fmt.Errorf("unable to parse end date: %v", err)
	}
	e.end = timeEnd

	return e, nil

}

// waterTimeManager is a struct to keep the queue of waterTime.
// It provides a channel used to notify when a new waterTime has been added.
// This channel is useful to understand when the manager has changed.
// After a change has been notified, user must call GetNextSlot to get the correct waterTime
// (could return the same waterTime)
type waterTimeManager struct {
	// queue of waterTime
	times []*waterTime
	// channel to notify that queue is changed
	resetTimer chan bool
	sync.RWMutex
}

// newWaterTimeManager returns a waterTimeManager
func newWaterTimeManager() *waterTimeManager {

	wtm := &waterTimeManager{
		times:      make([]*waterTime, 0),
		resetTimer: make(chan bool),
	}

	return wtm

}

// Append tries to append a new waterTime to the queue manager.
// It also reorders times and notify the changes on resetTimer channel.
// It could return some errors if waterTime collides with other times already in the queue.
// It's thread safe.
func (wtm *waterTimeManager) Append(wt *waterTime) (bool, error) {
	wtm.Lock()
	defer wtm.Unlock()

	// checkTime checks if a new time could fit the queue.
	checkTime := func(input []*waterTime, t *waterTime) error {

		if t.start.Before(time.Now()) {
			return fmt.Errorf("time start is before Now: %v - %v", t.start, time.Now())
		}

		if t.end.Before(t.start) {
			return fmt.Errorf("time end is before time start: %v - %v", t.end, t.start)
		}

		for _, oldTime := range input {
			if t.start.After(oldTime.start) && t.start.Before(oldTime.end) {
				return fmt.Errorf("time start '%v' is inside %v-%v range", t.start, oldTime.start, oldTime.end)
			}
			if t.end.After(oldTime.start) && t.end.Before(oldTime.end) {
				return fmt.Errorf("time end '%v' is inside %v-%v range", t.end, oldTime.start, oldTime.end)
			}
		}
		return nil
	}

	if err := checkTime(wtm.times, wt); err != nil {
		return false, fmt.Errorf("unable to add schedule to manager: %v", err)
	}

	// Now it's safe to add the new time to the queue
	// and the reorder by start time
	wtm.times = append(wtm.times, wt)
	sort.Slice(wtm.times, func(i, j int) bool { return wtm.times[i].start.Before(wtm.times[j].start) })

	// Notify listeners that the queue is changed
	wtm.resetTimer <- true
	return true, nil
}

// GetNextSlot returns the next scheduled time.
// It's thread safe.
func (wtm *waterTimeManager) GetNextSlot() *waterTime {
	wtm.Lock()
	defer wtm.Unlock()

	if len(wtm.times) == 0 {
		return nil
	}

	i := 0
	// Find the next waterTime to use
	totalTimes := len(wtm.times)
	for ; i < totalTimes; i++ {
		if wtm.times[i].end.After(time.Now()) {
			break
		}
	}

	wtm.times = wtm.times[i:]
	if i == totalTimes { // No waterTime founds.
		return nil
	}

	return wtm.times[0]
}

// consumerSchedule manages the ticker for the system
func consumerSchedule(wtm *waterTimeManager, eventer gobot.Eventer, wg *sync.WaitGroup) {

	defer wg.Done()

	commands := eventer.Subscribe()
	quit := make(chan bool)

	// This routine will wait events from commands channel
	// and in case a stopWorkers with "true" value
	// will be received, we close the scheduler
	go func() {
		for e := range commands {
			if e.Name != stopWorkers {
				continue
			}
			if exitNow, ok := e.Data.(bool); !ok || exitNow {
				quit <- true
				return
			}
		}

	}()

	var timer *time.Timer

WAIT_FIRST_SLOT:
	for {
		select {

		case <-wtm.resetTimer: // Wait the first incoming waterTime.
		WAIT_SLOTS:
			for {
				nextSlot := wtm.GetNextSlot() // Get the next waterTime.
				if nextSlot == nil {
					// No more work to do, empty schedule,
					// go to start and wait the first waterTime incoming.
					continue WAIT_FIRST_SLOT
				}
				d := time.Until(nextSlot.start)
				// If a duration is negative, means that the slot received is currently active.
				// This happens when the scheduler has been reseted during an active task.

				if d > 0 {
					log.Printf("Next timer will start at: %v", d)
					timer = time.AfterFunc(d, func() {
						eventer.Publish(startRelay, struct{}{})
					})
				}

				select {
				case <-time.After(time.Until(nextSlot.end)): // Wait the end of the process.
					eventer.Publish(stopWorkers, false)
					continue WAIT_SLOTS
				case <-wtm.resetTimer: // Wait if the meanwhile the manager has been reset.
					if timer != nil {
						timer.Stop()
					}
					log.Println("reset timer!")
					continue WAIT_SLOTS
				case <-quit: // Quit signal. Exits
					if timer != nil {
						timer.Stop()
					}
					eventer.Unsubscribe(commands)
					log.Printf("close the schedule")
					return
				}
			}

		case <-quit: // Quit signal. Exits
			eventer.Unsubscribe(commands)
			log.Printf("close the schedule")
			return

		}
	}

}

type sumWaterTime struct {
	start     time.Time
	end       time.Time
	started   bool
	willStart time.Duration
}

func (wtm *waterTimeManager) PrintStatus() []*sumWaterTime {
	wtm.RLock()
	defer wtm.RUnlock()

	times := make([]*sumWaterTime, len(wtm.times))
	for i, t := range wtm.times {
		times[i] = &sumWaterTime{
			start:     t.start,
			end:       t.end,
			started:   !t.start.After(time.Now()),
			willStart: time.Until(t.start),
		}
	}
	return times
}
