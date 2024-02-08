package grpc

import (
	"testing"

	"github.com/neblic/platform/controlplane/protos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetValueFromProto(t *testing.T) {
	// any message will do
	msg := protos.ClientToServer{
		ClientUid: "uid",
		Message: &protos.ClientToServer_SamplerConfReq{
			SamplerConfReq: &protos.ClientSamplerConfReq{
				SamplerName: "sampler",
			},
		},
	}

	// get a value from a top-level field
	value := getValueFromProto(&msg, "client_uid")
	str, ok := value.Interface().(string)
	require.True(t, ok)
	assert.Equal(t, "uid", str)

	// get a value from a nested field
	value = getValueFromProto(&msg, "sampler_conf_req.sampler_name")
	str, ok = value.Interface().(string)
	require.True(t, ok)
	assert.Equal(t, "sampler", str)

	// get a value from a non-existent field
	value = getValueFromProto(&msg, "non_existent_field")
	str, ok = value.Interface().(string)
	require.False(t, ok)
}
