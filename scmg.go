// Copyright 2019-2024 go-sccp authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package sccp

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

// SCMGType is type of SCMG message.
type SCMGType uint8

// Table 23/Q.713
const (
	_           SCMGType = iota
	SCMGTypeSSA          // SSA
	SCMGTypeSSP          // SSP
	SCMGTypeSST          // SST
	SCMGTypeSOR          // SOR
	SCMGTypeSOG          // SOG
	SCMGTypeSSC          // SSC
)

// SCMG represents a SCCP Management message (SCMG).
// Chapter 5.3/Q.713
type SCMG struct {
	Type                           SCMGType
	AffectedSSN                    uint8
	AffectedPC                     uint16
	SubsystemMultiplicityIndicator uint8
	SCCPCongestionLevel            uint8
}

// NewSCMG creates a new SCMG.
func NewSCMG(typ SCMGType, assn uint8, apc uint16, smi uint8, scl uint8) *SCMG {
	return &SCMG{
		Type:                           typ,
		AffectedSSN:                    assn,
		AffectedPC:                     apc,
		SubsystemMultiplicityIndicator: smi,
		SCCPCongestionLevel:            scl,
	}
}

// MarshalBinary returns the byte sequence generated from a SCMG instance.
func (s *SCMG) MarshalBinary() ([]byte, error) {
	b := make([]byte, s.MarshalLen())
	if err := s.MarshalTo(b); err != nil {
		return nil, err
	}

	return b, nil
}

// MarshalTo puts the byte sequence in the byte array given as b.
func (s *SCMG) MarshalTo(b []byte) error {
	l := len(b)

	if l < s.MarshalLen() {
		return io.ErrUnexpectedEOF
	}

	b[0] = uint8(s.Type)
	b[1] = s.AffectedSSN
	binary.LittleEndian.PutUint16(b[2:4], s.AffectedPC)
	b[4] = s.SubsystemMultiplicityIndicator
	if s.Type == SCMGTypeSSC {
		b[5] = s.SCCPCongestionLevel
	}

	return nil
}

// ParseSCMG decodes given byte sequence as a SCMG.
func ParseSCMG(b []byte) (*SCMG, error) {
	s := &SCMG{}
	if err := s.UnmarshalBinary(b); err != nil {
		return nil, err
	}

	return s, nil
}

// UnmarshalBinary sets the values retrieved from byte sequence in a SCMG.
func (s *SCMG) UnmarshalBinary(b []byte) error {
	l := len(b)
	if l < 5 {
		return io.ErrUnexpectedEOF
	}

	s.Type = SCMGType(b[0])
	s.AffectedSSN = b[1]
	s.AffectedPC = binary.LittleEndian.Uint16(b[2:4])
	s.SubsystemMultiplicityIndicator = b[4]

	if s.Type == SCMGTypeSSC {
		if l < 6 {
			return io.ErrUnexpectedEOF
		}
		s.SCCPCongestionLevel = b[5]
	}

	return nil
}

// MarshalLen returns the serial length.
func (s *SCMG) MarshalLen() int {
	// Table 24/Q.713 – SCMG messages
	l := 5

	// Table 25/Q.713 – SSC
	if s.Type == SCMGTypeSSC {
		l += 1
	}

	return l
}

// String returns the SCMG values in human readable format.
func (s *SCMG) String() string {
	return fmt.Sprintf("%s: {AffectedSSN: %v, AffectedPC: %v, SubsystemMultiplicityIndicator: %d, SCCPCongestionLevel: %d}",
		s.Type,
		s.AffectedSSN,
		s.AffectedPC,
		s.SubsystemMultiplicityIndicator,
		s.SCCPCongestionLevel,
	)
}

// MessageType returns the Message Type in int.
func (s *SCMG) MessageType() SCMGType {
	return s.Type
}

// MessageTypeName returns the Message Type in string.
func (s *SCMG) MessageTypeName() string {
	return s.Type.String()
}

// SSN State Management Methods for SSNStateManager

