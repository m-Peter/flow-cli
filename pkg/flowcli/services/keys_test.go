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

package services

import (
	"testing"

	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/onflow/flow-cli/pkg/flowcli/project"
	"github.com/onflow/flow-cli/tests"
)

func TestKeys(t *testing.T) {
	mock := &tests.MockGateway{}

	proj, err := project.Init(crypto.ECDSA_P256, crypto.SHA3_256)
	assert.NoError(t, err)

	keys := NewKeys(mock, proj, output.NewStdoutLogger(output.InfoLog))

	t.Run("Generate Keys", func(t *testing.T) {
		key, err := keys.Generate("", "ECDSA_P256")

		assert.NoError(t, err)
		assert.Equal(t, len(key.PrivateKey.String()), 66)
	})

	t.Run("Generate Keys with seed", func(t *testing.T) {
		key, err := keys.Generate("aaaaaaaaaaaaaaaaaaaaaaannndddddd_its_gone", "ECDSA_P256")

		assert.NoError(t, err)
		assert.Equal(t, key.PrivateKey.String(), "0x134f702d0872dba9c7aea15498aab9b2ffedd5aeebfd8ac3cf47c591f0d7ce52")
	})

	t.Run("Fail generate keys, too short seed", func(t *testing.T) {
		_, err := keys.Generate("im_short", "ECDSA_P256")

		assert.Equal(t, err.Error(), "failed to generate private key: crypto: insufficient seed length 8, must be at least 32 bytes for ECDSA_P256")
	})

	t.Run("Fail generate keys, invalid sig algo", func(t *testing.T) {
		_, err := keys.Generate("", "JUSTNO")

		assert.Equal(t, err.Error(), "invalid signature algorithm: JUSTNO")
	})

}