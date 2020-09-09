package actionhandlers

import (
	"github.com/christopher-dG/go-obs-websocket"
	streamdeck "github.com/magicmonkey/go-streamdeck"
	"github.com/rs/zerolog/log"
)

type OBSToggleMuteAction struct {
	Client obsws.Client
	btn    streamdeck.Button
    Source string
}

func (action *OBSToggleMuteAction) Pressed(btn streamdeck.Button) {

	log.Info().Msg("ToggleMute!")
	req := obsws.NewToggleMuteRequest(action.Source)
	_, err := req.SendReceive(action.Client)
	if err != nil {
		log.Warn().Err(err).Msg("OBS stream action error")
	}
}
