// Copyright 2019-2024 go-sccp authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package sccp

import (
	"fmt"
	"io"

	"github.com/cgngc/go-sccp/params"
)

// UDT represents a SCCP Message Unit Data (UDT).
type UDT struct {
	Type                MsgType
	ProtocolClass       *params.ProtocolClass
	SLS                 uint8
	CalledPartyAddress  *params.PartyAddress
	CallingPartyAddress *params.PartyAddress
	Data                *params.Data

	ptr1, ptr2, ptr3 uint8
}

// protocol class validation for UDT
func (u *UDT) ValidateProtocolClass() error {
	if u.ProtocolClass == nil {
		return fmt.Errorf("protocol class is nil")
	}

	class := u.ProtocolClass.GetProtocolClass()
	if class != 0 && class != 1 {
		return fmt.Errorf("invalid protocol class %d for UDT message (only 0 and 1 are valid)", class)
	}

	return nil
}

// check if UDT is valid for processing
func (u *UDT) IsValidForProcessing() bool {
	return u.ValidateProtocolClass() == nil
}

// NewUDT creates a new UDT.
/*func NewUDT(pcls int, retOnErr bool, sls uint8, cdpa, cgpa *params.PartyAddress, data []byte) *UDT {

	//protocol class validation
	if pcls != 0 && pcls != 1 {
		logf("Invalid protocol class %d for UDT, only 0 and 1 are valid", pcls)
		return nil
	}

	// Validate addresses for routing
	if cdpa != nil && !cdpa.IsValidForRouting() {
		logf("Warning: Called Party Address may not be valid for %s routing", cdpa.GetRoutingType())
	}

	if cgpa != nil && !cgpa.IsValidForRouting() {
		logf("Warning: Calling Party Address may not be valid for %s routing", cgpa.GetRoutingType())
	}

	u := &UDT{
		Type: MsgTypeUDT,
		//ProtocolClass:       params.NewProtocolClass(pcls, retOnErr),
		//SLS:                 sls,
		CalledPartyAddress:  cdpa,
		CallingPartyAddress: cgpa,
		Data:                params.NewData(data),
	}

	// Validate protocol class
	if err := u.ValidateProtocolClass(); err != nil {
		logf("UDT validation failed: %v", err)
		return nil
	}

	u.ptr1 = 3
	u.ptr2 = u.ptr1 + uint8(cdpa.MarshalLen()) - 1
	u.ptr3 = u.ptr2 + uint8(cgpa.MarshalLen()) - 1

	return u
}
*/
/*func NewUDT(pcls int, retOnErr bool, cdpa, cgpa *params.PartyAddress, data []byte) (*UDT, error) {
	// Validate protocol class
	if pcls != 0 && pcls != 1 {
		return nil, fmt.Errorf("invalid protocol class %d for UDT, only 0 and 1 are valid", pcls)
	}

	// Validate addresses and data
	if cdpa == nil {
		return nil, fmt.Errorf("called party address cannot be nil")
	}
	if cgpa == nil {
		return nil, fmt.Errorf("calling party address cannot be nil")
	}
	if data == nil {
		return nil, fmt.Errorf("data cannot be nil")
	}

	u := &UDT{
		Type:                MsgTypeUDT,
		ProtocolClass:       params.NewProtocolClass(pcls, retOnErr), // Include retOnErr
		CalledPartyAddress:  cdpa,
		CallingPartyAddress: cgpa,
		Data:                params.NewData(data),
	}

	// Calculate pointers
	u.ptr1 = 4
	cdpaLen := cdpa.MarshalLen()
	u.ptr2 = u.ptr1 + uint8(cdpaLen) + 1
	cgpaLen := cgpa.MarshalLen()
	u.ptr3 = u.ptr2 + uint8(cgpaLen) + 1

	fmt.Printf("Successfully created UDT: Class=%d, RetOnErr=%t, CalledLen=%d, CallingLen=%d, DataLen=%d\n",
		pcls, retOnErr, cdpaLen, cgpaLen, len(data))

	return u, nil
}
*/
// MarshalBinary returns the byte sequence generated from a UDT instance.
func (u *UDT) MarshalBinary() ([]byte, error) {
	b := make([]byte, u.MarshalLen())
	if err := u.MarshalTo(b); err != nil {
		return nil, err
	}

	return b, nil
}

