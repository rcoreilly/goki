// Copyright (c) 2018, Randall C. O'Reilly. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ki

import (
	"errors"
	"fmt"
	"reflect"
)

// implements general signal passing between Ki objects, like Qt's Signal / Slot system
// started from: github.com/tucnak/meta/

// SignalType provides standard signals -- can extend by starting at iota + last signal here
type SignalType int64

const (
	NilSignal             SignalType = iota
	SignalChildAdded      SignalType = iota // data is the added child
	SignalChildDeleted    SignalType = iota // data is deleted child
	SignalChildrenDeleted SignalType = iota // no data
	SignalNodeUpdated     SignalType = iota // entire node updated
	SignalFieldUpdated    SignalType = iota // a field was updated -- data is name of field
)

// generates signaltype_string.go -- contrary to some docs, apparently need to run go generate manually
//go:generate stringer -type=SignalType

// Receiver function type on receiver node -- gets the sending node and arbitrary additional data
type RecvFun func(receiver, sender Ki, sig SignalType, data interface{})

// Signal -- add one of these for each signal a node can emit
type Signal struct {
	DefSig SignalType
	Cons   []Connection
}

// Connection represents one connection between a signal and a receiving Ki and function to call
type Connection struct {
	// node that will receive the signal
	Recv Ki
	// function on the receiver node that will receive the signal
	Fun RecvFun
}

// Connect attaches a new receiver to the signal -- error if not ok
func (sig *Signal) Connect(recv Ki, recvfun RecvFun) error {
	if recv == nil {
		return errors.New("ki Signal Connect: no recv node provided")
	}
	if recvfun == nil {
		return errors.New("ki Signal Connect: no recv func provided")
	}

	con := Connection{
		Recv: recv,
		Fun:  recvfun,
	}
	sig.Cons = append(sig.Cons, con)

	// fmt.Printf("added connection to recv %v fun %v", recv.KiName(), reflect.ValueOf(recvfun))

	return nil
}

// Disconnect receiver and signal
func (sig *Signal) Disconnect(recv Ki, recvfun RecvFun) error {
	if recv == nil {
		return errors.New("ki Signal Disconnect: no recv node provided")
	}
	if recvfun == nil {
		return errors.New("ki Signal Disconnect: no recv func provided")
	}

	for i, con := range sig.Cons {
		if con.Recv == recv /* && con.Fun == recvfun */ {
			// this copy makes sure there are no memory leaks
			copy(sig.Cons[i:], sig.Cons[i+1:])
			sig.Cons = sig.Cons[:len(sig.Cons)-1]
			return nil
		}
	}
	return errors.New(fmt.Sprintf("ki Signal Disconnect: connection not found for node: %v func: %v", recv.KiName(), reflect.ValueOf(recvfun)))
}

// Emit sends the signal across all the connections to the receivers -- sequential
func (s *Signal) Emit(sender Ki, sig SignalType, data interface{}) {
	if sig == NilSignal && s.DefSig != NilSignal {
		sig = s.DefSig
	}
	for _, con := range s.Cons {
		con.Fun(con.Recv, sender, sig, data)
	}
}

// EmitGo concurrent version -- sends the signal across all the connections to the receivers
func (s *Signal) EmitGo(sender Ki, sig SignalType, data interface{}) {
	if sig == NilSignal && s.DefSig != NilSignal {
		sig = s.DefSig
	}
	for _, con := range s.Cons {
		go con.Fun(con.Recv, sender, sig, data)
	}
}
