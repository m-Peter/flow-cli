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
	"fmt"
	"strings"

	"github.com/onflow/flow-cli/pkg/flow/keys"

	"github.com/onflow/flow-cli/pkg/flow/config"

	"github.com/onflow/flow-cli/pkg/flow"

	"github.com/onflow/flow-cli/pkg/flow/gateway"
	"github.com/onflow/flow-cli/pkg/flow/util"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/templates"
)

// Accounts service handles all interactions for accounts
type Accounts struct {
	gateway gateway.Gateway
	project *flow.Project
	logger  util.Logger
}

// NewAccounts create new account service
func NewAccounts(
	gateway gateway.Gateway,
	project *flow.Project,
	logger util.Logger,
) *Accounts {
	return &Accounts{
		gateway: gateway,
		project: project,
		logger:  logger,
	}
}

// Get gets an account based on address
func (a *Accounts) Get(address string) (*flowsdk.Account, error) {
	flowAddress := flowsdk.HexToAddress(
		strings.ReplaceAll(address, "0x", ""),
	)

	return a.gateway.GetAccount(flowAddress)
}

func (a *Accounts) Add(
	name string,
	address string,
	signatureAlgorithm string,
	hashingAlgorithm string,
	keyIndex int,
	keyHex string,
	keyContext string,
	overwrite bool,
) (*flow.Account, error) {

	existingAccount := a.project.GetAccountByName(name)
	if existingAccount != nil && !overwrite {
		return nil, fmt.Errorf("account with name [%s] already exists in the config, use --overwrite flag if you want to overwrite it", name)
	}

	sigAlgo, hashAlgo, err := util.ConvertSigAndHashAlgo(signatureAlgorithm, hashingAlgorithm)
	if err != nil {
		return nil, err
	}

	if keyIndex < 0 {
		return nil, fmt.Errorf("key index must be positive number")
	}

	confAccount := config.Account{
		Name:    name,
		Address: flowsdk.HexToAddress(address),
		ChainID: "", // todo: chain id
		Keys:    nil,
	}

	accountKey := config.AccountKey{
		Type:     "",
		Index:    keyIndex,
		SigAlgo:  sigAlgo,
		HashAlgo: hashAlgo,
		Context:  nil,
	}

	// hex key
	if keyHex != "" {
		_, err := crypto.DecodePrivateKeyHex(sigAlgo, keyHex)
		if err != nil {
			return nil, fmt.Errorf("key hex could not be parsed")
		}

		accountKey.Type = config.KeyTypeHex
		accountKey.Context[config.PrivateKeyField] = keyHex

	} else if keyContext != "" {
		keyCtx, err := keys.KeyContextFromKMSResourceID(keyContext)
		if err != nil {
			return nil, fmt.Errorf("key context could not be parsed %s", keyContext)
		}

		accountKey.Type = config.KeyTypeGoogleKMS
		accountKey.Context = keyCtx

	} else {
		return nil, fmt.Errorf("either --privatekey or --context flag must be provided")
	}

	confAccount.Keys = []config.AccountKey{accountKey}

	account, err := flow.AccountFromConfig(confAccount)
	if err != nil {
		return nil, err
	}

	_, err = account.DefaultKey().Signer().Sign([]byte("test"))
	if err != nil {
		return nil, fmt.Errorf("could not sign with the new key")
	}

	a.project.AddAccount(account)

	return account, nil
}

