/*
 * Flow CLI
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tests

import (
	"fmt"

	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/spf13/afero"
)

type resource struct {
	Name     string
	Filename string
	Source   []byte
}

var ContractHelloString = resource{
	Name:     "Hello",
	Filename: "contractHello.cdc",
	Source: []byte(`
		pub contract Hello {
			pub let greeting: String
			init() {
				self.greeting = "Hello, World!"
			}
			pub fun hello(): String {
				return self.greeting
			}
		}
	`),
}

var ContractSimple = resource{
	Name:     "Simple",
	Filename: "contractSimple.cdc",
	Source: []byte(`
		pub contract Simple {}
	`),
}

var ContractSimpleUpdated = resource{
	Name:     "Simple",
	Filename: "contractSimpleUpdated.cdc",
	Source: []byte(`
		pub contract Simple {
			pub let greeting: String
			init() {
				self.greeting = "Foo"
			}
		}
	`),
}

var ContractEvents = resource{
	Name:     "ContractEvents",
	Filename: "contractEvents.cdc",
	Source: []byte(`
		pub contract ContractEvents {
			pub struct S {
				pub var x: Int
				pub var y: String
				
				init(x: Int, y: String) {
					self.x = x
					self.y = y
				}
			}

			pub event EventA(x: Int)
			pub event EventB(x: Int, y: Int)
			pub event EventC(x: UInt8)
			pub event EventD(x: String)
			pub event EventE(x: UFix64) 
			pub event EventF(x: Address)
			pub event EventG(x: [UInt8])
			pub event EventH(x: [[UInt8]])
			pub event EventI(x: {String: Int})
			pub event EventJ(x: S)
			
			init() {
				emit EventA(x: 1)				
				emit EventB(x: 4, y: 2)	
				emit EventC(x: 1)
				emit EventD(x: "hello")
				emit EventE(x: 10.2)
				emit EventF(x: 0x436164656E636521)
				emit EventG(x: "hello".utf8)
				emit EventH(x: ["hello".utf8, "world".utf8])
				emit EventI(x: { "hello": 1337 })
				emit EventJ(x: S(x: 1, y: "hello world"))
			}
		}
	`),
}

var TransactionArgString = resource{
	Filename: "transactionArg.cdc",
	Source: []byte(`
		transaction(greeting: String) {
			let guest: Address
			
			prepare(authorizer: AuthAccount) {
				self.guest = authorizer.address
			}
			
			execute {
				log(greeting.concat(",").concat(self.guest.toString()))
			}
		}
	`),
}

var TransactionSimple = resource{
	Filename: "transactionSimple.cdc",
	Source: []byte(`
		transaction() {}
	`),
}

var ScriptArgString = resource{
	Filename: "scriptArg.cdc",
	Source: []byte(`
		pub fun main(name: String): String {
		  return "Hello ".concat(name)
		}
	`),
}

var resources = []resource{
	ContractHelloString,
	TransactionArgString,
	ScriptArgString,
	ContractSimple,
	ContractSimpleUpdated,
	TransactionSimple,
}

func ReaderWriter() afero.Afero {
	var mockFS = afero.NewMemMapFs()

	for _, c := range resources {
		_ = afero.WriteFile(mockFS, c.Filename, c.Source, 0644)
	}

	return afero.Afero{mockFS}
}

func Alice() *flowkit.Account {
	a := &flowkit.Account{}
	a.SetAddress(flow.HexToAddress("0x1"))
	a.SetName("Alice")
	pk, _ := crypto.GeneratePrivateKey(crypto.ECDSA_P256, []byte("seedseedseedseedseedseedseedseedseedseedseedseed"))

	a.SetKey(flowkit.NewHexAccountKeyFromPrivateKey(0, crypto.SHA3_256, pk))

	return a
}

func PubKeys() []crypto.PublicKey {
	var pubKeys []crypto.PublicKey
	for x := 0; x < 5; x++ {
		pk, _ := crypto.GeneratePrivateKey(
			crypto.ECDSA_P256,
			[]byte(fmt.Sprintf("seedseedseedseedseedseedseedseedseedseedseedseed%d", x)),
		)
		pubKeys = append(pubKeys, pk.PublicKey())
	}
	return pubKeys
}
