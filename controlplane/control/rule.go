package control

import (
	"github.com/neblic/platform/controlplane/protos"
)

type SamplerStreamRuleUID string

type RuleLang int

const (
	SrlCel = iota + 1
)

func (srl RuleLang) String() string {
	switch srl {
	case SrlCel:
		return "CEL"
	default:
		return "Unknown"
	}
}

func NewRuleLangFromProto(lang protos.Rule_Language) RuleLang {
	var srl RuleLang
	switch lang {
	case protos.Rule_CEL:
		srl = SrlCel
	}

	return srl
}

func (srl RuleLang) ToProto() protos.Rule_Language {
	switch srl {
	case SrlCel:
		return protos.Rule_CEL
	default:
		return protos.Rule_UNKNOWN
	}
}

type Rule struct {
	Lang       RuleLang
	Expression string
}

func (r Rule) String() string {
	// Lang intentionally not shown since it is implicit it is always a CEL type for now
	return r.Expression
}

func NewRuleFromProto(sr *protos.Rule) Rule {
	if sr == nil {
		return Rule{}
	}

	return Rule{
		Lang:       NewRuleLangFromProto(sr.GetLanguage()),
		Expression: sr.Expression,
	}
}

func (r Rule) ToProto() *protos.Rule {
	return &protos.Rule{
		Language:   r.Lang.ToProto(),
		Expression: r.Expression,
	}
}
