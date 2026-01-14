package protocol

import "maps"

// These are the valid strings for the weather, clock and light types

type WeatherType string
type ClockType string
type LightType string

var weatherTypes = map[WeatherType]byte{
	"OUTDOOR_VERY_LIGHT_CLOUDS": 1,
	"CITY_CLOUDY":               3,
	"CITY_LIGHT_CLOUDS":         4,
	"THUNDERSTORM":              5,
	"RAIN":                      6,
	"SNOW":                      8,
	"FOG":                       9,
}

var clockTypes = map[ClockType]byte{
	"FULL_SCREEN":          0,
	"RAINBOW":              1,
	"BOXED":                2,
	"ANALOG_SQUARE":        3,
	"FULL_SCREEN_INVERTED": 4,
	"ANALOG_ROUND":         5,
}

var lightTypes = map[LightType]byte{
	"PLAIN":            0,
	"TINTED_PINK":      1,
	"RED_BLUE_STRIPED": 2,
}

var channels = map[string]byte{
	"CLOCK":         0,
	"LIGHT":         1,
	"CLOUD":         2,
	"VJ":            3,
	"VISUALISATION": 4,
	"ANIMATION":     5,
	"SCOREBOARD":    6,
}

var reverseChannels = func() map[byte]string {
	chans := make(map[byte]string, 0)
	for key := range maps.Keys(channels) {
		chans[channels[key]] = key
	}
	return chans
}()

// Envelope values
const (
	prefix  = 0x01
	postfix = 0x02
)

// Incoming values
const (
	header1 = 0x04
	header2 = 0x55

	volumeSet     = 0x09
	alarmConfig   = 0x13
	timeSet       = 0x18
	acknowledge   = 0x31
	brightnessSet = 0x32
	imageSet      = 0x44
	channelSet    = 0x45
	settingsSet   = 0x46
	animationSet  = 0x49
	buttonPress   = 0xBD
)

// Outgoing values
const (
	setVolume     = 0x08
	setTime       = 0x18
	setImage      = 0x44
	setChannel    = 0x45
	getSettings   = 0x46
	setAnimation  = 0x49
	setWeather    = 0x5F
	setBrightness = 0x74

	startOfFrame = 0xAA
	resetPalette = 0x00
)
