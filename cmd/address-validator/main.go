package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/mr-tron/base58"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ripemd160"
	"hash"
)

func main() {
	encoded := "XxCZ4fKJUa8sQkPF75EYuUEXTUGicf8PrfrEXXwknTFsLnz1T236pHvugHeDB"

	validateAddress(encoded)
}

func validateAddress(address string) error {
	b58, err := base58.Decode(address)
	if err != nil {
		return err
	}

	if len(b58) < 4 {
		return errors.New("Address is to short")
	}

	csum := b58[len(b58)-4:]
	log.Info("csum ", hex.EncodeToString(csum))

	data := b58[0 : len(b58)-4]
	log.WithField("len", len(data)).Info("data ", hex.EncodeToString(data))

	h := s256(s256(data))

	if hex.EncodeToString(h[0:4]) != hex.EncodeToString(csum) {
		return errors.New("Invalid checksum")
	}

	log.Info(hex.EncodeToString(data[0:20]))
	log.Info(hex.EncodeToString(data[20:40]))
	log.Info(hex.EncodeToString(data[40:41]))
	if len(data) == 41 {
		f := data[0:20]
		sh := string(Hash160(f))
		//b58 := base58.Encode(sh)
		log.Info("Spending address is ", sh)
	}

	return nil
}

func s256(i []byte) []byte {
	s := sha256.New()
	s.Write(i)

	return s.Sum(nil)
}

// Calculate the hash of hasher over buf.
func calcHash(buf []byte, hasher hash.Hash) []byte {
	hasher.Write(buf)
	return hasher.Sum(nil)
}

// Hash160 calculates the hash ripemd160(sha256(b)).
func Hash160(buf []byte) []byte {
	return calcHash(calcHash(buf, sha256.New()), ripemd160.New())
}
