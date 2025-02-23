package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

const (
	// E1.31 Constants
	RootPreambleSize  = uint16(0x0010)
	RootPostambleSize = uint16(0x0000)
	RootACNPacketID   = "ASC-E1.17"
	RootVector        = uint32(0x00000004)
	FramingVector     = uint32(0x00000002)
	DMPVector         = uint8(0x02)
	DefaultPriority   = uint8(100)
	DefaultStartCode  = uint8(0x00)
	UniverseChannels  = 512
	KeepAliveInterval = time.Second

	// Packet options
	StreamTerminateOptionsBit = 6

	// UDP
	SACNPort        = 5568
	MaxPacketSize   = 638
	HeaderLength    = 126
	RootLayerLength = 38
)

type Destination struct {
	Addr     *net.UDPAddr
	Priority uint8
}

type Transmitter struct {
	conn       *net.UDPConn
	universes  map[uint16]*Universe
	cid        [16]byte
	sourceName string
	mu         sync.RWMutex
	done       chan struct{}
	wg         sync.WaitGroup
}

type Universe struct {
	number           uint16
	sequence         uint8
	destinations     []Destination
	data             []byte
	dataChan         chan []byte
	lastSent         time.Time
	priority         uint8
	forceSync        bool
	streamTerminated bool
}

type TransmitterConfig struct {
	CID        [16]byte
	SourceName string
	Priority   uint8
}

func NewTransmitter(config TransmitterConfig) (*Transmitter, error) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{})
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP connection: %w", err)
	}

	t := &Transmitter{
		conn:       conn,
		universes:  make(map[uint16]*Universe),
		cid:        config.CID,
		sourceName: config.SourceName,
		done:       make(chan struct{}),
	}

	t.wg.Add(1)
	go t.keepAliveLoop()

	return t, nil
}

func calculateFlagsAndLength(length uint16) []byte {
	// Mask length to 12 bits and set high nibble to 0x7
	value := uint16(0x7000) | (length & 0x0FFF)

	// Split into two bytes
	return []byte{
		byte(value >> 8),   // High byte
		byte(value & 0xFF), // Low byte
	}
}

func (t *Transmitter) createPacket(universe *Universe) []byte {
	packet := make([]byte, MaxPacketSize)

	// Root Layer Preamble (bytes 0-15)
	binary.BigEndian.PutUint16(packet[0:], 0x0010)      // Preamble Size
	binary.BigEndian.PutUint16(packet[2:], 0x0000)      // Postamble Size
	copy(packet[4:16], []byte("ASC-E1.17\000\000\000")) // ACN Packet Identifier

	// Root Layer PDU (bytes 16-37)
	copy(packet[16:18], calculateFlagsAndLength(uint16(len(universe.data)+110)))
	copy(packet[18:22], []byte{0x00, 0x00, 0x00, 0x04}) // Root Vector
	copy(packet[22:38], t.cid[:])                       // CID

	// Framing Layer PDU
	copy(packet[38:40], calculateFlagsAndLength(uint16(len(universe.data)+88)))
	copy(packet[40:44], []byte{0x00, 0x00, 0x00, 0x02}) // Framing Vector
	copy(packet[44:108], padString(t.sourceName, 64))   // Source Name

	packet[108] = universe.priority     // Priority
	packet[109] = 0x00                  // Sync Address
	packet[110] = 0x00                  // Sync Address
	packet[111] = universe.sequence     // Sequence Number
	packet[112] = 0x00                  // Options Flags
	packet[113] = 0x00                  // First byte for Universe
	packet[114] = byte(universe.number) // Universe Number (in correct position)
	packet[115] = 0x72                  // Required by protocol
	packet[116] = 0x0b                  // Required by protocol
	packet[117] = DMPVector             // 0x02
	packet[118] = 0xa1                  // Address & Data Type

	// DMP Layer
	packet[119] = 0x00 // First Property Address
	packet[120] = 0x00 // First Property Address
	packet[121] = 0x00 // Address Increment
	packet[122] = 0x01 // Address Increment
	binary.BigEndian.PutUint16(packet[123:125], uint16(len(universe.data)+1))
	packet[125] = DefaultStartCode // DMX Start Code (0x00)

	// DMX Data
	copy(packet[126:], universe.data)

	return packet[:126+len(universe.data)]
}

func (t *Transmitter) Activate(universeNumber uint16) (chan<- []byte, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.universes[universeNumber]; exists {
		return nil, fmt.Errorf("universe %d already activated", universeNumber)
	}

	universe := &Universe{
		number:       universeNumber,
		sequence:     0,
		destinations: make([]Destination, 0),
		data:         make([]byte, UniverseChannels),
		dataChan:     make(chan []byte, 100),
		priority:     DefaultPriority,
	}

	t.universes[universeNumber] = universe

	t.wg.Add(1)
	go t.handleUniverse(universe)

	return universe.dataChan, nil
}

func (t *Transmitter) SetDestination(universe uint16, addr string, priority uint8) error {
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", addr, SACNPort))
	if err != nil {
		return fmt.Errorf("invalid address %s: %w", addr, err)
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	state, exists := t.universes[universe]
	if !exists {
		return fmt.Errorf("universe %d not activated", universe)
	}

	state.destinations = []Destination{{
		Addr:     udpAddr,
		Priority: priority,
	}}

	return nil
}

func (t *Transmitter) handleUniverse(universe *Universe) {
	defer t.wg.Done()

	for {
		select {
		case data := <-universe.dataChan:
			t.mu.Lock()
			universe.data = data
			universe.lastSent = time.Now()
			t.sendToDestinations(universe)
			t.mu.Unlock()
		case <-t.done:
			// TODO: deactivate the universe?
			return
		}
	}
}

func (t *Transmitter) sendToDestinations(universe *Universe) {
	if len(universe.data) == 0 || len(universe.data) > 512 {
		log.Printf("Invalid DMX data length for universe %d: %d", universe.number, len(universe.data))
		return
	}

	packet := t.createPacket(universe)
	universe.sequence++

	for _, dest := range universe.destinations {
		if config.LocalOnly {
			continue
		}

		if _, err := t.conn.WriteToUDP(packet, dest.Addr); err != nil {
			log.Printf("Error sending to universe %d: %v", universe.number, err)
			continue
		}
	}
}

func (t *Transmitter) keepAliveLoop() {
	defer t.wg.Done()

	ticker := time.NewTicker(KeepAliveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			t.mu.Lock()
			now := time.Now()
			for _, universe := range t.universes {
				if now.Sub(universe.lastSent) >= KeepAliveInterval {
					t.sendToDestinations(universe)
				}
			}
			t.mu.Unlock()
		case <-t.done:
			return
		}
	}
}

func (t *Transmitter) Close() error {
	close(t.done)
	t.wg.Wait()

	t.mu.Lock()
	defer t.mu.Unlock()

	for _, state := range t.universes {
		close(state.dataChan)
	}

	return t.conn.Close()
}

func padString(s string, length int) []byte {
	b := make([]byte, length)
	copy(b, s)
	return b
}
