package runtime

import (
	"mokapi/providers/pipeline/lang/ast"
	"mokapi/providers/pipeline/lang/parser"
	"mokapi/providers/pipeline/lang/types"
	"reflect"
	"testing"
)

type testCallback func(i interface{})

type testStep struct {
	types.AbstractStep
	callback testCallback
}

type testExecution struct {
	Message  interface{} `step:"message,position=0,required"`
	callback testCallback
}

func (e *testStep) Start() types.StepExecution {
	return &testExecution{callback: e.callback}
}

func (e *testExecution) Run(_ types.StepContext) (interface{}, error) {
	e.callback(e.Message)
	return nil, nil
}

func TestRunPipeline(t *testing.T) {
	var scope *ast.Scope
	var s []interface{}
	type test func(t *testing.T, src string)
	type setup func()
	callback := func(i interface{}) { s = append(s, i) }
	data := []struct {
		src    string
		verify test
		init   setup
	}{
		{"pipeline(){stages{stage(){steps{x := 12; test x}}}}",
			func(t *testing.T, src string) {
				e := 12.0
				x := s[0]
				if !reflect.DeepEqual(e, x) {
					t.Errorf("source(%q): got %v, expected %v", src, x, e)
				}
			},
			func() {},
		},
		{"pipeline(){stages{stage(){steps{x := [1,2,3,4]; test x}}}}",
			func(t *testing.T, src string) {
				e := []float64{1.0, 2.0, 3.0, 4.0}
				x := s[0].([]interface{})
				result := make([]float64, len(e))
				for i := range result {
					result[i] = x[i].(float64)
				}
				if !reflect.DeepEqual(e, result) {
					t.Errorf("source(%q): got %v, expected %v", src, x, e)
				}
			},
			func() {},
		},
		{"pipeline(){stages{stage(){steps{x := [1,2,3,4]; x = x.findAll {y => y > 2}; test x}}}}",
			func(t *testing.T, src string) {
				e := []float64{3.0, 4.0}
				x := s[0].([]interface{})
				result := make([]float64, len(e))
				for i := range result {
					result[i] = x[i].(float64)
				}
				if !reflect.DeepEqual(e, result) {
					t.Errorf("source(%q): got %v, expected %v", src, x, e)
				}
			},
			func() {},
		},
		{"pipeline(){stages{stage(){steps{x := [a: 1,b: 2,c: 3, d: 4]; test x}}}}",
			func(t *testing.T, src string) {
				e := map[string]interface{}{"a": 1.0, "b": 2.0, "c": 3.0, "d": 4.0}
				x := s[0]
				if !reflect.DeepEqual(e, x) {
					t.Errorf("source(%q): got %v, expected %v", src, x, e)
				}
			},
			func() {},
		},
		{"pipeline(){stages{stage(){steps{ids := [1, 2]; r := data.'*'.findAll {x => ids.contains x.id}; test r}}}}",
			func(t *testing.T, src string) {
				e := []map[string]interface{}{{"id": 1.0}, {"id": 2.0}}
				x := s[0].([]interface{})
				result := make([]map[string]interface{}, len(e))
				for i := range result {
					result[i] = x[i].(map[string]interface{})
				}
				if !reflect.DeepEqual(e, result) {
					t.Errorf("source(%q): got %v, expected %v", src, result, e)
				}
			},
			func() {
				a := types.NewArray()
				for i := 1; i < 3; i++ {
					expando := types.NewExpando()
					expando.SetField("id", types.NewNumber(float64(i)))
					a.Add(expando)
				}
				scope.SetSymbol("data", a)
			},
		},
		{"pipeline(){stages{stage(){steps{users := [[name: 'bob'],[name: 'sarah']]; r := users.'*'.select {x => x.name}; test r}}}}",
			func(t *testing.T, src string) {
				e := []string{"bob", "sarah"}
				x := s[0].([]interface{})
				result := make([]string, len(e))
				for i := range result {
					result[i] = x[i].(string)
				}
				if !reflect.DeepEqual(e, result) {
					t.Errorf("source(%q): got %v, expected %v", src, result, e)
				}
			},
			func() {
			},
		},
	}

	for _, d := range data {
		s = make([]interface{}, 0)
		scope = ast.NewScope(map[string]types.Object{"test": &testStep{callback: callback}})
		d.init()
		f, err := parser.ParseFile([]byte(d.src), scope)
		if err != nil {
			t.Errorf("ParseExpr(%q):%v", d.src, err)
		}
		err = RunPipeline(f, "")
		if err != nil {
			t.Errorf("source(%q):%v", d.src, err)
		}
		d.verify(t, d.src)
	}

}
