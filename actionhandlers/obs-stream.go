package actionhandlers

import (
	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/inputs"
	"github.com/andreykaipov/goobs/api/requests/stream"
	streamdeck "github.com/magicmonkey/go-streamdeck"
	"github.com/rs/zerolog/log"
)

type OBSStreamAction struct {
	Client *goobs.Client
	btn    streamdeck.Button
	Source string
}

func (action *OBSStreamAction) Pressed(btn streamdeck.Button) {

	log.Info().Msg("Stream!")
	_, err := action.Client.Stream.ToggleStream(&stream.ToggleStreamParams{})
	if err != nil {
		log.Warn().Err(err).Msg("OBS stream action error")
	}

	if action.Source == "" {
		return
	}

	_, err = action.Client.Inputs.SetInputMute(&inputs.SetInputMuteParams{InputMuted: &[]bool{true}[0], InputName: action.Source})
	if err != nil {
		log.Warn().Err(err).Msg("OBS stream action error")
	}
}