// MarshalTo puts the byte sequence in the byte array given as b.
// SCCP is dependent on the Pointers when serializing, which means that it might fail when invalid Pointers are set.
// Enhanced MarshalTo function using library only
func (u *UDT) MarshalTo(b []byte) error {
	l := len(b)
	if l < 6 {
		return io.ErrUnexpectedEOF
	}

	b[0] = uint8(u.Type)

	n := 1
	m, err := u.ProtocolClass.Write(b[1:])
	if err != nil {
		return err
	}
	n += m

	b[n] = u.ptr1
	b[n+1] = u.ptr2
	b[n+2] = u.ptr3
	n += 3

	cdpaEnd := int(u.ptr2 + 3)
	cgpaEnd := int(u.ptr3 + 4)

	if _, err := u.CalledPartyAddress.Write(b[n:cdpaEnd]); err != nil {
		return err
	}

	if _, err := u.CallingPartyAddress.Write(b[cdpaEnd:cgpaEnd]); err != nil {
		return err
	}

	if _, err := u.Data.Write(b[cgpaEnd:]); err != nil {
		return err
	}

	return nil
}

// Enhanced NewUDT with correct pointer calculation
/*func NewUDT(pcls int, retOnErr bool, cdpa, cgpa *params.PartyAddress, data []byte) (*UDT, error) {
	// ... validation code ...

	u := &UDT{
		Type:                MsgTypeUDT,
		ProtocolClass:       params.NewProtocolClass(pcls, retOnErr),
		SLS:                 0,
		CalledPartyAddress:  cdpa,
		CallingPartyAddress: cgpa,
		Data:                params.NewData(data),
	}

	// Use fixed 11-byte address lengths for pointer calculation
	u.ptr1 = 3  // Called Party starts after 3 pointer bytes
	u.ptr2 = 15 // 3 + 11 + 1 = 15 (Called + length byte)
	u.ptr3 = 27 // 15 + 11 + 1 = 27 (Calling + length byte)

	return u, nil
}*/
func NewUDT(pcls int, retOnErr bool, cdpa, cgpa *params.PartyAddress, data []byte) *UDT {
	u := &UDT{
		Type:                MsgTypeUDT,
		ProtocolClass:       params.NewProtocolClass(pcls, retOnErr),
		CalledPartyAddress:  cdpa,
		CallingPartyAddress: cgpa,
		Data:                params.NewData(data),
	}

	u.ptr1 = 3
	u.ptr2 = u.ptr1 + uint8(cdpa.MarshalLen()) - 1
	u.ptr3 = u.ptr2 + uint8(cgpa.MarshalLen()) - 1

	return u
}

// Enhanced MarshalLen calculation
func (u *UDT) MarshalLen() int {
	l := 1 // Message Type
	l += 1 // Protocol Class
	l += 1 // SLS
	l += 3 // Three pointer bytes

	// Called Party Address - use actual address length (11 bytes) not MarshalLen (12 bytes)
	if u.CalledPartyAddress != nil {
		l += 1 + 11 // length byte + 11 address bytes (without library's length prefix)
	}

	// Calling Party Address - same approach
	if u.CallingPartyAddress != nil {
		l += 1 + 11 // length byte + 11 address bytes
	}

	// Data
	if u.Data != nil {
		l += 1 + u.Data.MarshalLen() // length byte + data
	}

	return l
}

