package main

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func Test_newWaterTime(t *testing.T) {
	type args struct {
		row string
	}

	loc, _ := time.LoadLocation("Europe/Berlin")
	firstTimeStr := "2018-08-21 10:49:00"
	firsTime, _ := time.ParseInLocation(parseTimeConst, firstTimeStr, loc)
	secondTimeStr := "2018-08-21 10:49:10"
	secondTime, _ := time.ParseInLocation(parseTimeConst, secondTimeStr, loc)

	tests := []struct {
		name    string
		args    args
		want    *waterTime
		wantErr bool
	}{
		{
			"normal",
			args{fmt.Sprintf("%s - %s", firstTimeStr, secondTimeStr)},
			&waterTime{firsTime, secondTime},
			false,
		},
		{
			"missing args",
			args{fmt.Sprintf("%s -", firstTimeStr)},
			&waterTime{firsTime, secondTime},
			true,
		},
		{
			"wrong start",
			args{fmt.Sprintf("%s - %s", "2018-08-21 10:49", secondTimeStr)},
			nil,
			true,
		},
		{
			"wrong end",
			args{fmt.Sprintf("%s - %s", firstTimeStr, "2018-08-21 10:49")},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newWaterTime(tt.args.row)
			if (err != nil) != tt.wantErr {
				t.Errorf("unable to parse the date = %v", err)
				return
			}
			if tt.wantErr && got != nil {
				t.Errorf("want nil with error = %v", err)
				return
			}
			if tt.wantErr && got == nil {
				return
			}
			if got.start.String() != tt.want.start.String() {
				t.Errorf("start value = %s, want %s", got.start, tt.want.start)
			}
			if got.end.String() != tt.want.end.String() {
				t.Errorf("end value = %s, want %s", got.end, tt.want.end)
			}
		})
	}
}

func Test_newWaterTimeManager(t *testing.T) {

	got := newWaterTimeManager()
	if got == nil {
		t.Errorf("bad type = nil, want *waterTimeManager")
		return
	}
	if len(got.times) != 0 {
		t.Errorf("wrong dimension array, got %d, want %d", len(got.times), 0)
		return
	}

	// some basic test about resetTimer chan
	go func() {
		<-got.resetTimer
	}()
	select {
	case got.resetTimer <- true:
	case <-time.After(1 * time.Second):
		t.Errorf("unable to send to got.resetTimer chanel")
		return
	}

	// Try to append some value to got.times
	got.times = append(got.times, &waterTime{})
	if len(got.times) != 1 {
		t.Errorf("unable to append waterTime types to got.times")
	}

}

