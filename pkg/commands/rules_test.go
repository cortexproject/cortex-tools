package commands

import (
	"testing"

	"github.com/prometheus/prometheus/model/rulefmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cortexproject/cortex-tools/pkg/rules/rwrulefmt"
)

func TestCheckDuplicates(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   []rwrulefmt.RuleGroup
		want []compareRuleType
	}{
		{
			name: "no duplicates",
			in: []rwrulefmt.RuleGroup{{
				RuleGroup: rulefmt.RuleGroup{
					Name: "rulegroup",
					Rules: []rulefmt.Rule{
						{
							Record: "up",
							Expr:   "up==1",
						},
						{
							Record: "down",
							Expr:   "up==0",
						},
					},
				},
				RWConfigs: []rwrulefmt.RemoteWriteConfig{},
			}},
			want: nil,
		},
		{
			name: "with duplicates",
			in: []rwrulefmt.RuleGroup{{
				RuleGroup: rulefmt.RuleGroup{
					Name: "rulegroup",
					Rules: []rulefmt.Rule{
						{
							Record: "up",
							Expr:   "up==1",
						},
						{
							Record: "up",
							Expr:   "up==0",
						},
					},
				},
				RWConfigs: []rwrulefmt.RemoteWriteConfig{},
			}},
			want: []compareRuleType{{metric: "up", label: map[string]string(nil)}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, checkDuplicates(tc.in))
		})
	}
}

// TestSetupClientRequiresAddress verifies that setupClient returns an error when
// no address is configured, while setup (the PreAction) does not.
// This ensures local-only commands (lint, prepare, check) work without a Cortex address.
func TestSetupClientRequiresAddress(t *testing.T) {
	r := &RuleCommand{}

	// setupClient should fail when no address is set.
	err := r.setupClient()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cortex address is required")

	// setupClient should fail when address is set but tenant ID is missing.
	r.ClientConfig.Address = "http://cortex:9009"
	err = r.setupClient()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tenant ID is required")
}
