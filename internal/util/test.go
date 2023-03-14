package util

import (
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/mocks"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/tests"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/require"
	"testing"
)

var NoFlags = command.GlobalFlags{}
var NoLogger = output.NewStdoutLogger(output.NoneLog)

// TestMocks creates mock flowkit services, an empty state and a mock reader writer
func TestMocks(t *testing.T) (*mocks.MockServices, *flowkit.State, flowkit.ReaderWriter) {
	services := mocks.DefaultMockServices()
	rw, _ := tests.ReaderWriter()
	state, err := flowkit.Init(rw, crypto.ECDSA_P256, crypto.SHA3_256)
	require.NoError(t, err)

	return services, state, rw
}
