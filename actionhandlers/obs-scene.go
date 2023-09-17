package actionhandlers

import (
	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/scenes"
	streamdeck "github.com/magicmonkey/go-streamdeck"
	"github.com/rs/zerolog/log"
)

type OBSSceneAction struct {
	Client *goobs.Client
	Scene  string
	btn    streamdeck.Button
}

func (action *OBSSceneAction) Pressed(btn streamdeck.Button) {

	log.Info().Msg("Set scene: " + action.Scene)
	_, err := action.Client.Scenes.SetCurrentProgramScene(&scenes.SetCurrentProgramSceneParams{SceneName: action.Scene})
	if err != nil {
		log.Warn().Err(err).Msg("OBS scene action error")
	}

}
