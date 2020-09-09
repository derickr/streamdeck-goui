package actionhandlers

import (
	"github.com/christopher-dG/go-obs-websocket"
	streamdeck "github.com/magicmonkey/go-streamdeck"
	"github.com/rs/zerolog/log"
)

type OBSStreamAction struct {
	Client obsws.Client
	btn    streamdeck.Button
}

func (action *OBSStreamAction) Pressed(btn streamdeck.Button) {

	log.Info().Msg("Stream!")
	req := obsws.NewStartStopStreamingRequest()
	_, err := req.SendReceive(action.Client)
	if err != nil {
		log.Warn().Err(err).Msg("OBS stream action error")
	}
}