// HandleUserInService - Handle N-STATE Request with UIS
func (sm *SSNStateManager) HandleUserInService(pc uint16, ssn uint8) error {
	entry := sm.GetEntry(pc, ssn)
	if entry == nil {
		return fmt.Errorf("SSN entry not found: PC=%d, SSN=%d", pc, ssn)
	}

	if !entry.IsLocal {
		return fmt.Errorf("cannot change state of remote subsystem")
	}

	if entry.IsProhibited() {
		entry.MarkAllowed()

		// Trigger callbacks
		if sm.OnStateChange != nil {
			sm.OnStateChange(entry, SSNStateAllowed, ReasonUserInitiated)
		}

		// Broadcast SSA to network
		if sm.OnBroadcast != nil {
			sm.OnBroadcast(BroadcastSSA, entry)
		}

		logf("Local subsystem allowed: PC=%d, SSN=%d", pc, ssn)
	}

	return nil
}

// HandleUserOutOfService - Handle N-STATE Request with UOS
func (sm *SSNStateManager) HandleUserOutOfService(pc uint16, ssn uint8) error {
	entry := sm.GetEntry(pc, ssn)
	if entry == nil {
		return fmt.Errorf("SSN entry not found: PC=%d, SSN=%d", pc, ssn)
	}

	if !entry.IsLocal {
		return fmt.Errorf("cannot change state of remote subsystem")
	}

	if entry.IsAllowed() {
		entry.MarkProhibited()

		// Stop any ongoing tests
		sm.stopSST(entry)

		// Trigger callbacks
		if sm.OnStateChange != nil {
			sm.OnStateChange(entry, SSNStateProhibited, ReasonUserInitiated)
		}

		// Broadcast SSP to network
		if sm.OnBroadcast != nil {
			sm.OnBroadcast(BroadcastSSP, entry)
		}

		logf("Local subsystem prohibited: PC=%d, SSN=%d", pc, ssn)
	}

	return nil
}

// HandleSSA - Handle remote Subsystem Allowed message
func (sm *SSNStateManager) HandleSSA(pc uint16, ssn uint8) error {
	entry := sm.GetEntry(pc, ssn)
	if entry == nil {
		entry = sm.AddEntry(pc, ssn, false)
	}

	if entry.IsProhibited() {
		entry.MarkAllowed()

		// Stop subsystem testing
		sm.stopSST(entry)

		// Trigger callbacks
		if sm.OnStateChange != nil {
			sm.OnStateChange(entry, SSNStateAllowed, ReasonNetworkInitiated)
		}

		logf("Remote subsystem allowed: PC=%d, SSN=%d", pc, ssn)
	}

	return nil
}

// HandleSSP - Handle remote Subsystem Prohibited message
func (sm *SSNStateManager) HandleSSP(pc uint16, ssn uint8) error {
	entry := sm.GetEntry(pc, ssn)
	if entry == nil {
		entry = sm.AddEntry(pc, ssn, false)
	}

	if entry.IsAllowed() {
		entry.MarkProhibited()

		// Start subsystem testing
		sm.startSST(entry)

		// Trigger callbacks
		if sm.OnStateChange != nil {
			sm.OnStateChange(entry, SSNStateProhibited, ReasonNetworkInitiated)
		}

		logf("Remote subsystem prohibited: PC=%d, SSN=%d", pc, ssn)
	}

	return nil
}

// HandleSST - Handle Subsystem Test message
func (sm *SSNStateManager) HandleSST(pc uint16, ssn uint8) error {
	entry := sm.GetEntry(pc, ssn)
	if entry == nil {
		entry = sm.AddEntry(pc, ssn, false)
	}
	// If local subsystem is available, respond with SSA
	if entry.IsLocal && entry.IsAllowed() {
		// Send SSA response
		ssa := NewSCMG(SCMGTypeSSA, ssn, pc, 0, 0)
		ssaBytes, err := ssa.MarshalBinary()
		if err != nil {
			return fmt.Errorf("failed to marshal SSA response: %w", err)
		}

		logf("Responding to SST with SSA for PC=%d, SSN=%d: %x", pc, ssn, ssaBytes)
		// TODO: Send SSA response over network
	} else if entry.IsLocal && entry.IsProhibited() {
		// Local subsystem is prohibited, don't respond (let SST timeout)
		logf("Local subsystem prohibited, not responding to SST: PC=%d, SSN=%d", pc, ssn)
	}

	return nil
}

