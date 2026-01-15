package controllers

import (
	"fmt"

	"github.com/timendus/pixelbox/protocol"
)

type Scene struct {
	Name             string
	Id               string
	ChangeBrightness bool
	Brightness       *int
	ChangeVolume     bool
	Volume           *int
	SceneType        string
	Clock            Clock
	Weather          Weather
	Temperature      Temperature
	Calendar         Calendar
	Light            Light
	Effect           Effect
}

type Clock struct {
	Enabled bool
	CType   string `json:"type"`
	Color   string
}

type Weather struct {
	Enabled bool
	WType   string `json:"type"`
}

type Temperature struct {
	Enabled     bool
	Temperature *int
}

type Calendar struct {
	Enabled bool
}

type Light struct {
	LType string `json:"type"`
	Color string
}

type Effect struct {
	EType             string `json:"type"`
	VJType            *int
	VisualisationType *int
	ScoreRedPlayer    *int
	ScoreBluePlayer   *int
}

func (s *Scene) ToMessage() ([]byte, error) {
	result := make([]byte, 0)

	// Brightness setting
	if s.SceneType != "light" && s.ChangeBrightness && s.Brightness != nil {
		msg, err := protocol.SetBrightness(*s.Brightness)
		if err != nil {
			return nil, err
		}
		result = append(result, msg...)
	}

	// Volume setting
	if s.ChangeVolume && s.Volume != nil {
		msg, err := protocol.SetVolume(*s.Volume)
		if err != nil {
			return nil, err
		}
		result = append(result, msg...)
	}

	switch s.SceneType {
	case "clock":
		msg, err := protocol.ShowClock(
			protocol.ClockType(s.Clock.CType),
			s.Clock.Enabled,
			s.Weather.Enabled,
			s.Temperature.Enabled,
			s.Calendar.Enabled,
			protocol.ColorFromHex(s.Clock.Color),
		)
		if err != nil {
			return nil, err
		}
		result = append(result, msg...)

		if s.Weather.Enabled || s.Temperature.Enabled {
			temp := 0
			if s.Temperature.Temperature != nil {
				temp = *s.Temperature.Temperature
			}
			msg, err := protocol.SetWeather(temp, protocol.WeatherType(s.Weather.WType))
			if err != nil {
				return nil, err
			}
			result = append(result, msg...)
		}

	case "light":
		if s.Brightness == nil {
			return nil, fmt.Errorf("brightness needs to be set to be able to show a light")
		}
		msg, err := protocol.ShowLight(
			protocol.LightType(s.Light.LType),
			protocol.ColorFromHex(s.Light.Color),
			*s.Brightness,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, msg...)

	case "effects":
		switch s.Effect.EType {
		case "CLOUD":
			result = append(result, protocol.ShowCloud()...)

		case "VJ":
			if s.Effect.VJType == nil {
				return nil, fmt.Errorf("VJ type required for showing a VJ effect")
			}
			msg, err := protocol.ShowVJEffect(*s.Effect.VJType)
			if err != nil {
				return nil, err
			}
			result = append(result, msg...)

		case "VISUALISATION":
			if s.Effect.VisualisationType == nil {
				return nil, fmt.Errorf("visualisation type required for showing a visualisation")
			}
			msg, err := protocol.ShowVisualisation(*s.Effect.VisualisationType)
			if err != nil {
				return nil, err
			}
			result = append(result, msg...)

		case "SCOREBOARD":
			if s.Effect.ScoreBluePlayer == nil || s.Effect.ScoreRedPlayer == nil {
				return nil, fmt.Errorf("player scores required for showing the score board")
			}
			msg, err := protocol.ShowScoreBoard(*s.Effect.ScoreRedPlayer, *s.Effect.ScoreBluePlayer)
			if err != nil {
				return nil, err
			}
			result = append(result, msg...)
		}

	case "image":

	case "animation":

	}

	return result, nil
}