func Test_waterTimeManager_Append(t *testing.T) {

	//loc, _ := time.LoadLocation("Europe/Berlin")
	//firstTimeStr := "2018-08-21 10:49:00"
	//firsTime, _ := time.ParseInLocation(parseTimeConst, firstTimeStr, loc)
	//secondTimeStr := "2018-08-21 10:49:10"
	//secondTime, _ := time.ParseInLocation(parseTimeConst, secondTimeStr, loc)

	// Prepare a time starting from now
	now := time.Now()
	// First Slot: s1.
	tnow1Minute := now.Add(1 * time.Minute)
	tnow2Minute := now.Add(2 * time.Minute)

	// Second Slot: s2.
	tnow4Minute := now.Add(4 * time.Minute)
	tnow5Minute := now.Add(5 * time.Minute)

	// Third Slot: s3.
	tnow6Minute := now.Add(6 * time.Minute)
	tnow8Minute := now.Add(8 * time.Minute)

	tests := []struct {
		name      string
		args      []*waterTime
		order     []*waterTime
		want      bool
		wantErr   bool
		waitReset bool
	}{
		{name: "empty and before now", args: []*waterTime{&waterTime{}}, order: nil, wantErr: true},
		{name: "end before start", args: []*waterTime{&waterTime{tnow2Minute, tnow1Minute}}, order: nil, wantErr: true},
		{name: "start match end", args: []*waterTime{&waterTime{tnow1Minute, tnow4Minute}, &waterTime{tnow4Minute, tnow5Minute}}, order: []*waterTime{&waterTime{tnow1Minute, tnow4Minute}, &waterTime{tnow4Minute, tnow5Minute}}, want: true},
		{name: "start in between", args: []*waterTime{&waterTime{tnow1Minute, tnow4Minute}, &waterTime{tnow2Minute, tnow5Minute}}, order: []*waterTime{&waterTime{tnow1Minute, tnow4Minute}}, wantErr: true},
		{name: "end in between", args: []*waterTime{&waterTime{tnow2Minute, tnow5Minute}, &waterTime{tnow1Minute, tnow4Minute}}, order: []*waterTime{&waterTime{tnow2Minute, tnow5Minute}}, wantErr: true},
		{name: "completely inside", args: []*waterTime{&waterTime{tnow1Minute, tnow8Minute}, &waterTime{tnow4Minute, tnow5Minute}}, order: []*waterTime{&waterTime{tnow1Minute, tnow8Minute}}, wantErr: true},
		{name: "completely outside", args: []*waterTime{&waterTime{tnow4Minute, tnow5Minute}, &waterTime{tnow1Minute, tnow8Minute}}, order: []*waterTime{&waterTime{tnow4Minute, tnow5Minute}}, wantErr: true},
		{name: "start in between with three times", args: []*waterTime{&waterTime{tnow1Minute, tnow4Minute}, &waterTime{tnow6Minute, tnow8Minute}, &waterTime{tnow2Minute, tnow5Minute}}, order: []*waterTime{&waterTime{tnow1Minute, tnow4Minute}, &waterTime{tnow6Minute, tnow8Minute}}, wantErr: true},
		{name: "end in between with three times", args: []*waterTime{&waterTime{tnow1Minute, tnow2Minute}, &waterTime{tnow5Minute, tnow8Minute}, &waterTime{tnow4Minute, tnow6Minute}}, order: []*waterTime{&waterTime{tnow1Minute, tnow2Minute}, &waterTime{tnow5Minute, tnow8Minute}}, wantErr: true},
		{name: "start in a range, end in other range", args: []*waterTime{&waterTime{tnow1Minute, tnow4Minute}, &waterTime{tnow5Minute, tnow8Minute}, &waterTime{tnow2Minute, tnow6Minute}}, order: []*waterTime{&waterTime{tnow1Minute, tnow4Minute}, &waterTime{tnow5Minute, tnow8Minute}}, wantErr: true},
		{name: "simple", args: []*waterTime{&waterTime{tnow1Minute, tnow2Minute}}, order: []*waterTime{&waterTime{tnow1Minute, tnow2Minute}}, want: true},
		{name: "order times", args: []*waterTime{&waterTime{tnow4Minute, tnow5Minute}, &waterTime{tnow1Minute, tnow2Minute}}, order: []*waterTime{&waterTime{tnow1Minute, tnow2Minute}, &waterTime{tnow4Minute, tnow5Minute}}, want: true},
		{name: "three times: order s1, s2, s3", args: []*waterTime{&waterTime{tnow1Minute, tnow2Minute}, &waterTime{tnow4Minute, tnow5Minute}, &waterTime{tnow6Minute, tnow8Minute}}, order: []*waterTime{&waterTime{tnow1Minute, tnow2Minute}, &waterTime{tnow4Minute, tnow5Minute}, &waterTime{tnow6Minute, tnow8Minute}}, want: true},
		{name: "three times: order s1, s3, s2", args: []*waterTime{&waterTime{tnow1Minute, tnow2Minute}, &waterTime{tnow6Minute, tnow8Minute}, &waterTime{tnow4Minute, tnow5Minute}}, order: []*waterTime{&waterTime{tnow1Minute, tnow2Minute}, &waterTime{tnow4Minute, tnow5Minute}, &waterTime{tnow6Minute, tnow8Minute}}, want: true},
		{name: "three times: order s3, s1, s2", args: []*waterTime{&waterTime{tnow6Minute, tnow8Minute}, &waterTime{tnow1Minute, tnow2Minute}, &waterTime{tnow4Minute, tnow5Minute}}, order: []*waterTime{&waterTime{tnow1Minute, tnow2Minute}, &waterTime{tnow4Minute, tnow5Minute}, &waterTime{tnow6Minute, tnow8Minute}}, want: true},
		{name: "three times: order s3, s2, s1", args: []*waterTime{&waterTime{tnow6Minute, tnow8Minute}, &waterTime{tnow4Minute, tnow5Minute}, &waterTime{tnow1Minute, tnow2Minute}}, order: []*waterTime{&waterTime{tnow1Minute, tnow2Minute}, &waterTime{tnow4Minute, tnow5Minute}, &waterTime{tnow6Minute, tnow8Minute}}, want: true},
		{name: "three times: order s2, s1, s3", args: []*waterTime{&waterTime{tnow4Minute, tnow5Minute}, &waterTime{tnow1Minute, tnow2Minute}, &waterTime{tnow6Minute, tnow8Minute}}, order: []*waterTime{&waterTime{tnow1Minute, tnow2Minute}, &waterTime{tnow4Minute, tnow5Minute}, &waterTime{tnow6Minute, tnow8Minute}}, want: true},
		{name: "three times: order s2, s3, s1", args: []*waterTime{&waterTime{tnow4Minute, tnow5Minute}, &waterTime{tnow6Minute, tnow8Minute}, &waterTime{tnow1Minute, tnow2Minute}}, order: []*waterTime{&waterTime{tnow1Minute, tnow2Minute}, &waterTime{tnow4Minute, tnow5Minute}, &waterTime{tnow6Minute, tnow8Minute}}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wtm := newWaterTimeManager()

			// chanErr := make(chan error, 1)
			// if tt.waitReset {
			// 	go func() {
			// 		select {
			// 		case <-wtm.resetTimer:
			// 			chanErr <- nil
			// 		case <-time.After(2 * time.Second):
			// 			chanErr <- fmt.Errorf("waiting the signal from resetTimer")
			// 		}
			// 	}()
			// }

			// Catch all the signal we receive form wtm.resetTimer.
			go func() {
				for range wtm.resetTimer {

				}
			}()

			var err error
			var got bool
			for _, tm := range tt.args {
				got, err = wtm.Append(tm) // Try to append to the Water Time Manager.
			}

			// Check if we want some kind of error.
			if (err != nil) != tt.wantErr {
				t.Errorf("append want result %v; got %v with error %v", !tt.wantErr, got, err)
				return
			}

			// Chech the status.
			if got != tt.want {
				t.Errorf("status expected %v, got %v", tt.want, got)
				return
			}

			// Check the length.
			if len(wtm.times) != len(tt.order) {
				t.Errorf("len expected %d, got %d", len(tt.order), len(wtm.times))
				return
			}

			// Check the order of the elements.
			if tt.order != nil {
				for i := range tt.order {
					if ok := reflect.DeepEqual(tt.order[i], wtm.times[i]); !ok {
						t.Errorf("wrong order at position %d, want %v; got %v", i, tt.order[i], wtm.times[i])
						return
					}
				}
			}

			// if tt.waitReset {
			// 	err = <-chanErr
			// 	if err != nil {
			// 		t.Fatalf("%v", err)
			// 	}
			// }

		})
	}
}