// startSST - Start subsystem testing with exponential backoff
func (sm *SSNStateManager) startSST(entry *SSNEntry) {
	if entry.IsLocal {
		return // Don't test local subsystems
	}

	entry.mutex.Lock()
	defer entry.mutex.Unlock()

	// Stop existing timer
	if entry.TestTimer != nil {
		entry.TestTimer.Stop()
	}

	// Reset retry count and interval
	entry.TestRetries = 0
	entry.TestInterval = sm.DefaultTestInterval

	// Start testing
	sm.scheduleSST(entry)
	logf("Started SST for PC=%d, SSN=%d", entry.PointCode, entry.SSN)
}

// stopSST - Stop subsystem testing
func (sm *SSNStateManager) stopSST(entry *SSNEntry) {
	entry.mutex.Lock()
	defer entry.mutex.Unlock()

	if entry.TestTimer != nil {
		entry.TestTimer.Stop()
		entry.TestTimer = nil
	}

	entry.TestRetries = 0
	logf("Stopped SST for PC=%d, SSN=%d", entry.PointCode, entry.SSN)
}

// scheduleSST - Schedule next SST message
func (sm *SSNStateManager) scheduleSST(entry *SSNEntry) {
	entry.TestTimer = time.AfterFunc(entry.TestInterval, func() {
		sm.performSST(entry)
	})
}

// performSST - Perform actual SST with exponential backoff
func (sm *SSNStateManager) performSST(entry *SSNEntry) {
	entry.mutex.Lock()
	defer entry.mutex.Unlock()

	if entry.IsAllowed() {
		return // Subsystem became available, stop testing
	}

	// Send SST message
	if err := sm.sendSST(entry.PointCode, entry.SSN); err != nil {
		logf("Failed to send SST: %v", err)
	}

	entry.TestRetries++

	if entry.TestRetries >= entry.MaxTestRetries {
		logf("Max SST retries reached for PC=%d, SSN=%d", entry.PointCode, entry.SSN)
		return
	}

	// Exponential backoff
	entry.TestInterval *= 2
	if entry.TestInterval > sm.MaxTestInterval {
		entry.TestInterval = sm.MaxTestInterval
	}

	// Schedule next test
	sm.scheduleSST(entry)
}

// sendSST - Send SST SCMG message
func (sm *SSNStateManager) sendSST(pc uint16, ssn uint8) error {
	// Create SST SCMG message
	sst := NewSCMG(SCMGTypeSST, ssn, pc, 0, 0)

	// TODO: Integrate with your SCCP message sending mechanism
	// This is where you would send the SST message over your M3UA/SCTP connection
	logf("Sending SST to PC=%d, SSN=%d", pc, ssn)

	// For now, just log the message that would be sent
	sstBytes, err := sst.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to marshal SST: %w", err)
	}

	logf("SST message bytes: %x", sstBytes)
	return nil
}

// ProcessSCMGMessage - Process incoming SCMG messages
func (sm *SSNStateManager) ProcessSCMGMessage(scmg *SCMG) error {
	switch scmg.Type {
	case SCMGTypeSSA:
		return sm.HandleSSA(scmg.AffectedPC, scmg.AffectedSSN)
	case SCMGTypeSSP:
		return sm.HandleSSP(scmg.AffectedPC, scmg.AffectedSSN)
	case SCMGTypeSST:
		return sm.HandleSST(scmg.AffectedPC, scmg.AffectedSSN)
	default:
		logf("Unhandled SCMG message type: %v", scmg.Type)
		return nil
	}
}
