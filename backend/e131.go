package main

import (
	"log"

	"github.com/Hundemeier/go-sacn/sacn"
)

func setupSACN() map[uint16]chan<- []byte {
	trans, err := sacn.NewTransmitter("", [16]byte{0x47, 0x6f, 0x4c, 0x45, 0x44, 0x7a}, "GoLEDz")
	if err != nil {
		log.Fatal(err)
	}

	universeNumbers := []uint16{1, 3}
	universes := make(map[uint16]chan<- []byte)

	for _, universeNumber := range universeNumbers {
		universes[universeNumber], err = trans.Activate(universeNumber)
		if err != nil {
			log.Fatal(err)
		}
		trans.SetDestinations(universeNumber, []string{config.ControllerAddress})
	}

	return universes
}
