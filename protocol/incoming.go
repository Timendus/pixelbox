package protocol

// This file exposes the `ParseIncoming` function, which expects a slice of
// bytes received over the Bluetooth serial link. It will return `Message`s,
// which tell you something about what the device is trying to communicate. One
// slice of bytes can hold multiple messages, in which case each message will be
// decoded and returned. However, it is up to the user to make sure we have no
// incomplete messages in the input. Also; if any of the messages fails to
// decode, all messages fail to decode.

import (
	"fmt"
)

type Message struct {
	Command          byte
	Data             []byte
	CurrentChannel   byte
	Brightness       byte
	Volume           byte
	HumanDescription string
}

func (m *Message) String() string {
	if len(m.HumanDescription) > 0 {
		return m.HumanDescription
	}
	return fmt.Sprintf("Unknown message with command ID %d and data %d", m.Command, m.Data)
}

func ParseIncoming(envelope []byte) ([]*Message, error) {
	messages, err := unwrap(envelope)
	if err != nil {
		return nil, err
	}

	decoded := make([]*Message, 0)
	for _, message := range messages {
		decodedMessage, err := decodeMessage(message)
		if err != nil {
			return nil, err
		}
		decoded = append(decoded, decodedMessage)
	}

	return decoded, nil
}

func decodeMessage(bytes []byte) (*Message, error) {
	// Check the message adheres to the magic structure described
	if len(bytes) < 3 || bytes[0] != header1 || bytes[2] != header2 {
		return nil, fmt.Errorf("received a message format I don't understand")
	}

	message := Message{
		Command: bytes[1],
		Data:    bytes[3:],
	}

	// This is what we know about the returned data structures. Since this has
	// all been reverse engineered, there is no way of knowing if it is really
	// correct.
	switch message.Command {
	case settingsSet:
		// This command is probably the least well understood. There's a lot
		// more data in it than we're able to get out right now.
		if len(message.Data) >= 21 {
			message.CurrentChannel = message.Data[20]
			message.Brightness = message.Data[6]
			message.HumanDescription = fmt.Sprintf("Updated settings to brightness %d and channel %d (%s)", message.Brightness, message.CurrentChannel, reverseChannels[message.CurrentChannel])
		} else if len(message.Data) >= 20 {
			message.Brightness = message.Data[6]
			message.HumanDescription = fmt.Sprintf("Received requested settings with brightness %d", message.Brightness)
		}
	case channelSet:
		if len(message.Data) >= 1 {
			message.CurrentChannel = message.Data[0]
			message.HumanDescription = fmt.Sprintf("Set channel to %d", message.CurrentChannel)
		}
	case brightnessSet:
		if len(message.Data) >= 1 {
			message.Brightness = message.Data[0]
			message.HumanDescription = fmt.Sprintf("Set brightness to %d", message.Brightness)
		}
	case timeSet:
		message.HumanDescription = "Time was set"
	case imageSet:
		message.HumanDescription = "Image was shown"
	case animationSet:
		message.HumanDescription = "Animation was shown"
	case acknowledge:
		if len(message.Data) >= 1 {
			message.Brightness = message.Data[0]
			message.HumanDescription = fmt.Sprintf("Light or clock was set with brightness %d", message.Brightness)
		}
	case volumeSet:
		if len(message.Data) >= 1 {
			message.Volume = message.Data[0]
			message.HumanDescription = fmt.Sprintf("Volume was set to %d/16", message.Volume)
		}
	case buttonPress:
		if match(message.Data, []byte{19, 1, 50, 0}) {
			message.HumanDescription = "Play button was pressed"
		}
		if match(message.Data, []byte{23, 0}) {
			message.HumanDescription = "Light button was double-clicked"
		}
		if match(message.Data, []byte{19, 1, 30, 0}) {
			message.HumanDescription = "Clock button was double-clicked"
		}
	case alarmConfig:
		if len(message.Data) == 1 {
			switch message.Data[0] {
			case 0:
				message.HumanDescription = "Exit alarm config"
			case 10:
				message.HumanDescription = "Entered alarm config"
			}
		}
	}

	return &message, nil
}

func match(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, n := range a {
		if n != b[i] {
			return false
		}
	}
	return true
}
