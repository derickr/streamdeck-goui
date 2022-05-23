package addons

import (
	"fmt"
	"strconv"

	//"github.com/derickr/streamdeck-goui/actionhandlers"
	"github.com/magicmonkey/go-streamdeck"
	"github.com/magicmonkey/go-streamdeck/buttons"
	// sddecorators "github.com/magicmonkey/go-streamdeck/decorators"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Brightness struct {
	SD          *streamdeck.StreamDeck
	buttonIndex int
	Level       int
	levelLow    int
	levelNormal int
	levelHigh   int
	imageLow    string
	imageNormal string
	imageHigh   string
}

type BrightnessAction struct {
	Brightness *Brightness
}

func (a *BrightnessAction) Pressed(btn streamdeck.Button) {
	a.Brightness.Level = a.Brightness.Level + 1
	if a.Brightness.Level > 2 {
		a.Brightness.Level = 0
	}

	log.Debug().Msgf("Setting brightness to %d", a.Brightness.Level)
	a.Brightness.Update()
}

func (b *Brightness) Update() {
	var level int
	var image string

	if b.Level == 0 {
		level = b.levelLow
		image = b.imageLow
	} else if b.Level == 1 {
		level = b.levelNormal
		image = b.imageNormal
	} else {
		level = b.levelHigh
		image = b.imageHigh
	}

	b.SD.SetBrightness(level)

	if b.buttonIndex < 0 {
		return
	}

	buttonAction := &BrightnessAction{Brightness: b}
	button, err := buttons.NewImageFileButton(viper.GetString("buttons.images") + "/" + image)
	if err != nil {
		button := buttons.NewTextButton(fmt.Sprintf("L %d", b.Level))
		button.SetActionHandler(buttonAction)
		b.SD.AddButton(b.buttonIndex, button)
	} else {
		button.SetActionHandler(buttonAction)
		b.SD.AddButton(b.buttonIndex, button)
	}
}

func (b *Brightness) Init() {
	b.Level = 1
	b.buttonIndex = -1

	b.SD.SetBrightness(50)
}

func (b *Brightness) Buttons(offset int, arguments map[string]string) {
	b.buttonIndex = offset

	b.levelLow, _ = strconv.Atoi(arguments["LevelLow"])
	b.levelNormal, _ = strconv.Atoi(arguments["LevelNormal"])
	b.levelHigh, _ = strconv.Atoi(arguments["LevelHigh"])

	b.imageLow = arguments["ImageLow"]
	b.imageNormal = arguments["ImageNormal"]
	b.imageHigh = arguments["ImageHigh"]

	b.Update()
}
