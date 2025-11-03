// Copyright 2019-2024 go-sccp authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

/*
Package sccp provides encoding/decoding feature of Signalling Connection Control Part used in SS7/SIGTRAN protocol stack.

This is still an experimental project, and currently in its very early stage of development. Any part of implementations
(including exported APIs) may be changed before released as v1.0.0.
*/
package sccp

import (
	"encoding"
	"fmt"
	"io"
	"sync"
	"time"
)

// MsgType is type of SCCP message.
type MsgType uint8

// Message Type definitions.
const (
	_            MsgType = iota
	MsgTypeCR            // CR
	MsgTypeCC            // CC
	MsgTypeCREF          // CREF
	MsgTypeRLSD          // RLSD
	MsgTypeRLC           // RLC
	MsgTypeDT1           // DT1
	MsgTypeDT2           // DT2
	MsgTypeAK            // AK
	MsgTypeUDT           // UDT
	MsgTypeUDTS          // UDTS
	MsgTypeED            // ED
	MsgTypeEA            // EA
	MsgTypeRSR           // RSR
	MsgTypeRSC           // RSC
	MsgTypeERR           // ERR
	MsgTypeIT            // IT
	MsgTypeXUDT          // XUDT
	MsgTypeXUDTS         // XUDTS
	MsgTypeLUDT          // LUDT
	MsgTypeLUDTS         // LUDTS
)

// SSNState represents the state of a subsystem
type SSNState uint8

const (
	SSNStateProhibited SSNState = 0 // Out of Service
	SSNStateAllowed    SSNState = 1 // In Service
)

// State change reasons
type StateChangeReason uint8

const (
	ReasonUserInitiated    StateChangeReason = 1
	ReasonNetworkInitiated StateChangeReason = 2
	ReasonTestTimeout      StateChangeReason = 3
	ReasonTestResponse     StateChangeReason = 4
)

// Broadcast types
type BroadcastType uint8

const (
	BroadcastSSA BroadcastType = 1 // Subsystem Allowed
	BroadcastSSP BroadcastType = 2 // Subsystem Prohibited
)

// SSNEntry represents a subsystem entry with state management
type SSNEntry struct {
	SSN             uint8
	PointCode       uint16
	State           SSNState
	IsLocal         bool
	LastStateChange time.Time
	TestTimer       *time.Timer
	TestInterval    time.Duration
	TestRetries     int
	MaxTestRetries  int
	mutex           sync.RWMutex
}

// State check methods
func (s *SSNEntry) IsAllowed() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.State == SSNStateAllowed
}

func (s *SSNEntry) IsProhibited() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.State == SSNStateProhibited
}

func (s *SSNEntry) MarkAllowed() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.State = SSNStateAllowed
	s.LastStateChange = time.Now()
}

func (s *SSNEntry) MarkProhibited() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.State = SSNStateProhibited
	s.LastStateChange = time.Now()
}

// SSNStateManager manages all subsystem states
type SSNStateManager struct {
	entries map[string]*SSNEntry // key: "pc:ssn"
	mutex   sync.RWMutex

	// Configuration
	DefaultTestInterval time.Duration
	MaxTestInterval     time.Duration
	MaxTestRetries      int

	// Callbacks
	OnStateChange func(*SSNEntry, SSNState, StateChangeReason)
	OnBroadcast   func(BroadcastType, *SSNEntry)
}

func NewSSNStateManager() *SSNStateManager {
	return &SSNStateManager{
		entries:             make(map[string]*SSNEntry),
		DefaultTestInterval: 30 * time.Second,
		MaxTestInterval:     300 * time.Second,
		MaxTestRetries:      5,
	}
}

func (sm *SSNStateManager) getKey(pc uint16, ssn uint8) string {
	return fmt.Sprintf("%d:%d", pc, ssn)
}

func (sm *SSNStateManager) GetEntry(pc uint16, ssn uint8) *SSNEntry {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.entries[sm.getKey(pc, ssn)]
}

func (sm *SSNStateManager) AddEntry(pc uint16, ssn uint8, isLocal bool) *SSNEntry {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	key := sm.getKey(pc, ssn)
	entry := &SSNEntry{
		SSN:             ssn,
		PointCode:       pc,
		State:           SSNStateProhibited, // Start as prohibited
		IsLocal:         isLocal,
		LastStateChange: time.Now(),
		TestInterval:    sm.DefaultTestInterval,
		MaxTestRetries:  sm.MaxTestRetries,
	}

	sm.entries[key] = entry
	return entry
}

// Message is an interface that defines SCCP messages.
type Message interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
	MarshalTo([]byte) error
	MarshalLen() int
	MessageType() MsgType
	MessageTypeName() string
	fmt.Stringer
}

// ParseMessage decodes the byte sequence into Message by Message Type.
func ParseMessage(b []byte) (Message, error) {
	if len(b) < 1 {
		return nil, fmt.Errorf("invalid SCCP message %v: %w", b, io.ErrUnexpectedEOF)
	}
	var m Message
	switch MsgType(b[0]) {
	/* TODO: implement!
	case MsgTypeCR:
	case MsgTypeCC:
	case MsgTypeCREF:
	case MsgTypeRLSD:
	case MsgTypeRLC:
	case MsgTypeDT1:
	case MsgTypeDT2:
	case MsgTypeAK:
	*/
	case MsgTypeUDT:
		m = &UDT{}

		if err := m.UnmarshalBinary(b); err != nil {
			return nil, fmt.Errorf("failed to parse UDT message: %w", err)
		}

		//validation check
		/*if udt, ok := m.(*UDT); ok && !udt.IsValidForProcessing() {
			return nil, fmt.Errorf("UDT message has invalid protocol class")
		}*/

	/* TODO: implement!
	case MsgTypeUDTS:
	case MsgTypeED:
	case MsgTypeEA:
	case MsgTypeRSR:
	case MsgTypeRSC:
	case MsgTypeERR:
	case MsgTypeIT:
	*/
	case MsgTypeXUDT:
		m = &XUDT{}
	/* TODO: implement!
	case MsgTypeXUDTS:
	case MsgTypeLUDT:
	case MsgTypeLUDTS:
	*/
	default:
		return nil, UnsupportedTypeError(b[0])
	}

	if err := m.UnmarshalBinary(b); err != nil {
		return nil, err
	}
	return m, nil
}

// Global state manager instance
var DefaultSSNStateManager = NewSSNStateManager()
