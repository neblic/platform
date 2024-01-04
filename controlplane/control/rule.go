package control

import (
	"github.com/neblic/platform/controlplane/protos"
	"gopkg.in/yaml.v3"
)

type SamplerStreamRuleUID string

type RuleLang int

const (
	SrlUnknown RuleLang = iota
	SrlCel
)

func NewRuleLangFromString(s string) RuleLang {
	switch s {
	case "UNKNOWN":
		return SrlUnknown
	case "CEL":
		return SrlCel
	default:
		return SrlUnknown
	}
}

func (srl RuleLang) String() string {
	switch srl {
	case SrlCel:
		return "CEL"
	case SrlUnknown:
		return "UNKNOWN"
	default:
		return "UNKNOWN"
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

func (srl RuleLang) MarshalYAML() (interface{}, error) {
	return srl.String(), nil
}

func (srl *RuleLang) UnmarshalYAML(value *yaml.Node) error {
	*srl = NewRuleLangFromString(value.Value)

	return nil
}

type Rule struct {
	Lang       RuleLang
	Expression string
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
