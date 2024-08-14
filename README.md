# Homekit Occupancy Detector 
Homekit accessory to detect occupancy using BLE devices.

Implemented using -
- [Fork](https://github.com/gsapkal/hap) of [HAP golang implementation by brutella](https://github.com/brutella/hap)
- BLE Scanner implementation using [bluetooth](https://github.com/tinygo-org/bluetooth)

This accessory scans all the nearby BLE devices and filters the device by name. When the specific device is found it updates the occupancy state based on  RSSI threshold ( default -60).


Usage :
```
go build ./cmd/occupancy 
./occupancy --bind {LocalIP}:{Port} --store {accessory file store path} --device {Device Name}
```

The default pin for this accessory is `00102003`

Update 
