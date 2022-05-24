package addons

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/derickr/go-litra-driver"
	"github.com/magicmonkey/go-streamdeck"
	"github.com/magicmonkey/go-streamdeck/buttons"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type litraConfiguration struct {
	OnOff       bool
	Temperature int16
	Brightness  int
}

type Litra struct {
	SD             *streamdeck.StreamDeck
	buttonIndex    int
	litraDevice    *litra.LitraDevice
	CurrentConfig  int
	Configurations []litraConfiguration
}

type LitraAction struct {
	Litra *Litra
}

func (a *LitraAction) Pressed(btn streamdeck.Button) {
	a.Litra.CurrentConfig = a.Litra.CurrentConfig + 1
	if a.Litra.CurrentConfig >= len(a.Litra.Configurations) {
		a.Litra.CurrentConfig = 0
	}

	log.Debug().Msgf("Setting Litra Configuration to %d of %d", a.Litra.CurrentConfig+1, len(a.Litra.Configurations))
	a.Litra.Update()
}

func (l *Litra) Update() {
	cc := l.Configurations[l.CurrentConfig]

	if cc.OnOff {
		l.litraDevice.TurnOn()
		l.litraDevice.SetBrightness(cc.Brightness)
		l.litraDevice.SetTemperature(cc.Temperature)
	} else {
		l.litraDevice.TurnOff()
	}

	buttonAction := &LitraAction{Litra: l}
	imgFile := viper.GetString("buttons.images") +
		"/litra-config-" +
		fmt.Sprintf("%d", l.CurrentConfig) +
		".png"

	fmt.Printf("%v\n", imgFile)

	button, err := buttons.NewImageFileButton(imgFile)
	if err != nil {
		button := buttons.NewTextButton(fmt.Sprintf("L %d", l.CurrentConfig))
		button.SetActionHandler(buttonAction)
		l.SD.AddButton(l.buttonIndex, button)
	} else {
		button.SetActionHandler(buttonAction)
		l.SD.AddButton(l.buttonIndex, button)
	}
}

func (l *Litra) Init() {
	l.CurrentConfig = 1
	l.buttonIndex = -1

	dev, _ := litra.New()
	l.litraDevice = dev
	l.litraDevice.TurnOn()
}

func (l *Litra) Buttons(offset int, configurations map[string]string) {
	l.buttonIndex = offset

	l.Configurations = make([]litraConfiguration, len(configurations)+1)
	l.Configurations[0] = litraConfiguration{false, 0, 0}

	for i := 0; i < len(configurations); i++ {
		parts := strings.Split(configurations[fmt.Sprintf("%d", i)], ",")

		temp, _ := strconv.Atoi(parts[0])
		brightness, _ := strconv.Atoi(parts[1])

		l.Configurations[i+1] = litraConfiguration{true, int16(temp), brightness}
	}

	l.Update()
}
