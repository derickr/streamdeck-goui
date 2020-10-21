package actionhandlers

import (
	"github.com/christopher-dG/go-obs-websocket"
	streamdeck "github.com/magicmonkey/go-streamdeck"
	"github.com/rs/zerolog/log"
)

type OBSStreamAction struct {
	Client obsws.Client
	btn    streamdeck.Button
	Source string
}

func (action *OBSStreamAction) Pressed(btn streamdeck.Button) {

	log.Info().Msg("Stream!")
	req := obsws.NewStartStopStreamingRequest()
	_, err := req.SendReceive(action.Client)
	if err != nil {
		log.Warn().Err(err).Msg("OBS stream action error")
	}

	if action.Source == "" {
		return
	}

	req2 := obsws.NewSetMuteRequest(action.Source, false)
	_, err = req2.SendReceive(action.Client)
	if err != nil {
		log.Warn().Err(err).Msg("OBS stream action error")
	}
}
