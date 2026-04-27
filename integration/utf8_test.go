//go:build integration_utf8

package integration

import (
	"context"
	"testing"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/rulefmt"
	"github.com/stretchr/testify/require"

	"github.com/cortexproject/cortex-tools/pkg/client"
	"github.com/cortexproject/cortex-tools/pkg/rules"
	"github.com/cortexproject/cortex-tools/pkg/rules/rwrulefmt"
)

func TestUTF8RulesLoadListDelete(t *testing.T) {
	ctx := context.Background()
	c, err := client.New(client.Config{
		Address: cortexAddress(),
		ID:      "fake",
	})
	require.NoError(t, err)

	namespace := "utf8_test_namespace"
	group := rwrulefmt.RuleGroup{}
	group.Name = "utf8_rule_group"
	group.Rules = []rulefmt.Rule{
		{Record: "my.dotted.metric", Expr: "sum(up)"},
	}

	err = c.CreateRuleGroup(ctx, namespace, group)
	require.NoError(t, err, "CreateRuleGroup with dotted metric name should succeed on UTF-8 Cortex")

	ruleSet, err := c.ListRules(ctx, "")
	require.NoError(t, err)
	require.Contains(t, ruleSet, namespace)

	rg, err := c.GetRuleGroup(ctx, namespace, "utf8_rule_group")
	require.NoError(t, err)
	require.Equal(t, "utf8_rule_group", rg.Name)

	err = c.DeleteRuleNamespace(ctx, namespace)
	require.NoError(t, err)
}

func TestUTF8ParseBytes(t *testing.T) {
	t.Run("utf8 scheme accepts dotted metric names", func(t *testing.T) {
		content := []byte("groups:\n- name: test\n  rules:\n  - record: my.dotted.metric\n    expr: sum(up)\n")
		nss, errs := rules.ParseBytes(content, model.UTF8Validation)
		require.Empty(t, errs)
		require.Len(t, nss, 1)
		require.Equal(t, "my.dotted.metric", nss[0].Groups[0].Rules[0].Record)
	})

	t.Run("legacy scheme rejects dotted metric names", func(t *testing.T) {
		content := []byte("groups:\n- name: test\n  rules:\n  - record: my.dotted.metric\n    expr: sum(up)\n")
		_, errs := rules.ParseBytes(content, model.LegacyValidation)
		require.NotEmpty(t, errs)
	})
}
