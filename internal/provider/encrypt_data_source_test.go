// Copyright (c) vdbe
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"golang.org/x/crypto/nacl/box"
)

func generateKey() (public_key_base64 string, private_key_base64 string, err error) {
	c := 32
	b := make([]byte, c)
	_, err = rand.Read(b)
	if err != nil {
		return
	}

	rand := bytes.NewReader(b)

	pub_k, priv_k, err := box.GenerateKey(rand)
	if err != nil {
		return
	}

	public_key_base64 = base64.StdEncoding.EncodeToString(pub_k[:])
	private_key_base64 = base64.StdEncoding.EncodeToString(priv_k[:])

	return
}

func TestAccEncryptDataSource(t *testing.T) {
	public_key_base64, _, err := generateKey()
	if err != nil {
		t.Fatal(err)
	}

	value := "value"
	value_base64 := base64.StdEncoding.EncodeToString([]byte(value))

	// encrypted := encryption.Encrypt(&encryption.PubKey{Key: public_key}, []byte(value))
	// encrypted_base64 := encryption.Encrypt(&encryption.PubKey{Key: public_key}, []byte(value_base64))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Check defaults
			{
				ResourceName: "sodium_encrypt.test",
				Config: fmt.Sprintf(`
            data "sodium_encrypt" "test" {
              public_key_base64 = "%s"
              value = "%s"
            }
          `, public_key_base64, value),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.sodium_encrypt.test", "base64_encode", "false"),
				),
			},
			// Check b64 conversions
			{
				ResourceName: "sodium_encrypt.test",
				Config: fmt.Sprintf(`
            data "sodium_encrypt" "test" {
              public_key_base64 = "%s"
              value = "%s"
            }
          `, public_key_base64, value),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.sodium_encrypt.test", "value_base64", value_base64),
				),
			},
			{
				ResourceName: "sodium_encrypt.test",
				Config: fmt.Sprintf(`
            data "sodium_encrypt" "test" {
              public_key_base64 = "%s"
              value_base64 = "%s"
            }
          `, public_key_base64, value_base64),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.sodium_encrypt.test", "value", value),
				),
			},
			// TODO: Check encryption
			// base64_encode: true/false
			// decode with public key
		},
	})
}
