package main

import (
	"flag"
	"github.com/gsapkal/hap"
	"github.com/gsapkal/hap/accessory"
	"github.com/gsapkal/hap/characteristic"
	"github.com/gsapkal/hap/log"
	"path/filepath"
	"strings"
	"time"
	"tinygo.org/x/bluetooth"

	"context"
	syslog "log"
	"os"
	"os/signal"
	"syscall"
)

var adapter = bluetooth.DefaultAdapter

func main() {

	var logLevel string
	flag.StringVar(&logLevel, "loglevel", "INFO", "Logging level")

	var devices string
	flag.StringVar(&devices, "devices", "Ganesh's iphone", "BLE device name to monitor")

	var thresholdInt int
	flag.IntVar(&thresholdInt, "threshold", -60, "BLE RSSI threshold to detect occupancy")

	threshold := int16(thresholdInt)

	var fsStorePath string
	flag.StringVar(&fsStorePath, "store", "", "File system data store")

	var bindAddr string
	flag.StringVar(&bindAddr, "bind", "192.168.0.1:54321", "Network bind address in case you have multiple nic")

	flag.Parse()

	if fsStorePath == "" {
		homedir, _  := os.UserHomeDir()
		fsStorePath = filepath.Join(homedir, "work", "homekit",  "officeOccupancy")
	}



	o := accessory.NewOccupancySensor(accessory.Info{
		Name:         "OfficeOccupancySensor",
		SerialNumber: "886",
		Manufacturer: "gsapkal",
		Model:        "hap",
		Firmware:     "1.0",

	})

	s, err := hap.NewServer(hap.NewFsStore(fsStorePath), o.A)
	s.Addr = bindAddr
	if err != nil {
		log.Info.Panic(err)
	}

	if logLevel == "DEBUG" {
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
		must("start occupancy accessory ", s.ListenAndServe(ctx))
	}()

	// Enable BLE interface.
	must("enable BLE stack", adapter.Enable())
	// Start scanning.
	log.Info.Println("Detecting occupancy based on device ...", devices)

	start := time.Now()

	err = adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {

		if len(device.LocalName()) > 0 && strings.Contains(devices, device.LocalName()) {
			log.Debug.Println("found device:", device.Address.String(), device.RSSI, device.LocalName())
			if device.RSSI > threshold {
				start = time.Now()
				o.OccupancySensor.OccupancyDetected.SetValue(characteristic.OccupancyDetectedOccupancyDetected)
			} else {
				//Avoid flicker in case signal is not stable
				if time.Since(start).Seconds() > 3 {
					o.OccupancySensor.OccupancyDetected.SetValue(characteristic.OccupancyDetectedOccupancyNotDetected)
				}
			}
		} else {
			duration := time.Since(start)
			if duration.Seconds() > 60 {
				log.Debug.Println("Missing :", devices)
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
