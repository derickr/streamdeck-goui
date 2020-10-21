package addons

import (
	"image/color"
    "fmt"
    //"strconv"
	"time"

	//"github.com/derickr/streamdeck-goui/actionhandlers"
	"github.com/magicmonkey/go-streamdeck"
	"github.com/magicmonkey/go-streamdeck/buttons"
	//sddecorators "github.com/magicmonkey/go-streamdeck/decorators"
	//"github.com/rs/zerolog/log"
)

type TimerAction struct {
    StartTime  time.Time
	Clock     *Clock
}

func (t *TimerAction) Pressed(btn streamdeck.Button) {
	if t.Clock.TimerActive {
		t.Clock.TimerActive = false
		return;
	}

	t.Clock.StartTime = t.StartTime
	t.Clock.TimerActive = true
}

type Clock struct {
	SD         *streamdeck.StreamDeck
	ButtonIndex int
	done        chan bool
	ticker     *time.Ticker
	TimerActive bool
	StartTime   time.Time
}

func (c *Clock) Init() {
	c.done = make(chan bool)
	c.ButtonIndex = -1

	c.ticker = time.NewTicker(100 * time.Millisecond)

	go func() {
		for {
			select {
			case <-c.done:
				return
			case t := <-c.ticker.C:
				if c.ButtonIndex >= 0 {
					var button *buttons.TextButton

					if (c.TimerActive) {
						st := t.Sub(c.StartTime)

						out := time.Time{}.Add(st)
						if st > 1 * time.Hour {
							button = buttons.NewTextButtonWithColours(fmt.Sprintf("%s", out.Format("15h04m")), color.White, color.RGBA{255, 0, 10, 255})
						} else {
							button = buttons.NewTextButtonWithColours(fmt.Sprintf("%s", out.Format("4m05s")), color.White, color.RGBA{uint8(out.Minute() + 195), 0, 0, 255})
						}
					} else {
						button = buttons.NewTextButton(fmt.Sprintf("%02d:%02d:%02d", t.Hour(), t.Minute(), t.Second()))
					}
					button.SetActionHandler(&TimerAction{StartTime: t, Clock: c})
					c.SD.AddButton(c.ButtonIndex, button)
				}
			}
		}
	}()
}

func (c *Clock) SetClockButton(offset int) {
    c.ButtonIndex = offset
}

func (c *Clock) Reset() {
    c.ButtonIndex = -1
}
