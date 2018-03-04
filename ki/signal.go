// Copyright (c) 2018, Randall C. O'Reilly. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ki

import (
	"errors"
	"fmt"
	"log"
	"reflect"
)

// this implemeents general signal passing between Ki objects, like Qt's Signal / Slot system

// started from: github.com/tucnak/meta/

// standard signals -- can extend by starting at iota + last signal here

type SignalType int64

const (
	NoSignal            SignalType = iota
	SignalChildAdded    SignalType = iota
	SignalChildRemoved  SignalType = iota
	SignalChildrenReset SignalType = iota
	SignalFieldUpdated  SignalType = iota // a field was updated -- data typically name of field
	SignalNodeUpdated   SignalType = iota // entire node updated
)

// Receiver function type on receiver node -- gets the sending node and arbitrary additional data
type RecvFunc func(receiver, sender Ki, sig SignalType, data interface{})

// Signal -- add one of these for each signal a node can emit
type Signal struct {
	DefSig SignalType
	Cons   []Connection
}

type Connection struct {
	// node that will receive the signal
	Recv Ki
	// function on the receiver node that will receive the signal
	Func RecvFunc
}

// Connect attaches a new receiver to the signal -- error if not ok
func (sig *Signal) Connect(recv Ki, recvfun RecvFunc) error {
	if recv == nil {
		return errors.New("ki Signal Connect: no recv node provided")
	}
	if recvfun == nil {
		return errors.New("ki Signal Connect: no recv func provided")
	}

	con := Connection{
		Recv: recv,
		Func: recvfun,
	}
	sig.Cons = append(sig.Cons, con)

	log.Printf("added connection to recv %v fun %v", recv.KiName(), reflect.ValueOf(recvfun))

	return nil
}

// Disconnect receiver and signal
func (sig *Signal) Disconnect(recv Ki, recvfun RecvFunc) error {
	if recv == nil {
		return errors.New("ki Signal Disconnect: no recv node provided")
	}
	if recvfun == nil {
		return errors.New("ki Signal Disconnect: no recv func provided")
	}

	for i, con := range sig.Cons {
		if con.Recv == recv /* && con.Func == recvfun */ {
			// this copy makes sure there are no memory leaks
			copy(sig.Cons[i:], sig.Cons[i+1:])
			sig.Cons = sig.Cons[:len(sig.Cons)-1]
			return nil
		}
	}
	return errors.New(fmt.Sprintf("ki Signal Disconnect: connection not found for node: %v func: %v", recv.KiName(), reflect.ValueOf(recvfun)))
}

// Emit executes all the connected slots with data given.
func (s *Signal) Emit(sender Ki, sig SignalType, data interface{}) {
	if sig == NoSignal && s.DefSig != NoSignal {
		sig = s.DefSig
	}
	for _, con := range s.Cons {
		go con.Func(con.Recv, sender, sig, data)
	}
}

// Emit given signal on node
func (n *Node) Emit(s *Signal, sig SignalType, data interface{}) {
	s.Emit(n, sig, data)
}
