package actionhandlers

import (
	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/inputs"
	streamdeck "github.com/magicmonkey/go-streamdeck"
	"github.com/rs/zerolog/log"
)

type OBSToggleMuteAction struct {
	Client *goobs.Client
	btn    streamdeck.Button
	Source string
}

func (action *OBSToggleMuteAction) Pressed(btn streamdeck.Button) {

	_, err := action.Client.Inputs.ToggleInputMute(&inputs.ToggleInputMuteParams{InputName: action.Source})
	if err != nil {
		log.Warn().Err(err).Msg("OBS stream action error")
	}
}