/*func (u *UDT) MarshalTo(b []byte) error {
	l := len(b)

	// Get raw data from params.Data
	var dataBytes []byte
	if u.Data != nil {
		// Use the correct method to get data from params.Data
		dataBytes = u.Data.Value() // or u.Data.Data if that's the field name
		// If neither works, try marshaling to a temporary buffer:
		// tempBuf := make([]byte, u.Data.MarshalLen())
		// u.Data.Write(tempBuf)
		// dataBytes = tempBuf
	}

	// Marshal addresses
	var calledBytes, callingBytes []byte
	var err error

	if u.CalledPartyAddress != nil {
		calledBytes = u.CalledPartyAddress.MarshalBinary()
		if err != nil {
			fmt.Printf("Error marshaling called address: %v", err)
			return err
		}
	}

	if u.CallingPartyAddress != nil {
		callingBytes = u.CallingPartyAddress.MarshalBinary()
		if err != nil {
			fmt.Printf("Error marshaling calling address: %v", err)
			return err
		}
	}

	// Calculate required length
	requiredLen := 1 + // Message Type
		1 + len(calledBytes) + // Called address length + data
		1 + len(callingBytes) + // Calling address length + data
		1 + len(dataBytes) // Data length + data

	if l < requiredLen {
		fmt.Printf("Buffer too small: need %d, have %d", requiredLen, l)
		return io.ErrUnexpectedEOF
	}

	offset := 0

	// 1. Write Message Type
	b[offset] = uint8(u.Type)
	offset++

	// 2. Write Called Party Address
	b[offset] = byte(len(calledBytes))
	offset++
	if len(calledBytes) > 0 {
		copy(b[offset:], calledBytes)
		offset += len(calledBytes)
	}

	// 3. Write Calling Party Address
	b[offset] = byte(len(callingBytes))
	offset++
	if len(callingBytes) > 0 {
		copy(b[offset:], callingBytes)
		offset += len(callingBytes)
	}

	// 4. Write Data Buffer
	b[offset] = byte(len(dataBytes))
	offset++
	if len(dataBytes) > 0 {
		copy(b[offset:], dataBytes)
		offset += len(dataBytes)
	}

	fmt.Printf("Successfully encoded UDT: MsgType=%d, CalledLen=%d, CallingLen=%d, DataLen=%d\n",
		u.Type, len(calledBytes), len(callingBytes), len(dataBytes))

	return nil
}*/

