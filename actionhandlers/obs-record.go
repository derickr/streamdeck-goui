package actionhandlers

import (
	"github.com/christopher-dG/go-obs-websocket"
	streamdeck "github.com/magicmonkey/go-streamdeck"
	"github.com/rs/zerolog/log"
)

type OBSRecordAction struct {
	Client obsws.Client
	btn    streamdeck.Button
}

func (action *OBSRecordAction) Pressed(btn streamdeck.Button) {

	log.Info().Msg("Record!")
	req := obsws.NewStartStopRecordingRequest()
	_, err := req.SendReceive(action.Client)
	if err != nil {
		log.Warn().Err(err).Msg("OBS record action error")
	}
}
