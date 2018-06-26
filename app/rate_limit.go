// Copyright (c) 2017-2018 Townsourced Inc.
package app

import (
	"net/http"
	"sync"
	"time"
)

//TODO: This rate limiting is incorrect when used with multiple webservers
//  Add clustering settings for lex library

// Attempter is an interface that allows for multiple different rate limit types
type Attempter interface {
	Attempt(id string) (RateLeft, error)
}

type rateKey struct {
	id       string
	rateType string
}

var rates = struct {
	sync.RWMutex
	left map[rateKey]RateLeft
}{left: make(map[rateKey]RateLeft)}

// TODO: remove expired rates that haven't been used in a while?

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
	left := key.attempt(rl.Limit, time.Now().Add(rl.Period))
	if left.Remaining < 0 {
		return left, ErrTooManyRequests
	}
	return left, nil
}

// Attempt checks to see if a request is rate delayed. Will return an error if it is delayed the max amount
func (rd *RateDelay) Attempt(id string) (RateLeft, error) {
	key := rateKey{id: id, rateType: rd.Type}
	left := key.attempt(rd.Limit, time.Now().Add(rd.Period))
	if left.Remaining < 0 {
		delay := rd.Delay * (time.Duration(-1*left.Remaining) + 1)
		if delay >= rd.Max {
			return left, ErrTooManyRequests
		}

		time.Sleep(delay)
	}
	return left, nil
}

func (rk rateKey) attempt(limit int32, reset time.Time) RateLeft {
	rates.Lock()
	defer rates.Unlock()

	left, ok := rates.left[rk]
	if !ok || left.Reset.Before(time.Now()) {
		// add / update to fresh rate entry
		left = RateLeft{
			Limit:     limit,
			Reset:     reset,
			Remaining: limit,
		}
	}

	left.Remaining--

	rates.left[rk] = left
	return left
}