/*
	func (u *UDT) MarshalTo(b []byte) error {
		// Validate inputs
		if u == nil || b == nil || len(b) == 0 {
			return fmt.Errorf("invalid input parameters")
		}

		// Use library marshaling for addresses
		var calledBytes, callingBytes []byte
		//var err error

		// Marshal Called Party Address using library
		if u.CalledPartyAddress != nil {
			calledBuf := make([]byte, u.CalledPartyAddress.MarshalLen())
			n, err := u.CalledPartyAddress.Write(calledBuf)
			if err != nil {
				return fmt.Errorf("failed to marshal called party address: %v", err)
			}
			calledBytes = calledBuf[:n]
		}

		// Marshal Calling Party Address using library
		if u.CallingPartyAddress != nil {
			callingBuf := make([]byte, u.CallingPartyAddress.MarshalLen())
			n, err := u.CallingPartyAddress.Write(callingBuf)
			if err != nil {
				return fmt.Errorf("failed to marshal calling party address: %v", err)
			}
			callingBytes = callingBuf[:n]
		}

		// Get data bytes using library
		var dataBytes []byte
		if u.Data != nil && u.Data.MarshalLen() > 0 {
			tempBuf := make([]byte, u.Data.MarshalLen())
			n, err := u.Data.Write(tempBuf)
			if err != nil {
				return fmt.Errorf("failed to write data: %v", err)
			}
			dataBytes = tempBuf[:n]
		}

		// Debug the marshaled components
		fmt.Printf("DEBUG: Called bytes (%d): %x\n", len(calledBytes), calledBytes)
		fmt.Printf("DEBUG: Calling bytes (%d): %x\n", len(callingBytes), callingBytes)
		fmt.Printf("DEBUG: Data bytes (%d): %x\n", len(dataBytes), dataBytes)

		// Calculate pointers
		ptr1 := uint8(3)                            // Called Party starts after 3 pointer bytes
		ptr2 := ptr1 + uint8(len(calledBytes)) + 1  // +1 for length byte
		ptr3 := ptr2 + uint8(len(callingBytes)) + 1 // +1 for length byte

		fmt.Printf("DEBUG: Calculated pointers - ptr1: %d, ptr2: %d, ptr3: %d\n", ptr1, ptr2, ptr3)

		// Calculate total required length
		requiredLen := 1 + // Message Type
			1 + // Protocol Class
			3 + // Three pointer bytes
			1 + len(calledBytes) + // Called address length + data
			1 + len(callingBytes) + // Calling address length + data
			1 + len(dataBytes) // Data length + data

		if len(b) < requiredLen {
			return fmt.Errorf("buffer too small: need %d, have %d", requiredLen, len(b))
		}

		offset := 0

		// 1. Write Message Type
		b[offset] = uint8(u.Type) // Should be 0x09 for UDT
		offset++

		// 2. Write Protocol Class using library
		class := u.ProtocolClass.GetProtocolClass()
		hasReturn := u.ProtocolClass.HasReturnOption()

		protocolClass := uint8(class & 0x0F)
		if hasReturn {
			protocolClass |= 0x80
		}

		b[offset] = protocolClass // Should be 0x81 for Class 1 + Return on Error
		offset++

		// 3. Write Pointers
		b[offset] = ptr1
		b[offset+1] = ptr2
		b[offset+2] = ptr3
		offset += 3

		// 4. Write Called Party Address
		b[offset] = byte(len(calledBytes))
		offset++
		copy(b[offset:], calledBytes)
		offset += len(calledBytes)

		// 5. Write Calling Party Address
		b[offset] = byte(len(callingBytes))
		offset++
		copy(b[offset:], callingBytes)
		offset += len(callingBytes)

		// 6. Write Data
		b[offset] = byte(len(dataBytes))
		offset++
		copy(b[offset:], dataBytes)
		offset += len(dataBytes)

		fmt.Printf("Successfully encoded UDT using library: MsgType=0x%02x, Class=0x%02x, Ptr1=%d, Ptr2=%d, Ptr3=%d\n",
			u.Type, protocolClass, ptr1, ptr2, ptr3)

		return nil
	}
*/
func createSCCPAddress(ai uint8, pc uint16, ssn uint8) []byte {
	var addr []byte

	// Address Indicator
	addr = append(addr, ai)

	// Point Code (2 bytes, little endian)
	addr = append(addr, byte(pc&0xFF))      // Low byte
	addr = append(addr, byte((pc>>8)&0xFF)) // High byte

	// Subsystem Number
	addr = append(addr, ssn)

	return addr
}

// UnmarshalBinary sets the values retrieved from byte sequence in a SCCP UDT.
/*func (u *UDT) UnmarshalBinary(b []byte) error {
	l := len(b)
	if l <= 6 {
		return io.ErrUnexpectedEOF
	}

	u.Type = MsgType(b[0])

	offset := 1
	u.ProtocolClass = &params.ProtocolClass{}
	n, err := u.ProtocolClass.Read(b[offset:])
	if err != nil {
		return err
	}
	offset += n

	// protocol class validation after parsing
	if err := u.ValidateProtocolClass(); err != nil {
		return fmt.Errorf("invalid protocol class in UDT: %w", err)
	}

	//b[n] = u.SLS
	//n += 1 // <-- Increment offset

	u.SLS = b[offset] // Read SLS from byte array
	offset += 1

	u.ptr1 = b[offset]
	offsetPtr1 := 2 + int(u.ptr1)
	if l < offsetPtr1+1 { // where CdPA starts
		return io.ErrUnexpectedEOF
	}
	u.ptr2 = b[offset+1]
	offsetPtr2 := 3 + int(u.ptr2)
	if l < offsetPtr2+1 { // where CgPA starts
		return io.ErrUnexpectedEOF
	}
	u.ptr3 = b[offset+2]
	offsetPtr3 := 4 + int(u.ptr3)
	if l < offsetPtr3+1 { // where u.Data starts
		return io.ErrUnexpectedEOF
	}

	cdpaEnd := offsetPtr1 + int(b[offsetPtr1]) + 1 // +1 is the data length included from the beginning
	if l < cdpaEnd {                               // where CdPA ends
		return io.ErrUnexpectedEOF
	}
	cgpaEnd := offsetPtr2 + int(b[offsetPtr2]) + 1
	if l < cgpaEnd { // where CgPA ends
		return io.ErrUnexpectedEOF
	}
	dataEnd := offsetPtr3 + int(b[offsetPtr3]) + 1
	if l < dataEnd { // where Data ends
		return io.ErrUnexpectedEOF
	}

	u.CalledPartyAddress, _, err = params.ParseCalledPartyAddress(b[offsetPtr1:cdpaEnd])
	if err != nil {
		return err
	}

	u.CallingPartyAddress, _, err = params.ParseCallingPartyAddress(b[offsetPtr2:cgpaEnd])
	if err != nil {
		return err
	}

	u.Data = &params.Data{}
	if _, err := u.Data.Read(b[offsetPtr3:dataEnd]); err != nil {
		return err
	}

	return nil
}*/

