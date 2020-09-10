package addons

import (
    "fmt"
    //"strconv"
	"time"

	//"github.com/derickr/streamdeck-goui/actionhandlers"
	"github.com/magicmonkey/go-streamdeck"
	"github.com/magicmonkey/go-streamdeck/buttons"
	// sddecorators "github.com/magicmonkey/go-streamdeck/decorators"
	//"github.com/rs/zerolog/log"
)

type Clock struct {
	SD         *streamdeck.StreamDeck
	buttonIndex int
	done        chan bool
	ticker     *time.Ticker
}

func (c *Clock) Init() {
	c.done = make(chan bool)
	c.buttonIndex = -1

	c.ticker = time.NewTicker(1 * time.Second)

	go func() {
		for {
			select {
			case <-c.done:
				return
			case t := <-c.ticker.C:
				if c.buttonIndex >= 0 {
					button := buttons.NewTextButton(fmt.Sprintf("%02d:%02d:%02d", t.Hour(), t.Minute(), t.Second()))
					c.SD.AddButton(c.buttonIndex, button)
				}
			}
		}
	}()
}

func (c *Clock) SetClockButton(offset int) {
    c.buttonIndex = offset
}

func (c *Clock) Reset() {
    c.buttonIndex = -1
}
