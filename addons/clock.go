package addons

import (
	"fmt"
	"image/color"
	"math"
	"os/exec"
	"strconv"
	"strings"
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

func ReadComment() string {
	c, b := exec.Command("zenity", "--title=Enter Task Comment", "--entry", "--text=Comment:"), new(strings.Builder)
	c.Stdout = b
	c.Run()
	return b.String()
}

func (t *TimerAction) Pressed(btn streamdeck.Button) {
	index := btn.GetButtonIndex()

	if t.Clock.TimersActive[t.ButtonIndex] {
		var comment string

		durationF := float64(time.Now().Sub(t.Clock.StartTimes[index]).Nanoseconds()) * t.Clock.ClockSpeeds[index]
		duration := time.Duration(durationF)
		out := time.Time{}.Add(duration)

		if t.Clock.ClockNames[index] != "" {
			comment = ReadComment()
		} else {
			comment = ""
		}

		log.Info().
			Str("project", t.Clock.ClockNames[index]).
			Str("start", t.Clock.StartTimes[index].Format("2006-01-02 15:04:05")).
			Str("end", time.Now().Format("2006-01-02 15:04:05")).
			Str("textual_length", out.Format("15:04:05")).
			Float64("length", duration.Hours()).
			Str("comment", comment).
			Send()

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
	ClockSpeeds  [32]float64
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
								button = buttons.NewTextButtonWithColoursAndMargin(fmt.Sprintf("%s", out.Format("15h04")), color.White, color.RGBA{r, g, b, 255}, 0)
							} else {
								button = buttons.NewTextButtonWithColoursAndMargin(fmt.Sprintf("%s", out.Format("04:05")), color.White, color.RGBA{r, g, b, 255}, 0)
							}
						} else if c.ClockNames[index] != "" {
							r, g, b, _ := colorconv.HSLToRGB(float64(c.Hues[index]), 0.5, 0.5)
							button = buttons.NewTextButtonWithColoursAndMargin(c.ClockNames[index], color.White, color.RGBA{r, g, b, 255}, 0)
						} else {
							button = buttons.NewTextButtonWithMargin(fmt.Sprintf("%02d:%02d:%02d", t.Hour(), t.Minute(), t.Second()), 0)
						}
						button.SetActionHandler(&TimerAction{StartTime: t, Clock: c, ButtonIndex: index})
						c.SD.AddButton(index, button)
					}
				}
			}
		}(i)
	}
}

func (c *Clock) AddClockButton(offset int, hue string, inactiveImage string, speed float64) {
	c.ClockButtons[offset] = true
	c.Hues[offset], _ = strconv.Atoi(hue)
	c.ClockNames[offset] = inactiveImage
	c.ClockSpeeds[offset] = speed
}

func (c *Clock) Reset() {
	for i := 0; i < 32; i++ {
		c.ClockButtons[i] = false
		c.Hues[i] = 0
		c.ClockNames[i] = ""
		c.ClockSpeeds[i] = 1
	}
}
