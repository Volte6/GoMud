package term

import (
	"errors"
	"strconv"
)

type ANSIByte = byte

var (
	ANSI_ESC  byte = 27 // \033
	ANSI_BEEP byte = 7  // When sent to client, causes beep (or windows chime, or whatever)
)

// ESC [  <  0  ;  1  ;  1  m
// [27 91 60 48 59 49 59 49 109]
// ESC [  <  0  ;  1  4  0  ;  4  8  m
// [27 91 60 48 59 49 52 48 59 52 56 109]
func AnsiParseMouseClickPayload(info []byte) (xPos int, yPos int, err error) {
	if len(info) > 1 {
		for i := 0; i < len(info)-1; i++ {
			if info[i] == 59 { // ;
				xPos, err = strconv.Atoi(string(info[0:i]))
				if err != nil {
					return xPos, yPos, err
				}
				yPos, err = strconv.Atoi(string(info[i+1:]))
				return xPos, yPos, err
			}
		}
	}
	err = errors.New("invalid mouse click code")
	return xPos, yPos, err
}

func AnsiParseScreenSizePayload(info []byte) (width int, height int, err error) {
	if len(info) >= 1 {
		// find the semicolon
		for i := 0; i < len(info); i++ {
			// when we hit the SEMICOLON, make sure it's not the end of the []byte
			if info[i] == ';' && i < len(info)-1 {
				height, err = strconv.Atoi(string(info[0:i]))
				if err != nil {
					return width, height, err
				}
				width, err = strconv.Atoi(string(info[i+1:]))
				return width, height, err
			}
		}
	}
	return width, height, errors.New("invalid screen size code")
}

func AnsiParseMouseWheelScroll(info []byte) (xPos int, yPos int, err error) {
	if len(info) > 1 {
		for i := 0; i < len(info)-1; i++ {
			if info[i] == 59 { // ;
				xPos, err = strconv.Atoi(string(info[0:i]))
				if err != nil {
					return xPos, yPos, err
				}
				yPos, err = strconv.Atoi(string(info[i+1:]))
				return xPos, yPos, err
			}
		}
	}
	err = errors.New("invalid mouse scroll code")
	return xPos, yPos, err
}
