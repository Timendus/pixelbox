package protocol

// This file wraps and unwraps the commands to send and the messages to receive
// in the required envelope. The envelope looks like this:
//
//   [ PREFIX, LL, LL, <data>, CH, CH, POSTFIX ]
//
// The prefix and postfix are fixed numbers. The two LL bytes together encode
// the length of the data (plus two, for either the length itself or the
// checksum?) as an LSB value. The two CH bytes together form a checksum over
// the length plus the data, again as an LSB value.

import (
	"fmt"
)

func wrap(command []byte) []byte {
	envelope := []byte{prefix, 0, 0}
	envelope = append(envelope, command...)
	length := len(envelope) - 1
	envelope[1] = byte(length)
	envelope[2] = byte(length >> 8)
	checksum := calcChecksum(envelope[1:])
	envelope = append(envelope, byte(checksum), byte(checksum>>8))
	envelope = append(envelope, postfix)
	return envelope
}

func unwrap(envelope []byte) ([][]byte, error) {
	// We can be receiving multiple messages in one burst. So parse them in a
	// loop until we run out of bytes. This does assume that we always get full
	// messages and never partial ones. This is something that should be solved
	// by the user on a higher level.
	index := 0
	messages := make([][]byte, 0)
	for index < len(envelope) {
		if envelope[index] != prefix {
			return nil, fmt.Errorf("expected message to start with the right prefix")
		}

		length := int(envelope[index+1]) + int(envelope[index+2])<<8
		checksumIndex := index + 1 + length
		endIndex := checksumIndex + 2

		if envelope[endIndex] != postfix {
			return nil, fmt.Errorf("expected message to end with the right postfix")
		}

		checksum := uint16(envelope[checksumIndex]) + uint16(envelope[checksumIndex+1])<<8
		payload := envelope[index+1 : index+1+length]
		calculatedChecksum := calcChecksum(payload)

		if checksum != calculatedChecksum {
			return nil, fmt.Errorf("invalid checksum received, expected %d, got %d", calculatedChecksum, checksum)
		}

		// Leave the length out
		messages = append(messages, payload[2:])
		index += 1 + length + 2 + 1
	}
	return messages, nil
}

func calcChecksum(payload []byte) uint16 {
	checksum := uint16(0)
	for _, b := range payload {
		checksum += uint16(b)
	}
	return checksum
}
