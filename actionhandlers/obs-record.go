package actionhandlers

import (
	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/inputs"
	"github.com/andreykaipov/goobs/api/requests/record"
	streamdeck "github.com/magicmonkey/go-streamdeck"
	"github.com/rs/zerolog/log"
)

type OBSRecordAction struct {
	Client *goobs.Client
	btn    streamdeck.Button
	Source string
}

func (action *OBSRecordAction) Pressed(btn streamdeck.Button) {

	log.Info().Msg("Record!")
	_, err := action.Client.Record.ToggleRecord(&record.ToggleRecordParams{})
	if err != nil {
		log.Warn().Err(err).Msg("OBS record action error")
	}

	if action.Source == "" {
		return
	}

	_, err = action.Client.Inputs.SetInputMute(&inputs.SetInputMuteParams{InputMuted: &[]bool{true}[0], InputName: action.Source})
	if err != nil {
		log.Warn().Err(err).Msg("OBS stream action error")
	}
}
