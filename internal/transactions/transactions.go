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

package transactions

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	"github.com/onflow/flow-cli/internal/events"

	"github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:              "transactions",
	Short:            "Utilities to send transactions",
	TraverseChildren: true,
}

func init() {
	GetCommand.AddToParent(Cmd)
	SendCommand.AddToParent(Cmd)
	SignCommand.AddToParent(Cmd)
	BuildCommand.AddToParent(Cmd)
	SendSignedCommand.AddToParent(Cmd)
}

// TransactionResult represent result from all account commands
type TransactionResult struct {
	result *flow.TransactionResult
	tx     *flow.Transaction
	code   bool
}

// JSON convert result to JSON
func (r *TransactionResult) JSON() interface{} {
	result := make(map[string]string)
	result["id"] = r.tx.ID().String()
	result["payload"] = fmt.Sprintf("%x", r.tx.Encode())
	result["authorizers"] = fmt.Sprintf("%s", r.tx.Authorizers)
	result["payer"] = r.tx.Payer.String()

	if r.result != nil {
		result["events"] = fmt.Sprintf("%s", r.result.Events)
		result["status"] = r.result.Status.String()
		result["error"] = r.result.Error.Error()
	}

	return result
}

// String convert result to string
func (r *TransactionResult) String() string {
	var b bytes.Buffer
	writer := tabwriter.NewWriter(&b, 0, 8, 1, '\t', tabwriter.AlignRight)

	if r.result != nil {
		if r.result.Error != nil {
			fmt.Fprintf(writer, "❌ Transaction Error \n%s\n\n\n", r.result.Error.Error())
		}

		statusBadge := ""
		if r.result.Status == flow.TransactionStatusSealed {
			statusBadge = "✅"
		}
		fmt.Fprintf(writer, "Status\t%s %s\n", statusBadge, r.result.Status)
	}

	fmt.Fprintf(writer, "ID\t%s\n", r.tx.ID())
	fmt.Fprintf(writer, "Payer\t%s\n", r.tx.Payer.Hex())
	fmt.Fprintf(writer, "Authorizers\t%s\n", r.tx.Authorizers)

	fmt.Fprintf(writer,
		"\nProposal Key:\t\n    Address\t%s\n    Index\t%v\n    Sequence\t%v\n",
		r.tx.ProposalKey.Address, r.tx.ProposalKey.KeyIndex, r.tx.ProposalKey.SequenceNumber,
	)

	if len(r.tx.PayloadSignatures) == 0 {
		fmt.Fprintf(writer, "\nNo Payload Signatures\n")
	}

	if len(r.tx.EnvelopeSignatures) == 0 {
		fmt.Fprintf(writer, "\nNo Envelope Signatures\n")
	}

	for i, e := range r.tx.PayloadSignatures {
		fmt.Fprintf(writer, "\nPayload Signature %v:\n", i)
		fmt.Fprintf(writer, "    Address\t%s\n", e.Address)
		fmt.Fprintf(writer, "    Signature\t%x\n", e.Signature)
		fmt.Fprintf(writer, "    Key Index\t%d\n", e.KeyIndex)
	}

	for i, e := range r.tx.EnvelopeSignatures {
		fmt.Fprintf(writer, "\nEnvelope Signature %v:\n", i)
		fmt.Fprintf(writer, "    Address\t%s\n", e.Address)
		fmt.Fprintf(writer, "    Signature\t%x\n", e.Signature)
		fmt.Fprintf(writer, "    Key Index\t%d\n", e.KeyIndex)
	}

	if r.result != nil {
		e := events.EventResult{
			Events: r.result.Events,
		}

		eventsOutput := e.String()
		if eventsOutput == "" {
			eventsOutput = "None"
		}

		fmt.Fprintf(writer, "\n\nEvents:\t %s\n", eventsOutput)
	}

	if r.tx.Script != nil {
		if len(r.tx.Arguments) == 0 {
			fmt.Fprintf(writer, "\n\nArguments\tNo arguments\n")
		} else {
			fmt.Fprintf(writer, "\n\nArguments (%d):\n", len(r.tx.Arguments))
			for i, argument := range r.tx.Arguments {
				fmt.Fprintf(writer, "    - Argument %d: %s\n", i, argument)
			}
		}

		fmt.Fprintf(writer, "\nCode\n\n%s\n", r.tx.Script)
	}

	fmt.Fprintf(writer, "\n\nPayload:\n%x", r.tx.Encode())

	writer.Flush()
	return b.String()
}

// Oneliner show result as one liner grep friendly
func (r *TransactionResult) Oneliner() string {
	return fmt.Sprintf("ID: %s, Status: %s, Events: %s", r.tx.ID(), r.result.Status, r.result.Events)
}