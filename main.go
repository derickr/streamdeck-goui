package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync"

	"github.com/derickr/streamdeck-goui/addons"
	streamdeck "github.com/magicmonkey/go-streamdeck"
	"github.com/magicmonkey/go-streamdeck/buttons"
	_ "github.com/magicmonkey/go-streamdeck/devices"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var sd *streamdeck.StreamDeck
var obs_addon *addons.Obs
var brightness_addon *addons.Brightness
var clock_addon *addons.Clock

func loadConfigAndDefaults() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04"})

	// first set some default values
	viper.AddConfigPath(".")
	viper.SetDefault("buttons.images", "images/buttons") // location of button images
	viper.SetDefault("obs.host", "localhost")            // OBS webhooks endpoint
	viper.SetDefault("obs.port", 4444)                   // OBS webhooks endpoint
	viper.SetDefault("mqtt.uri", "tcp://10.1.0.1:1883")  // MQTT server location

	// now read in config for any overrides
	err := viper.ReadInConfig()
	if err != nil { // Handle errors reading the config file
		log.Warn().Msgf("Cannot read config file: %s \n", err)
	}
}

type CommandAction struct {
	Command string
}

func (c *CommandAction) Pressed(btn streamdeck.Button) {
	cmd := exec.Command(c.Command)
	log.Info().Msgf("Running: %s", c.Command);
	if err := cmd.Run(); err != nil {
		log.Warn().Err(err)
	}
}

type SdPage struct {
	Name     string `mapstructure:"name"`
	Image    string `mapstructure:"image"`
	Index    int    `mapstructure:"index"`
	Last     bool   `mapstructure:"last"`
}

type PageAction struct {
	Page string
}

type PageDefinition struct {
	Index int
	Type  string
	Arguments map[string] string
	Image string
	ImageOn string
	ImageOff string
}

func (action *PageAction) Pressed(btn streamdeck.Button) {
	log.Debug().Msg(action.Page)
	clock_addon.Reset()

	var page_definition []PageDefinition
	viper.UnmarshalKey(action.Page, &page_definition)

	// Clear all buttons
	for i := 0; i < viper.GetInt("button_count") - 1; i++ {
		sd.UnsetDecorator(i)
		textButton := buttons.NewTextButton("")
		sd.AddButton(i, textButton)
	}
	/* Reset OBS buttons */
	obs_addon.ClearButtons()

	for _, button := range page_definition {
		if button.Type == "pages" {
			setupPages(sd)
		}

		if button.Type == "obs_scenes" {
			maxScenes, _ := strconv.Atoi(button.Arguments["SceneCount"])
			obs_addon.Buttons(maxScenes, button.Index)
		}

		if button.Type == "obs_record" {
			imageFile := viper.GetString("buttons.images") + "/" + button.Image
			unmuteSource := button.Arguments["UnmuteSource"];
			obs_addon.SetRecordButton(button.Index, imageFile, unmuteSource)
		}

		if button.Type == "obs_stream" {
			imageFile := viper.GetString("buttons.images") + "/" + button.Image
			unmuteSource := button.Arguments["UnmuteSource"];
			obs_addon.SetStreamButton(button.Index, imageFile, unmuteSource)
		}

		if button.Type == "obs_toggle_mute" {
			imageFileOn := viper.GetString("buttons.images") + "/" + button.ImageOn
			imageFileOff := viper.GetString("buttons.images") + "/" + button.ImageOff
			obs_addon.SetToggleMuteButton(button.Index, button.Arguments["Source"], imageFileOn, imageFileOff)
		}

		if button.Type == "brightness" {
			brightness_addon.Buttons(button.Index, button.Arguments)
		}

		if button.Type == "home" {
			homeButton, _ := buttons.NewImageFileButton(viper.GetString("buttons.images") + "/" + button.Image)
			homeButton.SetActionHandler(&PageAction{Page: "Home"})
			sd.AddButton(button.Index, homeButton)
		}

		if button.Type == "command" {
			cmdButton, _ := buttons.NewImageFileButton(viper.GetString("buttons.images") + "/" + button.Image)
			cmdButton.SetActionHandler(&CommandAction{Command: button.Arguments["Command"]})
			sd.AddButton(button.Index, cmdButton)
		}

		if button.Type == "clock" {
			clock_addon.SetClockButton(button.Index)
		}
	}
}



func setupPages(sd *streamdeck.StreamDeck) {
	var sd_pages []SdPage
	viper.UnmarshalKey("pages", &sd_pages);

	i := 0;

	for _, page := range sd_pages {
		index := i
		if page.Last {
			index = viper.GetInt("button_count") - 1
		}
		if page.Index != 0 {
			index = page.Index
		}
		button, err := buttons.NewImageFileButton(viper.GetString("buttons.images") + "/" + page.Image)
		if err == nil {
			button.SetActionHandler(&PageAction{Page: page.Name})
			sd.AddButton(index, button)
		} else {
			button := buttons.NewTextButton(page.Name)
			button.SetActionHandler(&PageAction{Page: page.Name})
			sd.AddButton(index, button)
		}

		i = i + 1
	}
}

func main() {
	loadConfigAndDefaults()
	log.Info().Msg("Starting streamdeck tricks. Hai!")

	var err error
	sd, err = streamdeck.New()
	if err != nil {
		log.Error().Err(err).Msg("Error finding Streamdeck")
		panic(err)
	}

	obs_addon = &addons.Obs{SD: sd}
	obs_addon.ConnectOBS()
	obs_addon.ObsEventHandlers()

	brightness_addon = &addons.Brightness{SD: sd}
	brightness_addon.Init()

	clock_addon = &addons.Clock{SD: sd}
	clock_addon.Init()
/*
	// init Screenshot
	screenshot_addon := addons.Screenshot{SD: sd}
	screenshot_addon.Init()
	screenshot_addon.Buttons()

	// init WindowManager
	windowmgmt_addon := addons.WindowMgmt{SD: sd}
	windowmgmt_addon.Init()
	windowmgmt_addon.Buttons()

	// set up soundcaster
	caster_addon := addons.Caster{SD: sd}
	caster_addon.Init()
	caster_addon.Buttons()
*/
	action := &PageAction{Page: "Home"};
	action.Pressed(buttons.NewTextButton("TEST"))

	go webserver()

	log.Info().Msg("Up and running")
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}

func webserver() {
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	})

	http.ListenAndServe(":7001", nil)
}
