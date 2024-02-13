package rule

import (
	"context"
	"testing"

	"github.com/neblic/platform/controlplane/control"
	"github.com/neblic/platform/controlplane/protos"
	"github.com/neblic/platform/internal/pkg/data"
	"github.com/neblic/platform/internal/pkg/rule/function"
	"github.com/neblic/platform/sampler/sample"
	"github.com/stretchr/testify/assert"
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

			rule, err := rb.Build(tc.expression, control.Keyed{})
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

			rule, err := rb.Build(tc.expression, control.Keyed{})
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

			rule, err := rb.Build(tc.expression, control.Keyed{})
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
	rb, err := NewBuilder(sample.DynamicSchema{}, CheckFunctions)
	require.NoError(t, err)

	rule, err := rb.Build(`sequence(sample.id, "asc")`, control.Keyed{})
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
	rb, err := NewBuilder(sample.DynamicSchema{}, CheckFunctions)
	require.NoError(t, err)

	rule, err := rb.Build(`complete(sample.id, 1.0)`, control.Keyed{})
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

func TestEvalKeyedJSON(t *testing.T) {
	rb, err := NewBuilder(sample.DynamicSchema{}, CheckFunctions)
	require.NoError(t, err)

	rule, err := rb.Build(`sequence(sample.id, "asc")`, control.Keyed{Enabled: true, MaxKeys: 2})
	require.NoError(t, err)

	// key1 first eval is always true
	gotMatch, err := rule.EvalKeyed(context.Background(), "key1", data.NewSampleDataFromJSON(`{"id": 10}`))
	require.NoError(t, err)
	require.True(t, gotMatch)

	// key1 eval bigger number is true
	gotMatch, err = rule.EvalKeyed(context.Background(), "key1", data.NewSampleDataFromJSON(`{"id": 11}`))
	require.NoError(t, err)
	require.True(t, gotMatch)

	// key2 first eval is always true. A smaller number than key1 was set to state isolation
	gotMatch, err = rule.EvalKeyed(context.Background(), "key2", data.NewSampleDataFromJSON(`{"id": 0}`))
	require.NoError(t, err)
	require.True(t, gotMatch)

	// key2 eval bigger number is true
	gotMatch, err = rule.EvalKeyed(context.Background(), "key2", data.NewSampleDataFromJSON(`{"id": 1}`))
	require.NoError(t, err)
	require.True(t, gotMatch)

	// key1 eval smaller number is false
	gotMatch, err = rule.EvalKeyed(context.Background(), "key1", data.NewSampleDataFromJSON(`{"id": 9}`))
	require.NoError(t, err)
	require.False(t, gotMatch)

	// key2 eval smaller number is false
	gotMatch, err = rule.EvalKeyed(context.Background(), "key2", data.NewSampleDataFromJSON(`{"id": -1}`))
	require.NoError(t, err)
	require.False(t, gotMatch)

	// key3 first eval must return an error because the maximum number of keys was reached
	_, err = rule.EvalKeyed(context.Background(), "key3", data.NewSampleDataFromJSON(`{"id": 20}`))
	require.Error(t, function.ErrMaxKeys, err)

	// key2 eval bigger number must be true even after reaching the macimum number of keys
	gotMatch, err = rule.EvalKeyed(context.Background(), "key2", data.NewSampleDataFromJSON(`{"id": 2}`))
	require.NoError(t, err)
	require.True(t, gotMatch)
}

func TestEvaluateEvalTrue(t *testing.T) {
	tcs := []struct {
		name                string
		rule                string
		wantReturnStaticRes bool
		staticRes           bool
	}{
		{
			name:                "simplest true rule",
			rule:                `true`,
			wantReturnStaticRes: true,
			staticRes:           true,
		},
		{
			name:                "simplest false rule",
			rule:                `false`,
			wantReturnStaticRes: true,
			staticRes:           false,
		},
		{
			name:                "static true expression",
			rule:                `1 == 1`,
			wantReturnStaticRes: true,
			staticRes:           true,
		},
		{
			name:                "static false expression",
			rule:                `1 != 1`,
			wantReturnStaticRes: true,
			staticRes:           false,
		},
		{
			name:                "true expression with vars",
			rule:                `sample.id == 1 || 1 == 1`,
			wantReturnStaticRes: true,
			staticRes:           true,
		},
		{
			name:                "non static expression",
			rule:                `sample.id == 1 || 1 != 1`,
			wantReturnStaticRes: false,
		},
	}

	rb, err := NewBuilder(sample.NewDynamicSchema(), StreamFunctions)
	require.NoError(t, err)

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			rule, err := rb.Build(tc.rule, control.Keyed{})
			require.NoError(t, err)
			require.Equal(t, tc.wantReturnStaticRes, rule.returnsStaticRes)
			assert.Equal(t, tc.staticRes, rule.staticRes)
		})
	}
}
