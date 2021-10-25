package main

import (
	"context"
	_ "embed"
	"log"
	"time"

	"github.com/endocrimes/keylight-go"
	"github.com/getlantern/systray"
)

var (
	//go:embed icon/on.png
	iconOn []byte
	//go:embed icon/off.png
	iconOff []byte
)

func main() {
	systray.Run(onReady, nil)
}

func discoverLights() (<-chan *keylight.Device, error) {
	discovery, err := keylight.NewDiscovery()
	if err != nil {
		log.Println("failed to initialize keylight discovery: ", err.Error())
		return nil, err
	}

	go func() {
		err := discovery.Run(context.Background())
		if err != nil {
			log.Fatalln("Failed to discover lights: ", err.Error())
		}
	}()

	return discovery.ResultsCh(), nil
}

func isLightOn(state int) bool {
	if state == 0 {
		return true
	}
	return false
}

func togglePowerState(lg *keylight.LightGroup) *keylight.LightGroup {
	newLG := lg.Copy()
	for i, l := range lg.Lights {
		if isLightOn(l.On) {
			newLG.Lights[i].On = 1
		} else {
			newLG.Lights[i].On = 0
		}
	}
	return newLG
}

func writeDiscoverConfig() error {
	timeout := time.Duration(20 * time.Second)
	devicesCh, err := discoverLights()
	if err != nil {
		return err
	}

	count := 0

	var allLights []*keylight.Device
	dAllLights := systray.AddMenuItem("All Lights", "")
	dAllLights.Enable()
	systray.AddSeparator()

	go func(dAllLights *systray.MenuItem) {
		for {
			_, ok := <-dAllLights.ClickedCh
			if !ok {
				break
			}
			for _, d := range allLights {
				lg, _ := d.FetchLightGroup(context.TODO())
				d.UpdateLightGroup(context.TODO(), togglePowerState(lg))
			}

		}
	}(dAllLights)

	for {
		select {
		case device := <-devicesCh:
			if device == nil {
				return nil
			}

			allLights = append(allLights, device)

			dName := systray.AddMenuItem(device.Name, "")
			dName.Enable()
			go func(dName *systray.MenuItem) {
				for {
					_, ok := <-dName.ClickedCh
					if !ok {
						break
					}

					lg, _ := device.FetchLightGroup(context.TODO())
					device.UpdateLightGroup(context.TODO(), togglePowerState(lg))
				}
			}(dName)
			count++
		case <-time.After(timeout):
			return nil
		}
	}
}

func onReady() {
	systray.SetIcon(iconOff)

	writeDiscoverConfig()

	mExit := systray.AddMenuItem("Exit", "")
	go func() {
		<-mExit.ClickedCh
		systray.Quit()
	}()
}
