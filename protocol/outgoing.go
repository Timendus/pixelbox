package protocol

// This file exposes a bunch of command functions. Each function returns a slice
// of bytes to send over the Bluetooth serial link to the device. The function
// names and signatures will probably be self-descriptive enough.

import (
	"fmt"
	"image"
	"math"
	"time"
)

func GetSettings() []byte {
	return wrap([]byte{
		getSettings,
	})
}

func DisplayOff() []byte {
	return wrap([]byte{
		setChannel,
		channels["LIGHT"],
		0, 0, 0, // color
		0,       // brightness
		0,       // light type
		0,       // power off
		0, 0, 0, // fixed ending
	})
}

func SetTime(moment time.Time) []byte {
	return wrap([]byte{
		setTime,
		byte(moment.Year() % 100),
		byte(moment.Year() / 100),
		byte(moment.Month()),
		byte(moment.Day()),
		byte(moment.Hour()),
		byte(moment.Minute()),
		byte(moment.Second()),
		0x00,
	})
}

func SetVolume(volume int) ([]byte, error) {
	if volume < 0 || volume > 16 {
		return nil, fmt.Errorf("volume should be between 0 and 16")
	}

	return wrap([]byte{setVolume, byte(volume)}), nil
}

func SetBrightness(brightness int) ([]byte, error) {
	if brightness < 0 || brightness > 100 {
		return nil, fmt.Errorf("brightness should be between 0 and 100")
	}

	return wrap([]byte{setBrightness, byte(brightness)}), nil
}

func SetWeather(temperature int, wtype WeatherType) ([]byte, error) {
	weatherTypeId, ok := weatherTypes[wtype]
	if !ok {
		return nil, fmt.Errorf("invalid weather type")
	}

	if temperature <= -100 || temperature >= 100 {
		return nil, fmt.Errorf("temperature out of bounds")
	}

	return wrap([]byte{setWeather, byte(temperature), weatherTypeId}), nil
}

func ShowClock(ctype ClockType, showTime, showWeather, showTemp, showCal bool, color Color) ([]byte, error) {
	clockType, ok := clockTypes[ctype]
	if !ok {
		return nil, fmt.Errorf("invalid clock type")
	}

	return wrap([]byte{
		setChannel,
		channels["CLOCK"],
		0x01, // magic number for unknown reason
		clockType,
		conditional(showTime),
		conditional(showWeather),
		conditional(showTemp),
		conditional(showCal),
		color[0],
		color[1],
		color[2],
	}), nil
}

func ShowLight(ltype LightType, color Color, brightness int) ([]byte, error) {
	if brightness < 0 || brightness > 100 {
		return nil, fmt.Errorf("brightness should be between 0 and 100")
	}

	lightType, ok := lightTypes[ltype]
	if !ok {
		return nil, fmt.Errorf("invalid light type")
	}

	return wrap([]byte{
		setChannel,
		channels["LIGHT"],
		color[0],
		color[1],
		color[2],
		byte(brightness),
		lightType,
		1,       // power on
		0, 0, 0, // magic ending
	}), nil
}

func ShowCloud() []byte {
	return wrap([]byte{setChannel, channels["CLOUD"]})
}

func ShowVJEffect(effect int) ([]byte, error) {
	// This one doesn't seem to work for me. But it could be that I disabled it
	// at some point through the app, or maybe it needs music to be playing..?

	if effect < 0 || effect > 15 {
		return nil, fmt.Errorf("effect should be a value between 0 and 15")
	}

	return wrap([]byte{
		setChannel,
		channels["VJ"],
		byte(effect),
	}), nil
}

func ShowVisualisation(visualisation int) ([]byte, error) {
	if visualisation < 0 || visualisation > 11 {
		return nil, fmt.Errorf("visualisation should be a value between 0 and 11")
	}

	return wrap([]byte{
		setChannel,
		channels["VISUALISATION"],
		byte(visualisation),
	}), nil
}

func ShowScoreBoard(redPlayer, bluePlayer int) ([]byte, error) {
	if redPlayer < 0 || redPlayer > 999 || bluePlayer < 0 || bluePlayer > 999 {
		return nil, fmt.Errorf("player scores should be between 0 and 999")
	}

	return wrap([]byte{
		setChannel,
		channels["SCOREBOARD"],
		0x00, // magic number for unknown reason
		byte(redPlayer),
		byte(redPlayer >> 8),
		byte(bluePlayer),
		byte(bluePlayer >> 8),
	}), nil
}

func ShowImage(image *image.RGBA) ([]byte, error) {
	paletteData, imageData, err := convertImage(image)
	if err != nil {
		return nil, err
	}

	imageSize := 1 + // Start of frame indicator
		2 + // Frame size
		2 + // Frame time
		1 + // Palette reset
		1 + // Number of colours
		len(paletteData) + // New palette data
		len(imageData) // New image data

	command := []byte{
		setImage,
		0x00, 0x0A, 0x0A, 0x04, // Voodoo magic
		startOfFrame,    // Start of frame indicator
		byte(imageSize), // Size of frame from 0xAA header onward
		byte(imageSize >> 8),
		0x00, 0x00, // How long to show this frame (0 is infinite? ignored?)
		resetPalette,
		byte(len(paletteData) / 3), // Number of colours
	}
	command = append(command, paletteData...)
	command = append(command, imageData...)

	return wrap(command), nil
}

func ShowAnimation(frames []*image.RGBA, durationsMs []int) ([]byte, error) {
	// Convert frames to frame data as device expects it
	frameData := make([]byte, 0)
	for i, frame := range frames {
		paletteData, imageData, err := convertImage(frame)
		if err != nil {
			return nil, err
		}

		frameSize := 1 + // start of frame indicator
			2 + // Frame size
			2 + // Frame time
			1 + // Palette reset
			1 + // Number of colours
			len(paletteData) + // New palette data
			len(imageData) // new image data

		frame := []byte{
			startOfFrame,
			byte(frameSize),
			byte(frameSize >> 8),
			byte(durationsMs[i]),
			byte(durationsMs[i] >> 8),
			resetPalette,
			byte(len(paletteData) / 3),
		}
		frame = append(frame, paletteData...)
		frame = append(frame, imageData...)

		frameData = append(frameData, frame...)
	}

	// Create packets to stream to the Divoom
	packetNum := 0
	totalSize := len(frameData)
	commands := make([]byte, 0)
	for i := 0; i < len(frameData); i += 200 {
		command := []byte{
			setAnimation,
			byte(totalSize), // Size of all the frames
			byte(totalSize >> 8),
			byte(packetNum), // Packet number, starting from zero
		}
		j := int(math.Min(float64(i+400), float64(len(frameData))))
		command = append(command, frameData[i:j]...)
		commands = append(commands, wrap(command)...)
		packetNum++
	}

	return commands, nil
}

func conditional(input bool) byte {
	if input {
		return 1
	}
	return 0
}
