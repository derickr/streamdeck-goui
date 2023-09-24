package addons

import (
	"fmt"
	"image/color"
	"math"
	"strconv"
	"time"

	//"github.com/derickr/streamdeck-goui/actionhandlers"
	"github.com/crazy3lf/colorconv"
	"github.com/magicmonkey/go-streamdeck"
	"github.com/magicmonkey/go-streamdeck/buttons"
	//sddecorators "github.com/magicmonkey/go-streamdeck/decorators"
	"github.com/rs/zerolog/log"
)

type TimerAction struct {
	StartTime   time.Time
	Clock       *Clock
	ButtonIndex int
}

func (t *TimerAction) Pressed(btn streamdeck.Button) {
	index := btn.GetButtonIndex()

	if t.Clock.TimersActive[t.ButtonIndex] {
		duration := time.Now().Sub(t.Clock.StartTimes[index])
		out := time.Time{}.Add(duration)

		log.Info().Msgf("Elapsed time for %s: %s", t.Clock.ClockNames[index], out.Format("15:04:05"))
		t.Clock.TimersActive[t.ButtonIndex] = false
		return
	}

	t.Clock.StartTimes[t.ButtonIndex] = t.StartTime
	t.Clock.TimersActive[t.ButtonIndex] = true
}

type Clock struct {
	SD           *streamdeck.StreamDeck
	ClockButtons [32]bool
	Hues         [32]int
	ClockNames   [32]string
	dones        [32]chan bool
	Tickers      [32]*time.Ticker
	TimersActive [32]bool
	StartTimes   [32]time.Time
}

func (c *Clock) Init() {
	c.Reset()

	for i := 0; i < 32; i++ {
		c.dones[i] = make(chan bool)

		c.Tickers[i] = time.NewTicker(1000 * time.Millisecond)

		go func(index int) {
			for {
				select {
				case <-c.dones[index]:
					return
				case t := <-c.Tickers[index].C:
					if c.ClockButtons[index] {
						var button *buttons.TextButton

						if c.TimersActive[index] {
							st := t.Sub(c.StartTimes[index])
							out := time.Time{}.Add(st)

							r, g, b, _ := colorconv.HSLToRGB(float64(c.Hues[index]), math.Min(0.25+float64(out.Minute()/30.0), 0.75), 0.5)

							if st > 1*time.Hour {
								button = buttons.NewTextButtonWithColours(fmt.Sprintf("%s", out.Format("15h04")), color.White, color.RGBA{r, g, b, 255})
							} else {
								button = buttons.NewTextButtonWithColours(fmt.Sprintf("%s", out.Format("04:05")), color.White, color.RGBA{r, g, b, 255})
							}
						} else if c.ClockNames[index] != "" {
							r, g, b, _ := colorconv.HSLToRGB(float64(c.Hues[index]), 0.5, 0.5)
							button = buttons.NewTextButtonWithColours(c.ClockNames[index], color.White, color.RGBA{r, g, b, 255})
						} else {
							button = buttons.NewTextButton(fmt.Sprintf("%02d:%02d:%02d", t.Hour(), t.Minute(), t.Second()))
						}
						button.SetActionHandler(&TimerAction{StartTime: t, Clock: c, ButtonIndex: index})
						c.SD.AddButton(index, button)
					}
				}
			}
		}(i)
	}
}

func (c *Clock) AddClockButton(offset int, hue string, inactiveImage string) {
	c.ClockButtons[offset] = true
	c.Hues[offset], _ = strconv.Atoi(hue)
	c.ClockNames[offset] = inactiveImage
}

func (c *Clock) Reset() {
	for i := 0; i < 32; i++ {
		c.ClockButtons[i] = false
		c.Hues[i] = 0
		c.ClockNames[i] = ""
	}
}
