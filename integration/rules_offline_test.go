//go:build integration

package integration

import (
	"testing"

	"github.com/prometheus/prometheus/model/rulefmt"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/cortexproject/cortex-tools/pkg/rules"
	"github.com/cortexproject/cortex-tools/pkg/rules/rwrulefmt"
)

func TestRulesLint(t *testing.T) {
	t.Run("valid expression is unchanged", func(t *testing.T) {
		ns := rules.RuleNamespace{
			Groups: []rwrulefmt.RuleGroup{
				{RuleGroup: rulefmt.RuleGroup{
					Name:  "test",
					Rules: []rulefmt.RuleNode{ruleNode("test_metric", "sum(up)")},
				}},
			},
		}

		count, mod, err := ns.LintExpressions()
		require.NoError(t, err)
		require.Equal(t, 1, count, "should evaluate 1 rule")
		require.Equal(t, 0, mod, "canonical expression should not be modified")
	})

	t.Run("expression is reformatted", func(t *testing.T) {
		ns := rules.RuleNamespace{
			Groups: []rwrulefmt.RuleGroup{
				{RuleGroup: rulefmt.RuleGroup{
					Name: "test",
					Rules: []rulefmt.RuleNode{
						ruleNode("test_metric", "sum  by(job)(up)"),
					},
				}},
			},
		}

		count, mod, err := ns.LintExpressions()
		require.NoError(t, err)
		require.Equal(t, 1, count)
		require.Equal(t, 1, mod, "expression with extra whitespace should be reformatted")
		require.Equal(t, "sum by (job) (up)", ns.Groups[0].Rules[0].Expr.Value)
	})

	t.Run("invalid expression returns error", func(t *testing.T) {
		ns := rules.RuleNamespace{
			Groups: []rwrulefmt.RuleGroup{
				{RuleGroup: rulefmt.RuleGroup{
					Name:  "test",
					Rules: []rulefmt.RuleNode{ruleNode("test_metric", "not a valid expr !!!")},
				}},
			},
		}

		_, _, err := ns.LintExpressions()
		require.Error(t, err)
	})
}

func TestRulesCheck(t *testing.T) {
	t.Run("valid recording rule name", func(t *testing.T) {
		ns := rules.RuleNamespace{
			Groups: []rwrulefmt.RuleGroup{
				{RuleGroup: rulefmt.RuleGroup{
					Name:  "test",
					Rules: []rulefmt.RuleNode{ruleNode("level:metric:operation", "sum(up)")},
				}},
			},
		}

		count := ns.CheckRecordingRules(false)
		require.Equal(t, 0, count, "valid recording rule name should pass")
	})

	t.Run("invalid recording rule name", func(t *testing.T) {
		ns := rules.RuleNamespace{
			Groups: []rwrulefmt.RuleGroup{
				{RuleGroup: rulefmt.RuleGroup{
					Name:  "test",
					Rules: []rulefmt.RuleNode{ruleNode("no_colons_here", "sum(up)")},
				}},
			},
		}

		count := ns.CheckRecordingRules(false)
		require.Equal(t, 1, count, "recording rule without colon should fail")
	})

	t.Run("strict mode requires two colons", func(t *testing.T) {
		ns := rules.RuleNamespace{
			Groups: []rwrulefmt.RuleGroup{
				{RuleGroup: rulefmt.RuleGroup{
					Name:  "test",
					Rules: []rulefmt.RuleNode{ruleNode("level:metric", "sum(up)")},
				}},
			},
		}

		count := ns.CheckRecordingRules(true)
		require.Equal(t, 1, count, "strict mode should require level:metric:operation format")

		countNonStrict := ns.CheckRecordingRules(false)
		require.Equal(t, 0, countNonStrict, "non-strict should pass with one colon")
	})
}

func TestRulesPrepare(t *testing.T) {
	t.Run("adds aggregation label", func(t *testing.T) {
		ns := rules.RuleNamespace{
			Groups: []rwrulefmt.RuleGroup{
				{RuleGroup: rulefmt.RuleGroup{
					Name: "test",
					Rules: []rulefmt.RuleNode{
						{
							Record: yaml.Node{Kind: yaml.ScalarNode, Value: "test_metric"},
							Expr:   yaml.Node{Kind: yaml.ScalarNode, Value: "sum by (job) (up)"},
						},
					},
				}},
			},
		}

		count, mod, err := ns.AggregateBy("cluster", nil)
		require.NoError(t, err)
		require.Equal(t, 1, count)
		require.Equal(t, 1, mod, "should modify expression to include cluster label")
		require.Contains(t, ns.Groups[0].Rules[0].Expr.Value, "cluster")
	})

	t.Run("skips if label already present", func(t *testing.T) {
		ns := rules.RuleNamespace{
			Groups: []rwrulefmt.RuleGroup{
				{RuleGroup: rulefmt.RuleGroup{
					Name: "test",
					Rules: []rulefmt.RuleNode{
						{
							Record: yaml.Node{Kind: yaml.ScalarNode, Value: "test_metric"},
							Expr:   yaml.Node{Kind: yaml.ScalarNode, Value: "sum by (job, cluster) (up)"},
						},
					},
				}},
			},
		}

		count, mod, err := ns.AggregateBy("cluster", nil)
		require.NoError(t, err)
		require.Equal(t, 1, count)
		require.Equal(t, 0, mod, "should not modify when label already present")
	})

	t.Run("respects applyTo filter", func(t *testing.T) {
		ns := rules.RuleNamespace{
			Groups: []rwrulefmt.RuleGroup{
				{RuleGroup: rulefmt.RuleGroup{
					Name: "excluded_group",
					Rules: []rulefmt.RuleNode{
						{
							Record: yaml.Node{Kind: yaml.ScalarNode, Value: "test_metric"},
							Expr:   yaml.Node{Kind: yaml.ScalarNode, Value: "sum by (job) (up)"},
						},
					},
				}},
			},
		}

		applyTo := func(group rwrulefmt.RuleGroup, _ rulefmt.RuleNode) bool {
			return group.Name != "excluded_group"
		}

		count, mod, err := ns.AggregateBy("cluster", applyTo)
		require.NoError(t, err)
		require.Equal(t, 1, count)
		require.Equal(t, 0, mod, "excluded group should not be modified")
	})
}

func TestRulesParseFile(t *testing.T) {
	t.Run("parse basic namespace", func(t *testing.T) {
		nss, errs := rules.Parse("../pkg/rules/testdata/basic_namespace.yaml")
		require.Empty(t, errs)
		require.Len(t, nss, 1)
		require.Equal(t, "example_namespace", nss[0].Namespace)
		require.Len(t, nss[0].Groups, 1)
		require.Equal(t, "example_rule_group", nss[0].Groups[0].Name)
	})

	t.Run("parse multiple namespaces", func(t *testing.T) {
		nss, errs := rules.Parse("../pkg/rules/testdata/multiple_namespace.yaml")
		require.Empty(t, errs)
		require.Len(t, nss, 2)
		require.Equal(t, "example_namespace", nss[0].Namespace)
		require.Equal(t, "other_example_namespace", nss[1].Namespace)
	})

	t.Run("validate catches errors", func(t *testing.T) {
		content := []byte("groups:\n- name: \"\"\n  rules:\n  - expr: sum(up)\n    record: test\n")
		nss, errs := rules.ParseBytes(content)
		require.NotEmpty(t, errs, "empty group name should fail validation")
		require.Nil(t, nss)
	})
}
