package fileserver

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"strconv"
)

type ETagFunc func(r io.Reader) (string, error)

// Calculates the entity tag by md5 hashing r and quoting the hex encoded result.
func CalculateETag(r io.Reader) (string, error) {
	hasher := md5.New()
	_, err := io.Copy(hasher, r)
	if err != nil {
		return "", err
	}
	hash := hasher.Sum(nil)
	value := hex.EncodeToString(hash)
	return strconv.Quote(value), nil
}
