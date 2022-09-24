package main

import (
	"github.com/gsapkal/hap"
	"github.com/gsapkal/hap/accessory"
	"github.com/gsapkal/hap/characteristic"
	"github.com/gsapkal/hap/log"
	"strconv"
	"strings"
	"tinygo.org/x/bluetooth"

	"context"
	syslog "log"
	"os"
	"os/signal"
	"syscall"
)

var adapter = bluetooth.DefaultAdapter

func main() {

	i, err := strconv.ParseInt(getenv("THRESHOLD", "-60"), 10, 32)
	if err != nil {
		panic(err)
	}
	threshold := int16(i)

	o := accessory.NewOccupancySensor(accessory.Info{
		Name:         "OfficeOccupancySensor",
		SerialNumber: "886",
		Manufacturer: "gsapkal",
		Model:        "hap",
		Firmware:     "1.0",

	})

	s, err := hap.NewServer(hap.NewFsStore(getenv("STORE","./homekit/officeOccupancy")), o.A)
	if err != nil {
		log.Info.Panic(err)
	}

	if getenv("LOGLEVEL", "INFO") == "DEBUG" {
		mylogger := syslog.New(os.Stdout, "HAP ", syslog.LstdFlags|syslog.Lshortfile)
		log.Debug = &log.Logger{Logger: mylogger}
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-c
		signal.Stop(c) // stop delivering signals
		cancel()
	}()

	go func() {
		must("start occupancy accessary ", s.ListenAndServe(ctx))
	}()

	// Enable BLE interface.
	must("enable BLE stack", adapter.Enable())
	// Start scanning.
	println("scanning...")
	err = adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
		//		println("found device:", device.Address.String(), device.RSSI, device.LocalName())

		if strings.Contains(device.LocalName() , getenv("DEVICES", "My iphone")) {
			log.Debug.Println("found device:", device.Address.String(), device.RSSI, device.LocalName())
			if device.RSSI > threshold {
				o.OccupancySensor.OccupancyDetected.SetValue(characteristic.OccupancyDetectedOccupancyDetected)
			} else {
				o.OccupancySensor.OccupancyDetected.SetValue(characteristic.OccupancyDetectedOccupancyNotDetected)
			}

		}
	})
	must("start scan", err)

}

func must(action string, err error) {
	if err != nil {
		panic("failed to " + action + ": " + err.Error())
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}
