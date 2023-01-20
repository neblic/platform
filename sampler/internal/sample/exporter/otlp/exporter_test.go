package otlp

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/neblic/platform/logging"
	"github.com/neblic/platform/sampler/internal/sample/exporter/otlp/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
)

func TestExportLogs(t *testing.T) {
	// Start an OTLP-compatible receiver.
	ln, err := net.Listen("tcp", "localhost:")
	require.NoError(t, err, "Failed to find an available address to run the gRPC server: %v", err)
	rcv := mock.OtlpLogsReceiverOnGRPCServer(ln)
	defer rcv.Srv.GracefulStop()

	// Initialize OTLP exporter
	logger, err := logging.NewZapDev()
	require.NoError(t, err, "Failed to initialize logger: %v", err)
	exporter, err := New(context.Background(), logger, ln.Addr().String(), newDefaultOptions())
	require.NoError(t, err, "Failed to initialize exporter: %v", err)

	// Ensure that initially there is no data in the receiver.
	assert.EqualValues(t, 0, rcv.RequestCount.Load())

	// Generate and send logs
	logs := plog.NewLogs()
	rlogs := logs.ResourceLogs()
	rlogs.EnsureCapacity(4)
	for i := 0; i < 4; i++ {
		rl := rlogs.AppendEmpty()
		sl := rl.ScopeLogs().AppendEmpty()
		lr := sl.LogRecords().AppendEmpty()
		lr.SetTimestamp(pcommon.Timestamp(time.Now().UnixMilli() + int64(i)))
		lr.SetSeverityNumber(plog.SeverityNumberInfo)
		lr.Body().SetStr(fmt.Sprintf("log %d", i))
	}

	err = exporter.exportLogs(context.Background(), logs)
	require.NoError(t, err)

	// Wait until it is received.
	assert.Eventually(t, func() bool {
		return rcv.RequestCount.Load() > 0
	}, 10*time.Second, 5*time.Millisecond)

	// Verify received logs.
	assert.EqualValues(t, 1, rcv.RequestCount.Load())
	assert.EqualValues(t, 4, rcv.TotalItems.Load())
	assert.EqualValues(t, logs, rcv.GetLastRequest())

	err = exporter.Close(context.Background())
	require.NoError(t, err, "Failed to stop exporter: %v", err)
}