// Create creates an account with signer name, keys, algorithms, contracts and returns the new account
func (a *Accounts) Create(
	signerName string,
	keys []string,
	signatureAlgorithm string,
	hashingAlgorithm string,
	contracts []string,
) (*flowsdk.Account, error) {

	signer := a.project.GetAccountByName(signerName)
	if signer == nil {
		return nil, fmt.Errorf("signer account: [%s] doesn't exists in configuration", signerName)
	}

	accountKeys := make([]*flowsdk.AccountKey, len(keys))

	sigAlgo, hashAlgo, err := util.ConvertSigAndHashAlgo(signatureAlgorithm, hashingAlgorithm)
	if err != nil {
		return nil, err
	}

	for i, publicKeyHex := range keys {
		publicKey, err := crypto.DecodePublicKeyHex(
			sigAlgo,
			strings.ReplaceAll(publicKeyHex, "0x", ""),
		)
		if err != nil {
			return nil, fmt.Errorf("could not decode public key for key: %s, with signature algorith: %s", publicKeyHex, sigAlgo)
		}

		accountKeys[i] = &flowsdk.AccountKey{
			PublicKey: publicKey,
			SigAlgo:   sigAlgo,
			HashAlgo:  hashAlgo,
			Weight:    flowsdk.AccountKeyWeightThreshold,
		}
	}

	var contractTemplates []templates.Contract

	for _, contract := range contracts {
		contractFlagContent := strings.SplitN(contract, ":", 2)
		if len(contractFlagContent) != 2 {
			return nil, fmt.Errorf("wrong format for contract flag. Correct format is name:path, but got: %s", contract)
		}
		contractName := contractFlagContent[0]
		contractPath := contractFlagContent[1]

		contractSource, err := util.LoadFile(contractPath)
		if err != nil {
			return nil, err
		}

		contractTemplates = append(contractTemplates,
			templates.Contract{
				Name:   contractName,
				Source: string(contractSource),
			},
		)
	}

	tx := templates.CreateAccount(accountKeys, contractTemplates, signer.Address())
	tx, err = a.gateway.SendTransaction(tx, signer)
	if err != nil {
		return nil, err
	}

	a.logger.StartProgress("Waiting for transaction to be sealed...")

	result, err := a.gateway.GetTransactionResult(tx, true)
	if err != nil {
		return nil, err
	}

	a.logger.StopProgress("")

	events := flow.EventsFromTransaction(result)
	newAccountAddress := events.GetAddress()

	if newAccountAddress == nil {
		return nil, fmt.Errorf("new account address couldn't be fetched")
	}

	return a.gateway.GetAccount(*newAccountAddress)
}

// AddContract adds new contract to the account and returns the updated account
func (a *Accounts) AddContract(
	accountName string,
	contractName string,
	contractFilename string,
	updateExisting bool,
) (*flowsdk.Account, error) {

	account := a.project.GetAccountByName(accountName)
	if account == nil {
		return nil, fmt.Errorf("account: [%s] doesn't exists in configuration", accountName)
	}

	return a.addContract(account, contractName, contractFilename, updateExisting)
}

// AddContractForAddress adds new contract to the address using private key specified
func (a *Accounts) AddContractForAddress(
	accountAddress string,
	accountPrivateKey string,
	contractName string,
	contractFilename string,
	updateExisting bool,
) (*flowsdk.Account, error) {
	account, err := util.AccountFromAddressAndKey(accountAddress, accountPrivateKey)
	if err != nil {
		return nil, err
	}

	return a.addContract(account, contractName, contractFilename, updateExisting)
}

func (a *Accounts) addContract(
	account *flow.Account,
	contractName string,
	contractFilename string,
	updateExisting bool,
) (*flowsdk.Account, error) {
	contractSource, err := util.LoadFile(contractFilename)
	if err != nil {
		return nil, err
	}

	tx := templates.AddAccountContract(
		account.Address(),
		templates.Contract{
			Name:   contractName,
			Source: string(contractSource),
		},
	)

	// if we are updating contract
	if updateExisting {
		tx = templates.UpdateAccountContract(
			account.Address(),
			templates.Contract{
				Name:   contractName,
				Source: string(contractSource),
			},
		)
	}

	// send transaction with contract
	tx, err = a.gateway.SendTransaction(tx, account)
	if err != nil {
		return nil, err
	}

	// we wait for transaction to be sealed
	_, err = a.gateway.GetTransactionResult(tx, true)
	if err != nil {
		return nil, err
	}

	return a.gateway.GetAccount(account.Address())
}

// RemoveContracts removes a contract from the account
func (a *Accounts) RemoveContract(
	contractName string,
	accountName string,
) (*flowsdk.Account, error) {
	account := a.project.GetAccountByName(accountName)
	if account == nil {
		return nil, fmt.Errorf("account: [%s] doesn't exists in configuration", accountName)
	}

	return a.removeContract(contractName, account)
}

// RemoveContractForAddress removes contract from address using private key
func (a *Accounts) RemoveContractForAddress(
	contractName string,
	accountAddress string,
	accountPrivateKey string,
) (*flowsdk.Account, error) {
	account, err := util.AccountFromAddressAndKey(accountAddress, accountPrivateKey)
	if err != nil {
		return nil, err
	}

	return a.removeContract(contractName, account)
}

func (a *Accounts) removeContract(
	contractName string,
	account *flow.Account,
) (*flowsdk.Account, error) {
	tx := templates.RemoveAccountContract(account.Address(), contractName)
	tx, err := a.gateway.SendTransaction(tx, account)
	if err != nil {
		return nil, err
	}

	return a.gateway.GetAccount(account.Address())
}
