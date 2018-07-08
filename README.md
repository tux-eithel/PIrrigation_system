# Raspberry Pi experiments

## Libraries

Using libraries:
 - [gobot.io](https://gobot.io/)
 - [periph.io](https://periph.io/)

To use gobot with the current release of periph I've modify the file
`gobot.io/x/gobot/drivers/spi/spi.go` line 83 from:
```
c, err := p.Connect(maxSpeed, xspi.Mode(mode), bits)
```
to:
```
c, err := p.Connect(physic.Frequency(maxSpeed), xspi.Mode(mode), bits)
```

## Full project

The main goal is a *irrigation system* with:
- relay to start the pump
- some analog sensor to read the "water"
- a time scheduler
- some electric valve


## Current Status

Right now it has been implemented:
- relay to start the pump
- some analog sensor to read the "water" (type of sensor: TBD)
- a time schedule (http in future)
