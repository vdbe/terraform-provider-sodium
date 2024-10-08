// Copyright (c) vdbe
// SPDX-License-Identifier: MPL-2.0

package encryption

import (
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/nacl/box"
)

type PubKey struct {
	Key string
}

type encryptResult struct {
	Err     error
	Encoded string
	Raw     []byte
}

func Encrypt(pk_b64 *string, secret []byte) (res encryptResult) {
	res = encryptResult{}

	pk, err := base64.StdEncoding.DecodeString(*pk_b64)
	if err != nil {
		res.Err = fmt.Errorf("failed to decode public key: %w", err)
		return
	}

	if len(pk) < 32 {
		res.Err = fmt.Errorf("invalid public key: length under 32")
		return
	}

	var peersPubKey [32]byte
	copy(peersPubKey[:], pk[0:32])

	var rand io.Reader
	eBody, err := box.SealAnonymous(nil, secret[:], &peersPubKey, rand)
	if err != nil {
		res.Err = fmt.Errorf("failed to encrypt body: %w", err)
		return
	}

	encoded := base64.StdEncoding.EncodeToString(eBody)

	res.Raw = eBody
	res.Encoded = encoded

	return res
}

// TODO: testing
