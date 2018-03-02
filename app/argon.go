// Copyright (c) 2017-2018 Townsourced Inc.

package app

import (
	"crypto/rand"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/argon2"
)

const passwordDelim = "$"

// argon uses the argon2 hash against a sha515 hashed password to prevent DOS attacks
// version, salt and work factors are stored with the hash
type argon struct {
	time    uint32
	memory  uint32
	threads uint8
	keyLen  uint32
}

func (a *argon) hash(password string) ([]byte, error) {
	maxSaltSize := 16

	salt := make([]byte, maxSaltSize)
	_, err := io.ReadFull(rand.Reader, salt)
	if err != nil {
		return nil, err
	}

	return a.hashWithSalt(password, salt)
}

func (a *argon) hashWithSalt(password string, salt []byte) ([]byte, error) {
	sha := sha512.Sum512_256([]byte(password))

	hash := argon2.Key(sha[:], salt, a.time, a.memory, a.threads, a.keyLen)

	// version$time$memory$threads$keyLen$salt$hash
	// salt := unencodedSalt)
	result := strings.Join([]string{
		fmt.Sprintf("%02d", argon2.Version),
		fmt.Sprintf("%02d", a.time),
		fmt.Sprintf("%02d", a.memory),
		fmt.Sprintf("%02d", a.threads),
		fmt.Sprintf("%02d", a.keyLen),
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	}, passwordDelim)

	return []byte(result), nil
}

func (a *argon) compare(password string, hash []byte) error {
	parts := strings.Split(string(hash), passwordDelim)

	if len(parts) != 7 {
		return errors.Errorf("%d is an invalid password delimited length for argon2", len(parts))
	}

	version, err := strconv.Atoi(parts[0])
	if err != nil {
		return errors.Wrap(err, "Getting version from password")
	}

	if version != argon2.Version {
		return errors.New("Invalid argon2 version")
	}

	time, err := strconv.Atoi(parts[1])
	if err != nil {
		return errors.Wrap(err, "Getting time factor from password")
	}
	memory, err := strconv.Atoi(parts[2])
	if err != nil {
		return errors.Wrap(err, "Getting memory factor from password")
	}

	threads, err := strconv.Atoi(parts[3])
	if err != nil {
		return errors.Wrap(err, "Getting threads factor from password")
	}
	keyLen, err := strconv.Atoi(parts[4])
	if err != nil {
		return errors.Wrap(err, "Getting keyLen factor from password")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return errors.Wrap(err, "Decoding salt from DB")
	}
	ar := &argon{
		time:    uint32(time),
		memory:  uint32(memory),
		threads: uint8(threads),
		keyLen:  uint32(keyLen),
	}

	other, err := ar.hashWithSalt(password, salt)
	if err != nil {
		return errors.Wrap(err, "Hashing password for comparison")
	}

	if subtle.ConstantTimeCompare(hash, other) == 1 {
		return nil
	}

	return ErrPasswordMismatch
}
