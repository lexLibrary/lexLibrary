// Copyright (c) 2018 Townsourced Inc.
package app

import (
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

//TODO: This rate limiting is incorrect when used with multiple webservers
//  Add clustering settings for lex library

type rateKey struct {
	id       string
	rateType string
}

//TODO: periodically clear out expired rates
var rates = struct {
	sync.RWMutex
	r map[rateKey]*RateLeft
}{r: make(map[rateKey]*RateLeft)}

// RateLimit Fails immediately when the limit is reached within the Reset period
type RateLimit struct {
	Type   string
	Limit  int32
	Period time.Duration
}

// RateLeft is the status of a given user's rate limit
type RateLeft struct {
	Limit     int32
	Remaining int32
	Reset     time.Time
}

// RateDelay delays access if it is being called too frequently
// once Limit is reached within Period, Delay * (amount past Limit) is added to each request up to Max
// when Max is reached, an error is returned
type RateDelay struct {
	Type   string
	Limit  int32
	Delay  time.Duration
	Period time.Duration
	Max    time.Duration
}

// ErrTooManyRequests is the failure return when a user is being rate limited
var ErrTooManyRequests = NewFailureWithStatus("Too many requests", http.StatusTooManyRequests)

// Attempt checks to see if a request is rate limited. Will return an error if it is
func (rl *RateLimit) Attempt(id string) (RateLeft, error) {
	key := rateKey{id: id, rateType: rl.Type}
	result, ok := key.find()
	if ok {
		if result.Remaining <= 0 && result.Reset.After(time.Now()) {
			return *result, ErrTooManyRequests
		}
		result.decrement()

		return *result, nil
	}
	return key.add(rl.Limit, time.Now().Add(rl.Period)), nil
}

// Attempt checks to see if a request is rate delayed. Will return an error if it is delayed the max amount
func (rd *RateDelay) Attempt(id string) error {
	key := rateKey{id: id, rateType: rd.Type}
	result, ok := key.find()
	if ok {
		if result.Remaining <= 0 && result.Reset.After(time.Now()) {
			delay := rd.Delay * (time.Duration(-1*result.Remaining) + 1)
			if delay >= rd.Max {
				return ErrTooManyRequests
			}

			result.decrement()
			time.Sleep(delay)
			return nil
		}
		result.decrement()

		return nil
	}
	key.add(rd.Limit, time.Now().Add(rd.Period))
	return nil
}

func (rk *rateKey) find() (*RateLeft, bool) {
	rates.RLock()
	defer rates.RUnlock()

	result, ok := rates.r[*rk]
	return result, ok
}

func (rk *rateKey) add(limit int32, reset time.Time) RateLeft {
	rates.Lock()
	defer rates.Unlock()
	rr := &RateLeft{
		Limit:     limit,
		Reset:     reset,
		Remaining: limit,
	}

	rates.r[*rk] = rr
	return *rr
}

func (rl *RateLeft) decrement() {
	remains := atomic.AddInt32(&rl.Remaining, -1)
	rl.Remaining = remains
}
