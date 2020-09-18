// Copyright 2019 - MinIO, Inc. All rights reserved.
// Use of this source code is governed by the AGPLv3
// license that can be found in the LICENSE file.

package kes

import (
	"bufio"
	"encoding/json"
	"io"
	"time"
)

// NewErrorStream returns an new ErrorStream that
// splits r into lines and tries to parse each
// line as JSON-encoded ErrorEvent.
func NewErrorStream(r io.Reader) *ErrorStream {
	s := &ErrorStream{
		scanner: bufio.NewScanner(r),
	}
	if closer, ok := r.(io.Closer); ok {
		s.closer = closer
	}
	return s
}

// ErrorStream provides a convenient interface for
// iterating over a stream of ErrorEvents. Successive
// calls to the Next method will step through the error
// events of an io.Reader.
//
// By default, the ErrorStream breaks the underlying
// stream into lines and expects a JSON-encoded ErrorEvent
// per line - unless the line is empty. Empty lines will
// be ignored.
//
// Iterating stops at the end of the stream, the first I/O
// error, a ErrorEvent event too large to fit in the buffer,
// or when the stream gets closed.
//
// Closing an ErrorStream closes the underlying io.Reader,
// if it implements io.Closer, and any subsequent call to
// Next will return false.
type ErrorStream struct {
	scanner *bufio.Scanner

	event ErrorEvent
	err   error

	closer io.Closer
	closed bool
}

// Err returns the first non-EOF error that was encountered
// while iterating over the stream and un-marshaling ErrorEvents.
//
// Err does not return any error returned from Close.
func (s *ErrorStream) Err() error { return s.err }

// Event returns the most recent ErrorEvent generated by a
// call to Next.
func (s *ErrorStream) Event() ErrorEvent { return s.event }

// Bytes returns the most recent raw ErrorEvent content generated
// by a call to Next. It may not contain valid JSON.
//
// The underlying array may point to data that will be overwritten
// by a subsequent call to Next. It does no allocation.
func (s *ErrorStream) Bytes() []byte { return s.scanner.Bytes() }

// Next advances the stream to the next ErrorEvent, which will then
// be available through the Event and Bytes method. It returns false
// when the stream iteration stops - i.e. by reaching the end of the
// stream, closing the stream or in case of an error.
// After Next returns false, the Err method will return any error that
// occurred while iterating and parsing the stream.
func (s *ErrorStream) Next() bool {
	if s.err != nil || s.closed {
		return false
	}

	// Iterate over the stream until we find a non-empty line.
	for {
		if !s.scanner.Scan() {
			if !s.closed { // Once the stream is closed we ignore the error
				s.err = s.scanner.Err()
			}
			return false
		}
		if len(s.scanner.Bytes()) != 0 {
			break
		}
	}
	if err := json.Unmarshal(s.scanner.Bytes(), &s.event); err != nil {
		if !s.closed { // Once the stream is closed we ignore the error
			s.err = err
		}
		return false
	}
	return true
}

// Close closes the underlying stream - i.e. the io.Reader if
// if implements io.Closer. After Close has been called once
// the Next method will return false.
func (s *ErrorStream) Close() (err error) {
	if s.closer != nil {
		s.closed = true
		err = s.closer.Close()
	}
	return err
}

// ErrorEvent is the event type the KES server produces when it
// encounters and logs an error.
//
// When a clients subscribes to the KES server error log it
// receives a stream of JSON-encoded error events separated
// by a newline.
type ErrorEvent struct {
	Message string `json:"message"` // The logged error message
}

// NewAuditStream returns a new AuditStream that
// splits r into lines and tries to parse each
// line as JSON-encoded AuditEvent.
func NewAuditStream(r io.Reader) *AuditStream {
	s := &AuditStream{
		scanner: bufio.NewScanner(r),
	}
	if closer, ok := r.(io.Closer); ok {
		s.closer = closer
	}
	return s
}

// AuditStream provides a convenient interface for
// iterating over a stream of AuditEvents. Successive
// calls to the Next method will step through the audit
// events of an io.Reader.
//
// By default, the AuditStream breaks the underlying
// stream into lines and expects a JSON-encoded AuditEvent
// per line - unless the line is empty. Empty lines will
// be ignored.
//
// Iterating stops at the end of the stream, the first I/O
// error, an AuditEvent event too large to fit in the buffer,
// or when the stream gets closed.
//
// Closing an AuditStream closes the underlying io.Reader,
// if it implements io.Closer, and any subsequent call to
// Next will return false.
type AuditStream struct {
	scanner *bufio.Scanner

	event AuditEvent
	err   error

	closer io.Closer
	closed bool
}

// Err returns the first non-EOF error that was encountered
// while iterating over the stream and un-marshaling AuditEvents.
//
// Err does not return any error returned from Close.
func (s *AuditStream) Err() error { return s.err }

// Event returns the most recent AuditEvent generated by a
// call to Next.
func (s *AuditStream) Event() AuditEvent { return s.event }

// Bytes returns the most recent raw AuditEvent content generated
// by a call to Next. It may not contain valid JSON.
//
// The underlying array may point to data that will be overwritten
// by a subsequent call to Next. It does no allocation.
func (s *AuditStream) Bytes() []byte { return s.scanner.Bytes() }

// Next advances the stream to the next AuditEvent, which will then
// be available through the Event and Bytes method. It returns false
// when the stream iteration stops - i.e. by reaching the end of the
// stream, closing the stream or in case of an error.
// After Next returns false, the Err method will return any error that
// occurred while iterating and parsing the stream.
func (s *AuditStream) Next() bool {
	if s.err != nil || s.closed {
		return false
	}

	// Iterate over the stream until we find a non-empty line.
	for {
		if !s.scanner.Scan() {
			if !s.closed { // Once the stream is closed we ignore the error
				s.err = s.scanner.Err()
			}
			return false
		}
		if len(s.scanner.Bytes()) != 0 {
			break
		}
	}

	if err := json.Unmarshal(s.scanner.Bytes(), &s.event); err != nil {
		if !s.closed { // Once the stream is closed we ignore the error
			s.err = err
		}
		return false
	}
	return true
}

// Close closes the underlying stream - i.e. the io.Reader if
// if implements io.Closer. After Close has been called once
// the Next method will return false.
func (s *AuditStream) Close() (err error) {
	if s.closer != nil {
		s.closed = true
		err = s.closer.Close()
	}
	return err
}

// AuditEvent is the event type the KES server produces when it
// has handled a request right before responding to the client.
//
// When a clients subscribes to the KES server audit log it
// receives a stream of JSON-encoded audit events separated
// by a newline.
type AuditEvent struct {
	// Time is the point in time when the
	// audit event has been created.
	Time time.Time `json:"time"`

	// Request contains audit log information
	// about the request received from a client.
	Request AuditEventRequest `json:"request"`

	// Response contains audit log information
	// about the response sent to the client.
	Response AuditEventResponse `json:"response"`
}

// AuditEventRequest contains the audit information
// about a request sent by a client to a KES server.
//
// In particular, it contains the identity of the
// client and other audit-related information.
type AuditEventRequest struct {
	Path     string `json:"path"`
	Identity string `json:"identity"`
}

// AuditEventResponse contains the audit information
// about a response sent to a client by a KES server.
//
// In particular, it contains the response status code
// and other audit-related information.
type AuditEventResponse struct {
	StatusCode int           `json:"code"`
	Time       time.Duration `json:"time"`
}