// UnmarshalBinary sets the values retrieved from byte sequence in a SCCP UDT.
/*func (u *UDT) UnmarshalBinary(b []byte) error {
	l := len(b)
	if l <= 4 {
		return io.ErrUnexpectedEOF
	}

	u.Type = MsgType(b[0])

	// Start reading pointers from position 1 (no protocol class and SLS)
	offset := 1

	u.ptr1 = b[offset]
	u.ptr2 = b[offset+1]
	u.ptr3 = b[offset+2]

	// Calculate absolute positions - pointers are relative to the pointer section start
	ptrSectionStart := offset
	offsetPtr1 := ptrSectionStart + int(u.ptr1)
	offsetPtr2 := ptrSectionStart + int(u.ptr2)
	offsetPtr3 := ptrSectionStart + int(u.ptr3)

	if l < offsetPtr1+1 { // where CdPA starts
		return io.ErrUnexpectedEOF
	}
	if l < offsetPtr2+1 { // where CgPA starts
		return io.ErrUnexpectedEOF
	}
	if l < offsetPtr3+1 { // where u.Data starts
		return io.ErrUnexpectedEOF
	}

	cdpaEnd := offsetPtr1 + int(b[offsetPtr1]) + 1 // +1 is the data length included from the beginning
	if l < cdpaEnd {                               // where CdPA ends
		return io.ErrUnexpectedEOF
	}
	cgpaEnd := offsetPtr2 + int(b[offsetPtr2]) + 1
	if l < cgpaEnd { // where CgPA ends
		return io.ErrUnexpectedEOF
	}
	dataEnd := offsetPtr3 + int(b[offsetPtr3]) + 1
	if l < dataEnd { // where Data ends
		return io.ErrUnexpectedEOF
	}

	var err error
	u.CalledPartyAddress, _, err = params.ParseCalledPartyAddress(b[offsetPtr1:cdpaEnd])
	if err != nil {
		return err
	}

	u.CallingPartyAddress, _, err = params.ParseCallingPartyAddress(b[offsetPtr2:cgpaEnd])
	if err != nil {
		return err
	}

	u.Data = &params.Data{}
	if _, err = u.Data.Read(b[offsetPtr3:dataEnd]); err != nil {
		return err
	}

	return nil
}*/
func (u *UDT) UnmarshalBinary(b []byte) error {
	l := len(b)
	if l <= 4 {
		return io.ErrUnexpectedEOF
	}

	u.Type = MsgType(b[0])

	offset := 1
	u.ptr1 = b[offset]
	u.ptr2 = b[offset+1]
	u.ptr3 = b[offset+2]

	// Called Party Address
	offsetPtr1 := offset + int(u.ptr1) // 1 + 3 = 4
	if l < offsetPtr1+1 {
		return io.ErrUnexpectedEOF
	}
	cdpaLen := int(b[offsetPtr1])       // 0x0b = 11
	cdpaEnd := offsetPtr1 + cdpaLen + 1 // 4 + 11 + 1 = 16

	// Calling Party Address - starts after Called Party Address
	offsetPtr2 := offset + int(u.ptr2) // 1 + 14 = 15, but this should be relative to start
	// Actually, let's recalculate based on the actual structure
	offsetPtr2 = cdpaEnd // Start right after Called Party Address = 16
	if l < offsetPtr2+1 {
		return io.ErrUnexpectedEOF
	}
	cgpaLen := int(b[offsetPtr2]) // Should be the length at position 16
	cgpaEnd := offsetPtr2 + cgpaLen + 1

	// Data section
	offsetPtr3 := cgpaEnd
	if l < offsetPtr3+1 {
		return io.ErrUnexpectedEOF
	}
	dataLen := int(b[offsetPtr3])
	dataEnd := offsetPtr3 + dataLen + 1

	if l < cdpaEnd || l < cgpaEnd || l < dataEnd {
		return io.ErrUnexpectedEOF
	}

	var err error
	u.CalledPartyAddress, _, err = params.ParseCalledPartyAddress(b[offsetPtr1:cdpaEnd])
	if err != nil {
		return err
	}

	u.CallingPartyAddress, _, err = params.ParseCallingPartyAddress(b[offsetPtr2:cgpaEnd])
	if err != nil {
		return err
	}

	u.Data = &params.Data{}
	if _, err = u.Data.Read(b[offsetPtr3:dataEnd]); err != nil {
		return err
	}

	return nil
}

