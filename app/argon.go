// Copyright (c) 2017 Townsourced Inc.

package app

import (
	"bytes"
	"crypto/rand"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"io"
	"strconv"

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

	unencodedSalt := make([]byte, maxSaltSize)
	_, err := io.ReadFull(rand.Reader, unencodedSalt)
	if err != nil {
		return nil, err
	}

	salt := base64.RawStdEncoding.EncodeToString(unencodedSalt)
	return a.hashWithSalt(password, []byte(salt))
}

func (a *argon) hashWithSalt(password string, salt []byte) ([]byte, error) {
	sha := sha512.Sum512_256([]byte(password))

	hash := argon2.Key(sha[:], salt, a.time, a.memory, a.threads, a.keyLen)

	// version$time$memory$threads$keyLen$salt$hash
	result := bytes.Join([][]byte{
		[]byte(fmt.Sprintf("%02d", argon2.Version)),
		[]byte(fmt.Sprintf("%02d", a.time)),
		[]byte(fmt.Sprintf("%02d", a.memory)),
		[]byte(fmt.Sprintf("%02d", a.threads)),
		[]byte(fmt.Sprintf("%02d", a.keyLen)),
		salt,
		hash,
	}, []byte(passwordDelim))

	return result, nil

}

func (a *argon) compare(password string, hash []byte) error {
	parts := bytes.Split(hash, []byte(passwordDelim))

	if len(parts) != 7 {
		return errors.New("Invalid password delimited length for argon2")
	}

	version, err := strconv.Atoi(string(parts[0]))
	if err != nil {
		return errors.Wrap(err, "Getting version from password")
	}

	if version != argon2.Version {
		return errors.New("Invalid argon2 version")
	}

	time, err := strconv.Atoi(string(parts[1]))
	if err != nil {
		return errors.Wrap(err, "Getting time factor from password")
	}
	memory, err := strconv.Atoi(string(parts[2]))
	if err != nil {
		return errors.Wrap(err, "Getting memory factor from password")
	}

	threads, err := strconv.Atoi(string(parts[3]))
	if err != nil {
		return errors.Wrap(err, "Getting threads factor from password")
	}
	keyLen, err := strconv.Atoi(string(parts[4]))
	if err != nil {
		return errors.Wrap(err, "Getting keyLen factor from password")
	}

	salt := parts[5]
	arHash := parts[6]

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

	if subtle.ConstantTimeCompare(arHash, other) == 1 {
		return nil
	}

	return ErrLogonFailure
}
