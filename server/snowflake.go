package server

import (
	"fmt"
	"sync/atomic"
	"time"
)

const (
	epoch        = 1491696000000
	serverBits   = 10
	sequenceBits = 12
	timeBits     = 42
	serverShift  = sequenceBits
	timeShift    = sequenceBits + serverBits
	serverMax    = ^(-1 << serverBits)
	sequenceMask = ^(-1 << sequenceBits)
	timeMask     = ^(-1 << timeBits)
)

type Generator struct {
	state   uint64
	machine uint64
}

func NewGenerator(machineID int) *Generator {
	if machineID < 0 || machineID > serverMax {
		panic(fmt.Errorf("invalid machine id; must be 0 ≤ id < %d", serverMax))
	}
	return &Generator{
		state:   0,
		machine: uint64(machineID << serverShift),
	}
}

func (g *Generator) MachineID() int {
	return int(g.machine >> serverShift)
}

func (g *Generator) Next() uint64 {
	var state uint64

	// we attempt 100 times to update the millisecond part of the state
	// and increment the sequence atomically. each attempt is approx ~30ns
	// so we spend around ~3µs total.
	for i := 0; i < 100; i++ {
		t := (now() - epoch) & timeMask
		current := atomic.LoadUint64(&g.state)
		currentTime := current >> timeShift & timeMask
		currentSeq := current & sequenceMask

		// this sequence of conditionals ensures a monotonically increasing
		// state.

		switch {
		// if our time is in the future, use that with a zero sequence number.
		case t > currentTime:
			state = t << timeShift

		// we now know that our time is at or before the current time.
		// if we're at the maximum sequence, bump to the next millisecond
		case currentSeq == sequenceMask:
			state = (currentTime + 1) << timeShift

		// otherwise, increment the sequence.
		default:
			state = current + 1
		}

		if atomic.CompareAndSwapUint64(&g.state, current, state) {
			break
		}

		state = 0
	}

	// since we failed 100 times, there's high contention. bail out of the
	// loop to bound the time we'll spend in this method, and just add
	// one to the counter. this can cause millisecond drift, but hopefully
	// some CAS eventually succeeds and fixes the milliseconds. additionally,
	// if the sequence is already at the maximum, adding 1 here can cause
	// it to roll over into the machine id. giving the CAS 100 attempts
	// helps to avoid these problems.
	if state == 0 {
		state = atomic.AddUint64(&g.state, 1)
	}

	return state | g.machine
}

func now() uint64 { return uint64(time.Now().UnixNano() / 1e6) }