// MarshalLen returns the serial length.
/*func (u *UDT) MarshalLen() int {
	l := 6 // MsgType, ProtocolClass, pointers

	l += int(u.ptr3) - 1 // length without Data
	if param := u.Data; param != nil {
		l += param.MarshalLen()
	}

	return l
}
*/
/*func (u *UDT) MarshalLen() int {
	l := 1 // Message Type
	l += 1 // Protocol Class - THIS WAS MISSING!
	l += 3 // Three pointer bytes

	// Called Party Address
	if u.CalledPartyAddress != nil {
		l += 1 + u.CalledPartyAddress.MarshalLen() // length byte + address
	} else {
		l += 1 // Just length byte (0)
	}

	// Calling Party Address
	if u.CallingPartyAddress != nil {
		l += 1 + u.CallingPartyAddress.MarshalLen() // length byte + address
	} else {
		l += 1 // Just length byte (0)
	}

	// Data
	if u.Data != nil {
		l += 1 + u.Data.MarshalLen() // length byte + data
	} else {
		l += 1 // Just length byte (0)
	}

	return l
}
*/
// String returns the UDT values in human readable format.
func (u *UDT) String() string {
	return fmt.Sprintf("%s: {ProtocolClass: %s, SLS: %d, CalledPartyAddress: %v, CallingPartyAddress: %v, Data: %s}",
		u.Type,
		u.ProtocolClass,
		u.SLS,
		u.CdAddress(),
		u.CgAddress(),
		u.Data,
	)
}

// MessageType returns the Message Type in int.
func (u *UDT) MessageType() MsgType {
	return MsgTypeUDT
}

// MessageTypeName returns the Message Type in string.
func (u *UDT) MessageTypeName() string {
	return u.MessageType().String()
}

// CdGT returns the GT in CalledPartyAddress in human readable string.
func (u *UDT) CdGT() string {
	if u.CalledPartyAddress == nil {
		return ""
	}
	return u.CalledPartyAddress.Address()
}

// CgGT returns the GT in CalledPartyAddress in human readable string.
func (u *UDT) CgGT() string {
	if u.CallingPartyAddress == nil {
		return ""
	}
	return u.CallingPartyAddress.Address()
}

func (u *UDT) CdAddress() string {
	if u.CalledPartyAddress == nil {
		return ""
	}
	return u.CalledPartyAddress.AddressWithDetails()
}

func (u *UDT) CgAddress() string {
	if u.CallingPartyAddress == nil {
		return ""
	}
	return u.CallingPartyAddress.AddressWithDetails()
}

// method to get protocol class info
func (u *UDT) GetProtocolClassInfo() (class int, hasReturnOption bool) {
	if u.ProtocolClass == nil {
		return 0, false
	}
	return u.ProtocolClass.GetProtocolClass(), u.ProtocolClass.HasReturnOption()
}
