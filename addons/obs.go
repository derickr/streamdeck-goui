package addons

import (
	"image/color"
	"strings"

	obsws "github.com/christopher-dG/go-obs-websocket"
	"github.com/derickr/streamdeck-goui/actionhandlers"
	"github.com/magicmonkey/go-streamdeck"
	"github.com/magicmonkey/go-streamdeck/buttons"
	sddecorators "github.com/magicmonkey/go-streamdeck/decorators"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Obs struct {
	SD         *streamdeck.StreamDeck
	obs_client obsws.Client
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
    o.toggleMuteIndex   = -1
}

func (o *Obs) SetRecordButton(index int, image string, unmuteSource string) {
    if o.obs_client.Connected() == false {
        o.ConnectOBS()
        o.ObsEventHandlers()
    }

	if o.obs_client.Connected() == false {
        return;
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
    if o.obs_client.Connected() == false {
        o.ConnectOBS()
        o.ObsEventHandlers()
    }

	if o.obs_client.Connected() == false {
        return;
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
    req := obsws.NewGetMuteRequest(source)

    result, err := req.SendReceive(o.obs_client)
	if err != nil {
		log.Warn().Err(err).Msg("OBS stream action error")
	}

    return result.Muted
}

func (o *Obs) SetToggleMuteButton(index int, source string, imageOn string, imageOff string) {
    if o.obs_client.Connected() == false {
        o.ConnectOBS()
        o.ObsEventHandlers()
    }

	if o.obs_client.Connected() == false {
        return;
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
	o.obs_client = obsws.Client{
        Host: viper.GetString("obs.host"),
        Port: viper.GetInt("obs.port"),
        Password: viper.GetString("obs.password"),
    }
	err := o.obs_client.Connect()
	if err != nil {
		log.Warn().Err(err).Msg("Cannot connect to OBS")
	}
    o.ClearButtons()
}

func (o *Obs) ObsEventHandlers() {
    if o.obs_client.Connected() == false {
        return;
    }

    log.Debug().Msg("Setting up handlers...")

    o.obs_client.AddEventHandler("SwitchScenes", func(e obsws.Event) {
        // Make sure to assert the actual event type.
        scene := strings.ToLower(e.(obsws.SwitchScenesEvent).SceneName)
        log.Info().Msg("Old scene: " + obs_current_scene)
        // undecorate the old
        if oldb, ok := buttons_obs[obs_current_scene]; ok {
            log.Info().Int("button", oldb.ButtonId).Msg("Clear original button decoration")
            o.SD.UnsetDecorator(oldb.ButtonId)
        }
        // decorate the new
        log.Info().Msg("New scene: " + scene)
        if eventb, ok := buttons_obs[scene]; ok {
            log.Info().Int("button", eventb.ButtonId).Msg("Highlight new scene button")
            decorator2 := sddecorators.NewBorder(4, color.RGBA{255, 255, 0, 255})
            o.SD.SetDecorator(eventb.ButtonId, decorator2)
        }
        obs_current_scene = scene
    })

    o.obs_client.AddEventHandler("RecordingStarted", func(e obsws.Event) {
        if o.recordButtonIndex >= 0 {
            decorator2 := sddecorators.NewBorder(8, color.RGBA{255, 0, 0, 255})
            log.Info().Msg("Recording Started")
            o.SD.SetDecorator(o.recordButtonIndex, decorator2)
        }
    })

    o.obs_client.AddEventHandler("RecordingStopped", func(e obsws.Event) {
        if o.recordButtonIndex >= 0 {
            log.Info().Msg("Recording Stopped")
            o.SD.UnsetDecorator(o.recordButtonIndex)
        }
    })

    o.obs_client.AddEventHandler("StreamStarted", func(e obsws.Event) {
        if o.streamButtonIndex >= 0 {
            decorator2 := sddecorators.NewBorder(8, color.RGBA{0, 0, 255, 255})
            log.Info().Msg("Stream Started")
            o.SD.SetDecorator(o.streamButtonIndex, decorator2)
        }
    })
    o.obs_client.AddEventHandler("StreamStopped", func(e obsws.Event) {
        if o.streamButtonIndex >= 0 {
            log.Info().Msg("Stream Stopped")
            o.SD.UnsetDecorator(o.streamButtonIndex)
        }
    })

    o.obs_client.AddEventHandler("SourceMuteStateChanged", func(e obsws.Event) {
        if o.toggleMuteIndex < 0 {
            return
        }

        o.updateMuteButton(e.(obsws.SourceMuteStateChangedEvent).Muted)
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
    if o.obs_client.Connected() == false {
        o.ConnectOBS()
        o.ObsEventHandlers()
    }

    if o.obs_client.Connected() == false {
        return
    }

	// OBS Scenes to Buttons
	buttons_obs = make(map[string]*ObsScene)
	viper.UnmarshalKey("obs_scenes", &buttons_obs)

    // offset for what number button to start at
    image_path := viper.GetString("buttons.images")
    var image string

    // what scenes do we have? (max 8 for the top row of buttons)
    scene_req := obsws.NewGetSceneListRequest()
    scenes, err := scene_req.SendReceive(o.obs_client)
    if err != nil {
        log.Warn().Err(err)
    }
    obs_current_scene = strings.ToLower(scenes.CurrentScene)

    // make buttons for these scenes
    log.Debug().Msgf("Max Scenes: %d", maxScenes)
    for i, scene := range scenes.Scenes {
        // only need a few scenes
        if i >= maxScenes {
            break
        }

        log.Debug().Msg("Scene: " + scene.Name)
        image = ""
        oaction := &actionhandlers.OBSSceneAction{Scene: scene.Name, Client: o.obs_client}
        sceneName := strings.ToLower(scene.Name)

        if s, ok := buttons_obs[sceneName]; ok {
            if s.Image != "" {
                image = image_path + "/" + s.Image
            }
        } else {
            // there wasn't an entry in the buttons for this scene so add one
            buttons_obs[sceneName] = &ObsScene{}
        }

        if image != "" {
            // try to make an image button

            obutton, err := buttons.NewImageFileButton(image)
            if err == nil {
                obutton.SetActionHandler(oaction)
                o.SD.AddButton(i+offset, obutton)
                // store which button we just set
                buttons_obs[sceneName].SetButtonId(i + offset)
            } else {
                // something went wrong with the image, use a default one
                image = image_path + "/play.jpg"
                obutton, err := buttons.NewImageFileButton(image)
                if err == nil {
                    obutton.SetActionHandler(oaction)
                    o.SD.AddButton(i+offset, obutton)
                    // store which button we just set
                    buttons_obs[sceneName].SetButtonId(i + offset)
                }
            }
        } else {
            // use a text button
            oopbutton := buttons.NewTextButton(scene.Name)
            oopbutton.SetActionHandler(oaction)
            o.SD.AddButton(i+offset, oopbutton)
            // store which button we just set
            buttons_obs[sceneName].SetButtonId(i + offset)
        }
    }

    // highlight the active scene
    if eventb, ok := buttons_obs[obs_current_scene]; ok {
        decorator2 := sddecorators.NewBorder(5, color.RGBA{255, 255, 0, 255})
        log.Info().Int("button", eventb.ButtonId).Msg("Highlight current scene")
        o.SD.SetDecorator(eventb.ButtonId, decorator2)
    }
}
