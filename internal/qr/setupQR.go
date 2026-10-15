package qr

import (
	"fmt"
	"strings"

	"github.com/mdp/qrterminal/v4"
)

func Render(url string) (string, error) {
	var buf strings.Builder

	config := qrterminal.Config{
		Level:      qrterminal.L,
		Writer:     &buf,
		HalfBlocks: true, 
		QuietZone:  1,
	}

	qrterminal.GenerateWithConfig(url, config)

	return buf.String(), nil
}

func Print(url string) error {
	qrText, err := Render(url)
	if err != nil {
		return err
	}

	fmt.Print(qrText)

	return nil
}
