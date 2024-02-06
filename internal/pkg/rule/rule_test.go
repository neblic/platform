package rule

import (
	"context"
	"testing"

	"github.com/neblic/platform/controlplane/protos"
	"github.com/neblic/platform/internal/pkg/data"
	"github.com/neblic/platform/sampler/sample"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestEvalJSON(t *testing.T) {
	for _, tc := range []struct {
		name       string
		expression string
		sample     string
		wantMatch  bool
	}{
		{
			name:       "simple match",
			expression: `sample.sub_struct.id == 11`,
			sample:     `{"id": 1, "sub_struct": {"id": 11 }}`,
			wantMatch:  true,
		},
		{
			name:       "simple mismatch",
			expression: `sample.id == 2`,
			sample:     `{"id": 1}`,
			wantMatch:  false,
		},
		{
			name:       "sequence check",
			expression: `sequence(sample.id, "asc")`,
			sample:     `{"id": 1}`,
			wantMatch:  true,
		},
		{
			name:       "complete check",
			expression: `complete(sample.id, 1.0)`,
			sample:     `{"id": 1}`,
			wantMatch:  true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rb, err := NewBuilder(sample.DynamicSchema{}, CheckFunctions)
			require.NoError(t, err)

			rule, err := rb.Build(tc.expression)
			require.NoError(t, err)

			s := data.NewSampleDataFromJSON(tc.sample)

			gotMatch, err := rule.Eval(context.Background(), s)
			require.NoError(t, err)

			if gotMatch != tc.wantMatch {
				t.Errorf("expected cel(%q, %s) to be %v", tc.expression, tc.sample, tc.wantMatch)
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
		name       string
		expression string
		sample     sampleStruct
		wantMatch  bool
	}{
		{
			name:       "simple match",
			expression: `sample.SubStruct.ID == 11`,
			sample:     sampleStruct{ID: 1, SubStruct: sampleSubStruct{ID: 11}},
			wantMatch:  true,
		}, {
			name:       "simple mismatch",
			expression: `sample.ID == 2`,
			sample:     sampleStruct{ID: 1},
			wantMatch:  false,
		},
		{
			name:       "sequence check",
			expression: `sequence(sample.ID, "asc")`,
			sample:     sampleStruct{ID: 1},
			wantMatch:  true,
		},
		{
			name:       "complete check",
			expression: `complete(sample.ID, 1)`,
			sample:     sampleStruct{ID: 1},
			wantMatch:  true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rb, err := NewBuilder(sample.NewDynamicSchema(), CheckFunctions)
			require.NoError(t, err)

			rule, err := rb.Build(tc.expression)
			require.NoError(t, err)

			s := data.NewSampleDataFromNative(tc.sample)

			gotMatch, err := rule.Eval(context.Background(), s)
			require.NoError(t, err)

			if gotMatch != tc.wantMatch {
				t.Errorf("expected cel(%q, %+v) to be %v", tc.expression, tc.sample, tc.wantMatch)
			}
		})
	}
}

func TestEvalProto(t *testing.T) {
	for _, tc := range []struct {
		name       string
		expression string
		sample     proto.Message
		wantMatch  bool
	}{
		{
			name:       "simple match",
			expression: `sample.name == "sampler_name_value"`,
			sample: &protos.SamplerToServer{
				Name:       "sampler_name_value",
				SamplerUid: "sampler_uid_value",
				Message: &protos.SamplerToServer_RegisterReq{
					RegisterReq: &protos.SamplerRegisterReq{},
				}},
			wantMatch: true,
		},
		{
			name:       "simple mismatch",
			expression: `sample.sampler_uid == "non_matching_value"`,
			sample:     &protos.SamplerToServer{SamplerUid: "sampler_uid_value"},
			wantMatch:  false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rb, err := NewBuilder(sample.NewProtoSchema(&protos.SamplerToServer{}), CheckFunctions)
			require.NoError(t, err)

			rule, err := rb.Build(tc.expression)
			require.NoError(t, err)

			s := data.NewSampleDataFromProto(tc.sample)

			gotMatch, err := rule.Eval(context.Background(), s)
			require.NoError(t, err)

			if gotMatch != tc.wantMatch {
				t.Errorf("expected cel(%q, %+v) to be %v", tc.expression, tc.sample, tc.wantMatch)
			}
		})
	}
}

func TestEvalSequence(t *testing.T) {
	rb, err := NewBuilder(defs.DynamicSchema{}, CheckFunctions)
	require.NoError(t, err)

	rule, err := rb.Build(`sequence(sample.id, "asc")`)
	require.NoError(t, err)

	gotMatch, err := rule.Eval(context.Background(), data.NewSampleDataFromJSON(`{"id": 1}`))
	require.NoError(t, err)
	require.True(t, gotMatch)

	gotMatch, err = rule.Eval(context.Background(), data.NewSampleDataFromJSON(`{"id": 2}`))
	require.NoError(t, err)
	require.True(t, gotMatch)

	gotMatch, err = rule.Eval(context.Background(), data.NewSampleDataFromJSON(`{"id": -1}`))
	require.NoError(t, err)
	require.False(t, gotMatch)
}

func TestEvalComplete(t *testing.T) {
	rb, err := NewBuilder(defs.DynamicSchema{}, CheckFunctions)
	require.NoError(t, err)

	rule, err := rb.Build(`complete(sample.id, 1.0)`)
	require.NoError(t, err)

	gotMatch, err := rule.Eval(context.Background(), data.NewSampleDataFromJSON(`{"id": 1}`))
	require.NoError(t, err)
	require.True(t, gotMatch)

	gotMatch, err = rule.Eval(context.Background(), data.NewSampleDataFromJSON(`{"id": 2}`))
	require.NoError(t, err)
	require.True(t, gotMatch)

	gotMatch, err = rule.Eval(context.Background(), data.NewSampleDataFromJSON(`{"id": 1000}`))
	require.NoError(t, err)
	require.False(t, gotMatch)
}
