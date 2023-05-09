package rule

import (
	"context"
	"testing"

	"github.com/neblic/platform/controlplane/protos"
	"github.com/neblic/platform/sampler/defs"
	"github.com/neblic/platform/sampler/internal/sample"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestEvalJSON(t *testing.T) {
	for _, tc := range []struct {
		name      string
		filter    string
		sample    string
		wantMatch bool
	}{{
		name:      "simple match",
		filter:    `sample.sub_struct.id == 11`,
		sample:    `{"id": 1, "sub_struct": {"id": 11 }}`,
		wantMatch: true,
	}, {
		name:      "simple mismatch",
		filter:    `sample.id == 2`,
		sample:    `{"id": 1}`,
		wantMatch: false,
	}} {
		t.Run(tc.name, func(t *testing.T) {
			rb, err := NewBuilder(defs.DynamicSchema{})
			require.NoError(t, err)

			rule, err := rb.Build(tc.filter)
			require.NoError(t, err)

			s := sample.NewSampleDataFromJSON(tc.sample)

			gotMatch, err := rule.Eval(context.Background(), s)
			require.NoError(t, err)

			if gotMatch != tc.wantMatch {
				t.Errorf("expected cel(%q, %s) to be %v", tc.filter, tc.sample, tc.wantMatch)
			}
		})
	}
}

type sampleSubStruct struct {
	ID int
}

type sampleStruct struct {
	ID        int
	SubStruct sampleSubStruct
}

func TestEvalNative(t *testing.T) {
	for _, tc := range []struct {
		name      string
		filter    string
		sample    sampleStruct
		wantMatch bool
	}{{
		name:      "simple match",
		filter:    `sample.SubStruct.ID == 11`,
		sample:    sampleStruct{ID: 1, SubStruct: sampleSubStruct{ID: 11}},
		wantMatch: true,
	}, {
		name:      "simple mismatch",
		filter:    `sample.ID == 2`,
		sample:    sampleStruct{ID: 1},
		wantMatch: false,
	}} {
		t.Run(tc.name, func(t *testing.T) {
			rb, err := NewBuilder(defs.NewDynamicSchema())
			require.NoError(t, err)

			rule, err := rb.Build(tc.filter)
			require.NoError(t, err)

			s := sample.NewSampleDataFromNative(tc.sample)

			gotMatch, err := rule.Eval(context.Background(), s)
			require.NoError(t, err)

			if gotMatch != tc.wantMatch {
				t.Errorf("expected cel(%q, %+v) to be %v", tc.filter, tc.sample, tc.wantMatch)
			}
		})
	}
}

func TestEvalProto(t *testing.T) {
	for _, tc := range []struct {
		name      string
		filter    string
		sample    proto.Message
		wantMatch bool
	}{{
		name:   "simple match",
		filter: `sample.register_req.sampler_name == "sampler_name_value"`,
		sample: &protos.SamplerToServer{
			SamplerUid: "sampler_uid_value",
			Message: &protos.SamplerToServer_RegisterReq{
				RegisterReq: &protos.SamplerRegisterReq{SamplerName: "sampler_name_value"},
			}},
		wantMatch: true,
	}, {
		name:      "simple mismatch",
		filter:    `sample.sampler_uid == "non_matching_value"`,
		sample:    &protos.SamplerToServer{SamplerUid: "sampler_uid_value"},
		wantMatch: false,
	}} {
		t.Run(tc.name, func(t *testing.T) {
			rb, err := NewBuilder(defs.NewProtoSchema(&protos.SamplerToServer{}))
			require.NoError(t, err)

			rule, err := rb.Build(tc.filter)
			require.NoError(t, err)

			s := sample.NewSampleDataFromProto(tc.sample)

			gotMatch, err := rule.Eval(context.Background(), s)
			require.NoError(t, err)

			if gotMatch != tc.wantMatch {
				t.Errorf("expected cel(%q, %+v) to be %v", tc.filter, tc.sample, tc.wantMatch)
			}
		})
	}
}
