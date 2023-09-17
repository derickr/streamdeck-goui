package addons

import (
	"image/color"
	"strings"

	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/events"
	"github.com/andreykaipov/goobs/api/requests/inputs"
	"github.com/andreykaipov/goobs/api/requests/scenes"
	"github.com/derickr/streamdeck-goui/actionhandlers"
	"github.com/magicmonkey/go-streamdeck"
	"github.com/magicmonkey/go-streamdeck/buttons"
	sddecorators "github.com/magicmonkey/go-streamdeck/decorators"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Obs struct {
	SD                *streamdeck.StreamDeck
	obs_client        *goobs.Client
	recordButtonIndex int
	streamButtonIndex int
	toggleMuteIndex   int
	muteOnImage       string
	muteOffImage      string
	muteSource        string
}

var obs_current_scene string

type ObsScene struct {
	Name     string `mapstructure:"name"`
	Image    string `mapstructure:"image"`
	ButtonId int
}

func (scene *ObsScene) SetButtonId(id int) {
	scene.ButtonId = id
}

var buttons_obs map[string]*ObsScene // scene name and image name

func (o *Obs) ClearButtons() {
	o.recordButtonIndex = -1
	o.streamButtonIndex = -1
	o.toggleMuteIndex = -1
}

func (o *Obs) SetRecordButton(index int, image string, unmuteSource string) {
	if o.obs_client == nil {
		return
	}

	oaction := &actionhandlers.OBSRecordAction{Client: o.obs_client, Source: unmuteSource}

	recordButton, err := buttons.NewImageFileButton(image)
	if err == nil {
		recordButton.SetActionHandler(oaction)
		o.SD.AddButton(index, recordButton)
	} else {
		recordButton := buttons.NewTextButton("Record")
		recordButton.SetActionHandler(oaction)
		o.SD.AddButton(index, recordButton)
	}

	o.recordButtonIndex = index
}

func (o *Obs) SetStreamButton(index int, image string, unmuteSource string) {
	if o.obs_client == nil {
		return
	}

	oaction := &actionhandlers.OBSStreamAction{Client: o.obs_client, Source: unmuteSource}

	streamButton, err := buttons.NewImageFileButton(image)
	if err == nil {
		streamButton.SetActionHandler(oaction)
		o.SD.AddButton(index, streamButton)
	} else {
		streamButton := buttons.NewTextButton("Stream")
		streamButton.SetActionHandler(oaction)
		o.SD.AddButton(index, streamButton)
	}

	o.streamButtonIndex = index
}

func (o *Obs) getMuteCurrentStatus(source string) bool {
	log.Info().Msg("Fetching current mute status")
	result, err := o.obs_client.Inputs.GetInputMute(&inputs.GetInputMuteParams{InputName: source})

	if err != nil {
		log.Warn().Err(err).Msg("OBS stream action error")
	}

	return result.InputMuted
}

func (o *Obs) SetToggleMuteButton(index int, source string, imageOn string, imageOff string) {
	if o.obs_client == nil {
		return
	}

	oaction := &actionhandlers.OBSToggleMuteAction{Client: o.obs_client, Source: source}

	toggleMuteButton, err := buttons.NewImageFileButton(imageOn)
	if err == nil {
		toggleMuteButton.SetActionHandler(oaction)
		o.SD.AddButton(index, toggleMuteButton)
	} else {
		toggleMuteButton := buttons.NewTextButton("Mute/Unmute")
		toggleMuteButton.SetActionHandler(oaction)
		o.SD.AddButton(index, toggleMuteButton)
	}

	o.toggleMuteIndex = index
	o.muteOnImage = imageOn
	o.muteOffImage = imageOff
	o.muteSource = source

	o.updateMuteButton(o.getMuteCurrentStatus(source))
}

func (o *Obs) ConnectOBS() {
	log.Debug().Msg("Connecting to OBS...")
	log.Info().Msgf("%#v\n", viper.Get("obs.host"))
	tmpObsClient, err := goobs.New(
		viper.GetString("obs.host"),
		goobs.WithPassword(viper.GetString("obs.password")),
	)
	o.obs_client = tmpObsClient
	if err != nil {
		log.Warn().Err(err).Msg("Cannot connect to OBS")
	}
	o.ClearButtons()
}

func (o *Obs) ObsEventHandlers() {
	if o.obs_client == nil {
		return
	}

	log.Debug().Msg("Setting up handlers...")

	go o.obs_client.Listen(func(event interface{}) {
		switch e := event.(type) {
		case *events.CurrentProgramSceneChanged:
			// Make sure to assert the actual event type.
			scene := e.SceneName
			lScene := strings.ToLower(scene)

			log.Info().Msg("Old scene: " + obs_current_scene)
			// undecorate the old
			if oldb, ok := buttons_obs[obs_current_scene]; ok {
				log.Info().Int("button", oldb.ButtonId).Msg("Clear original button decoration")
				o.SD.UnsetDecorator(oldb.ButtonId)
			}
			// decorate the new
			log.Info().Msg("New scene: " + scene)
			if eventb, ok := buttons_obs[lScene]; ok {
				log.Info().Int("button", eventb.ButtonId).Msg("Highlight new scene button")
				decorator2 := sddecorators.NewBorder(4, color.RGBA{255, 255, 0, 255})
				o.SD.SetDecorator(eventb.ButtonId, decorator2)
			}
			obs_current_scene = lScene

		case *events.RecordStateChanged:
			if o.recordButtonIndex >= 0 {
				if e.OutputActive {
					decorator2 := sddecorators.NewBorder(8, color.RGBA{255, 0, 0, 255})
					log.Info().Msg("Recording Started")
					o.SD.SetDecorator(o.recordButtonIndex, decorator2)
				} else {
					log.Info().Msg("Recording Stopped")
					o.SD.UnsetDecorator(o.recordButtonIndex)
				}
			}

		case *events.StreamStateChanged:
			if o.streamButtonIndex >= 0 {
				if e.OutputActive {
					decorator2 := sddecorators.NewBorder(8, color.RGBA{0, 0, 255, 255})
					log.Info().Msg("Stream Started")
					o.SD.SetDecorator(o.streamButtonIndex, decorator2)
				} else {
					log.Info().Msg("Stream Stopped")
					o.SD.UnsetDecorator(o.streamButtonIndex)
				}
			}

		case *events.InputMuteStateChanged:
			if o.toggleMuteIndex < 0 {
				return
			}
			o.updateMuteButton(e.InputMuted)
		}
	})
}

func (o *Obs) updateMuteButton(muted bool) {
	var text string
	var image string
	var decorator *sddecorators.Border

	if muted {
		image = o.muteOffImage

		decorator = sddecorators.NewBorder(8, color.RGBA{255, 0, 0, 255})
		text = " Muted "
	} else {
		image = o.muteOnImage

		decorator = sddecorators.NewBorder(8, color.RGBA{0, 255, 0, 255})

		text = " Unmuted "
	}

	log.Info().Msg(text)

	oaction := &actionhandlers.OBSToggleMuteAction{Client: o.obs_client, Source: o.muteSource}

	toggleMuteButton, err := buttons.NewImageFileButton(image)
	if err == nil {
		toggleMuteButton.SetActionHandler(oaction)
		o.SD.AddButton(o.toggleMuteIndex, toggleMuteButton)
	} else {
		toggleMuteButton := buttons.NewTextButton(text)
		toggleMuteButton.SetActionHandler(oaction)
		o.SD.SetDecorator(o.toggleMuteIndex, decorator)
		o.SD.AddButton(o.toggleMuteIndex, toggleMuteButton)
	}
}

func (o *Obs) Buttons(maxScenes int, offset int) {
	o.ConnectOBS()

	if o.obs_client == nil {
		return
	}

	o.ObsEventHandlers()

	// OBS Scenes to Buttons
	buttons_obs = make(map[string]*ObsScene)
	viper.UnmarshalKey("obs_scenes", &buttons_obs)

	// offset for what number button to start at
	image_path := viper.GetString("buttons.images")
	var image string

	// what scenes do we have? (max 8 for the top row of buttons)
	scenes, err := o.obs_client.Scenes.GetSceneList(&scenes.GetSceneListParams{})
	if err != nil {
		log.Warn().Err(err)
	}

	// what is the current schene?
	obs_current_scene = strings.ToLower(scenes.CurrentProgramSceneName)

	// make buttons for these scenes
	log.Debug().Msgf("Max Scenes: %d", maxScenes)
	for j := len(scenes.Scenes) - 1; j >= 0; j-- {
		scene := scenes.Scenes[j]
		i := len(scenes.Scenes) - 1 - j
		log.Debug().Msg("Scene: " + scene.SceneName)
		image = ""
		sceneName := scene.SceneName
		lSceneName := strings.ToLower(sceneName)
		// only need a few scenes
		if i >= maxScenes {
			continue
		}

		oaction := &actionhandlers.OBSSceneAction{Scene: scene.SceneName, Client: o.obs_client}

		if s, ok := buttons_obs[lSceneName]; ok {
			if s.Image != "" {
				image = image_path + "/" + s.Image
			}
		} else {
			// there wasn't an entry in the buttons for this scene so add one
			buttons_obs[lSceneName] = &ObsScene{}
		}

		if image != "" {
			// try to make an image button

			obutton, err := buttons.NewImageFileButton(image)
			if err == nil {
				obutton.SetActionHandler(oaction)
				o.SD.AddButton(i+offset, obutton)
				// store which button we just set
				buttons_obs[lSceneName].SetButtonId(i + offset)
			} else {
				// something went wrong with the image, use a default one
				image = image_path + "/play.jpg"
				obutton, err := buttons.NewImageFileButton(image)
				if err == nil {
					obutton.SetActionHandler(oaction)
					o.SD.AddButton(i+offset, obutton)
					// store which button we just set
					buttons_obs[lSceneName].SetButtonId(i + offset)
				}
			}
		} else {
			// use a text button
			oopbutton := buttons.NewTextButton(scene.SceneName)
			oopbutton.SetActionHandler(oaction)
			o.SD.AddButton(i+offset, oopbutton)
			// store which button we just set
			buttons_obs[lSceneName].SetButtonId(i + offset)
		}
	}

	// highlight the active scene
	if eventb, ok := buttons_obs[obs_current_scene]; ok {
		decorator2 := sddecorators.NewBorder(5, color.RGBA{255, 255, 0, 255})
		log.Info().Int("button", eventb.ButtonId).Msg("Highlight current scene")
		o.SD.SetDecorator(eventb.ButtonId, decorator2)
	}
}
