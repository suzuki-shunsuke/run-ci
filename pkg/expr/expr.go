package expr

import (
	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
)

type Expr struct {
	prog *vm.Program
}

func New(expression string) (Expr, error) {
	prog, err := expr.Compile(expression, expr.AsBool())
	if err != nil {
		return Expr{}, err
	}
	return Expr{
		prog: prog,
	}, nil
}

func (ex Expr) Match(params interface{}) (bool, error) {
	output, err := expr.Run(ex.prog, params)
	if err != nil {
		return false, err
	}
	if f, ok := output.(bool); !ok || !f {
		return false, nil
	}
	return true, nil
}

func LabelNames(labels interface{}) []string {
	if labels == nil {
		return []string{}
	}
	a := labels.([]interface{})
	b := make([]string, len(a))
	for i, label := range a {
		b[i] = label.(map[string]interface{})["name"].(string)
	}
	return b
}
