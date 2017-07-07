/*
Copyright 2017 Ankyra

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package core

import (
	"github.com/ankyra/escape-core/templates"
	"github.com/ankyra/escape-core/variables"
	. "gopkg.in/check.v1"
	"reflect"
)

func (s *metadataSuite) Test_Diff_simple_types(c *C) {
	testCases := [][]interface{}{
		[]interface{}{"ApiVersion", 1, 2},
		[]interface{}{"Branch", "test", "not-test"},
		[]interface{}{"Description", "test", "not-test"},
		[]interface{}{"Logo", "test", "not-test"},
		[]interface{}{"Name", "test", "not-test"},
		[]interface{}{"Project", "test", "not-test"},
		[]interface{}{"Revision", "test", "not-test"},
		[]interface{}{"Version", "1.0", "1.0.0"},
		[]interface{}{"Repository", "test", "not-test"},
	}
	for _, test := range testCases {
		m1 := NewReleaseMetadata("test", "1.0")
		m2 := NewReleaseMetadata("test", "1.0")

		thisVal := reflect.Indirect(reflect.ValueOf(m1))
		otherVal := reflect.Indirect(reflect.ValueOf(m2))

		thisVal.FieldByName(test[0].(string)).Set(reflect.ValueOf(test[1]))
		otherVal.FieldByName(test[0].(string)).Set(reflect.ValueOf(test[2]))

		changes := Diff(m1, m2)
		c.Assert(changes, HasLen, 1, Commentf("Field %s", test[0]))
		c.Assert(changes[0].Field, DeepEquals, test[0])
		c.Assert(changes[0].PreviousValue, DeepEquals, test[1], Commentf("Field %s", test[0]))
		c.Assert(changes[0].NewValue, DeepEquals, test[2], Commentf("Field %s", test[1]))
	}
}

func (s *metadataSuite) Test_Diff_maps(c *C) {
	emptyDict := map[string]string{}
	oldDict := map[string]string{
		"newfile.txt": "123",
	}
	newDict := map[string]string{
		"newfile.txt": "123123123",
	}
	testCases := [][]interface{}{

		[]interface{}{"Files", oldDict, newDict, Change{Removed: false, Added: false}},
		[]interface{}{"Files", emptyDict, newDict, Change{Removed: false, Added: true}},
		[]interface{}{"Files", oldDict, emptyDict, Change{Removed: true, Added: false}},

		[]interface{}{"Metadata", oldDict, newDict, Change{Removed: false, Added: false}},
		[]interface{}{"Metadata", emptyDict, newDict, Change{Removed: false, Added: true}},
		[]interface{}{"Metadata", oldDict, emptyDict, Change{Removed: true, Added: false}},

		[]interface{}{"VariableCtx", oldDict, newDict, Change{Removed: false, Added: false}},
		[]interface{}{"VariableCtx", emptyDict, newDict, Change{Removed: false, Added: true}},
		[]interface{}{"VariableCtx", oldDict, emptyDict, Change{Removed: true, Added: false}},
	}
	for _, test := range testCases {
		m1 := NewReleaseMetadata("test", "1.0")
		m2 := NewReleaseMetadata("test", "1.0")

		thisVal := reflect.Indirect(reflect.ValueOf(m1))
		otherVal := reflect.Indirect(reflect.ValueOf(m2))

		thisVal.FieldByName(test[0].(string)).Set(reflect.ValueOf(test[1]))
		otherVal.FieldByName(test[0].(string)).Set(reflect.ValueOf(test[2]))

		expected := test[3].(Change)
		changes := Diff(m1, m2)
		c.Assert(changes, HasLen, 1, Commentf("Field %s %v %v", test[0], expected.Removed, expected.Added))
		c.Assert(changes[0].Field, DeepEquals, test[0].(string)+`["newfile.txt"]`)
		c.Assert(changes[0].Removed, DeepEquals, expected.Removed)
		c.Assert(changes[0].Added, DeepEquals, expected.Added)
		if !expected.Removed {
			c.Assert(changes[0].NewValue, DeepEquals, "123123123", Commentf("Field %s", test[0]))
		}
		if !expected.Added {
			c.Assert(changes[0].PreviousValue, DeepEquals, "123", Commentf("Field %s", test[0]))
		}
	}
}

func (s *metadataSuite) Test_Diff_Stages(c *C) {
	emptyDict := map[string]*ExecStage{}
	oldDict := map[string]*ExecStage{
		"test": &ExecStage{Script: "test.sh"},
	}
	newDict := map[string]*ExecStage{
		"test": &ExecStage{Script: "test2.sh"},
	}
	testCases := [][]interface{}{
		[]interface{}{oldDict, newDict, Change{Removed: false, Added: false}},
		[]interface{}{emptyDict, newDict, Change{Removed: false, Added: true}},
		[]interface{}{oldDict, emptyDict, Change{Removed: true, Added: false}},
	}
	for _, test := range testCases {
		m1 := NewReleaseMetadata("test", "1.0")
		m2 := NewReleaseMetadata("test", "1.0")

		thisVal := reflect.Indirect(reflect.ValueOf(m1))
		otherVal := reflect.Indirect(reflect.ValueOf(m2))

		thisVal.FieldByName("Stages").Set(reflect.ValueOf(test[0]))
		otherVal.FieldByName("Stages").Set(reflect.ValueOf(test[1]))

		expected := test[2].(Change)
		changes := Diff(m1, m2)
		c.Assert(changes, HasLen, 1, Commentf("Field Stages"))
		c.Assert(changes[0].Field, DeepEquals, `Stages["test"]`)
		c.Assert(changes[0].Removed, DeepEquals, expected.Removed)
		c.Assert(changes[0].Added, DeepEquals, expected.Added)
		if !expected.Added {
			c.Assert(changes[0].PreviousValue, DeepEquals, "test.sh")
		}
		if !expected.Removed {
			c.Assert(changes[0].NewValue, DeepEquals, "test2.sh")
		}
	}
}

func (s *metadataSuite) Test_Diff_Errands(c *C) {
	errand1 := map[interface{}]interface{}{
		"script": "test.sh",
	}
	errand2 := map[interface{}]interface{}{
		"script": "test2.sh",
	}
	errand3 := map[interface{}]interface{}{
		"script":      "test.sh",
		"description": "Description",
	}

	testCases := [][]interface{}{
		[]interface{}{errand1, errand2, Change{Removed: false, Added: false, Field: `Errands["test"].Script`}, "test.sh", "test2.sh"},
		[]interface{}{errand1, errand3, Change{Removed: false, Added: false, Field: `Errands["test"].Description`}, "", "Description"},
	}
	for _, test := range testCases {
		m1 := NewReleaseMetadata("test", "1.0")
		m2 := NewReleaseMetadata("test", "1.0")

		e1, err := NewErrandFromDict("test", test[0].(map[interface{}]interface{}))
		e2, err2 := NewErrandFromDict("test", test[1].(map[interface{}]interface{}))
		c.Assert(err, IsNil)
		c.Assert(err2, IsNil)
		m1.Errands["test"] = e1
		m2.Errands["test"] = e2

		expected := test[2].(Change)
		changes := Diff(m1, m2)
		c.Assert(changes, HasLen, 1, Commentf("Field %s", test[0]))
		c.Assert(changes[0].Field, DeepEquals, expected.Field)
		c.Assert(changes[0].Removed, DeepEquals, expected.Removed)
		c.Assert(changes[0].Added, DeepEquals, expected.Added, Commentf(expected.Field))
		if !expected.Added {
			c.Assert(changes[0].PreviousValue, DeepEquals, test[3], Commentf("Field %s", expected.Field))
		}
		if !expected.Removed {
			c.Assert(changes[0].NewValue, DeepEquals, test[4], Commentf("Field %s", expected.Field))
		}
	}
}

func (s *metadataSuite) Test_Diff_Variables(c *C) {
	empty := []map[interface{}]interface{}{}
	var1 := []map[interface{}]interface{}{
		map[interface{}]interface{}{
			"id": "test",
		},
	}
	var2 := []map[interface{}]interface{}{
		map[interface{}]interface{}{
			"id": "test2",
		},
	}
	var3 := []map[interface{}]interface{}{
		map[interface{}]interface{}{
			"id":   "test",
			"type": "integer",
		},
	}
	var4 := []map[interface{}]interface{}{
		map[interface{}]interface{}{
			"id": "test2",
		},
		map[interface{}]interface{}{
			"id": "test",
		},
	}
	v1, err := variables.NewVariableFromDict(var1[0])
	c.Assert(err, IsNil)

	testCases := [][]interface{}{
		[]interface{}{"Inputs", var1, var2, Change{Removed: false, Added: false, Field: "Inputs[0].Id"}, "test", "test2"},
		[]interface{}{"Inputs", var1, var3, Change{Removed: false, Added: false, Field: "Inputs[0].Type"}, "string", "integer"},
		[]interface{}{"Inputs", var1, empty, Change{Removed: true, Added: false, Field: "Inputs"}, v1, "integer"},
		[]interface{}{"Inputs", empty, var1, Change{Removed: false, Added: true, Field: "Inputs"}, nil, v1},
		[]interface{}{"Inputs", var4, var2, Change{Removed: true, Added: false, Field: "Inputs"}, v1, nil},
		[]interface{}{"Inputs", var2, var4, Change{Removed: false, Added: true, Field: "Inputs"}, nil, v1},

		[]interface{}{"Outputs", var1, var2, Change{Removed: false, Added: false, Field: "Outputs[0].Id"}, "test", "test2"},
		[]interface{}{"Outputs", var1, var3, Change{Removed: false, Added: false, Field: "Outputs[0].Type"}, "string", "integer"},
		[]interface{}{"Outputs", var1, empty, Change{Removed: true, Added: false, Field: "Outputs"}, v1, "integer"},
		[]interface{}{"Outputs", empty, var1, Change{Removed: false, Added: true, Field: "Outputs"}, nil, v1},
		[]interface{}{"Outputs", var4, var2, Change{Removed: true, Added: false, Field: "Outputs"}, v1, nil},
		[]interface{}{"Outputs", var2, var4, Change{Removed: false, Added: true, Field: "Outputs"}, nil, v1},

		[]interface{}{"Errands", var1, var2, Change{Removed: false, Added: false, Field: `Errands["test"].Inputs[0].Id`}, "test", "test2"},
		[]interface{}{"Errands", var1, var3, Change{Removed: false, Added: false, Field: `Errands["test"].Inputs[0].Type`}, "string", "integer"},
		[]interface{}{"Errands", var1, empty, Change{Removed: true, Added: false, Field: `Errands["test"].Inputs`}, v1, "integer"},
		[]interface{}{"Errands", empty, var1, Change{Removed: false, Added: true, Field: `Errands["test"].Inputs`}, nil, v1},
		[]interface{}{"Errands", var4, var2, Change{Removed: true, Added: false, Field: `Errands["test"].Inputs`}, v1, nil},
		[]interface{}{"Errands", var2, var4, Change{Removed: false, Added: true, Field: `Errands["test"].Inputs`}, nil, v1},
	}
	for _, test := range testCases {
		errand1, err := NewErrandFromDict("test", map[interface{}]interface{}{
			"script": "test.sh",
		})
		c.Assert(err, IsNil)
		errand2, err := NewErrandFromDict("test", map[interface{}]interface{}{
			"script": "test.sh",
		})
		c.Assert(err, IsNil)
		m1 := NewReleaseMetadata("test", "1.0")
		m2 := NewReleaseMetadata("test", "1.0")
		m1.Errands["test"] = errand1
		m2.Errands["test"] = errand2
		typ := test[0].(string)
		for _, varDict := range test[1].([]map[interface{}]interface{}) {
			v, err := variables.NewVariableFromDict(varDict)
			c.Assert(err, IsNil)
			if typ == "Inputs" {
				m1.AddInputVariable(v)
			} else if typ == "Outputs" {
				m1.AddOutputVariable(v)
			} else {
				errand1.Inputs = append(errand1.Inputs, v)
			}
		}
		for _, varDict := range test[2].([]map[interface{}]interface{}) {
			v, err := variables.NewVariableFromDict(varDict)
			c.Assert(err, IsNil)
			if typ == "Inputs" {
				m2.AddInputVariable(v)
			} else if typ == "Outputs" {
				m2.AddOutputVariable(v)
			} else {
				errand2.Inputs = append(errand2.Inputs, v)
			}
		}

		expected := test[3].(Change)
		expectedField := expected.Field
		changes := Diff(m1, m2)
		c.Assert(changes, HasLen, 1, Commentf(expectedField))
		c.Assert(changes[0].Field, DeepEquals, expectedField)
		c.Assert(changes[0].Removed, DeepEquals, expected.Removed)
		c.Assert(changes[0].Added, DeepEquals, expected.Added, Commentf(expectedField))
		if !expected.Added {
			c.Assert(changes[0].PreviousValue, DeepEquals, test[4], Commentf("Field %s", expectedField))
		}
		if !expected.Removed {
			c.Assert(changes[0].NewValue, DeepEquals, test[5], Commentf("Field %s", expectedField))
		}
	}
}

func (s *metadataSuite) Test_Diff_Slices(c *C) {
	testCases := [][]interface{}{
		[]interface{}{"Consumes", []string{"test"}, []string{}, Change{Removed: true, Added: false, Field: "Consumes"}, "test", ""},
		[]interface{}{"Consumes", []string{}, []string{"test"}, Change{Removed: false, Added: true, Field: "Consumes"}, "", "test"},
		[]interface{}{"Consumes", []string{"test"}, []string{"kubernetes"}, Change{Removed: false, Added: false, Field: "Consumes[0]"}, "test", "kubernetes"},

		[]interface{}{"Provides", []string{"test"}, []string{}, Change{Removed: true, Added: false, Field: "Provides"}, "test", ""},
		[]interface{}{"Provides", []string{}, []string{"test"}, Change{Removed: false, Added: true, Field: "Provides"}, "", "test"},
		[]interface{}{"Provides", []string{"test"}, []string{"kubernetes"}, Change{Removed: false, Added: false, Field: "Provides[0]"}, "test", "kubernetes"},

		[]interface{}{"Depends", []string{"test"}, []string{}, Change{Removed: true, Added: false, Field: "Depends"}, "test", ""},
		[]interface{}{"Depends", []string{}, []string{"test"}, Change{Removed: false, Added: true, Field: "Depends"}, "", "test"},
		[]interface{}{"Depends", []string{"test"}, []string{"kubernetes"}, Change{Removed: false, Added: false, Field: "Depends[0]"}, "test", "kubernetes"},

		[]interface{}{"Extends", []string{"test"}, []string{}, Change{Removed: true, Added: false, Field: "Extends"}, "test", ""},
		[]interface{}{"Extends", []string{}, []string{"test"}, Change{Removed: false, Added: true, Field: "Extends"}, "", "test"},
		[]interface{}{"Extends", []string{"test"}, []string{"kubernetes"}, Change{Removed: false, Added: false, Field: "Extends[0]"}, "test", "kubernetes"},
	}
	for _, test := range testCases {
		m1 := NewReleaseMetadata("test", "1.0")
		m2 := NewReleaseMetadata("test", "1.0")
		typ := test[0].(string)
		if typ == "Consumes" {
			m1.SetConsumes(test[1].([]string))
			m2.SetConsumes(test[2].([]string))
		} else if typ == "Provides" {
			m1.SetProvides(test[1].([]string))
			m2.SetProvides(test[2].([]string))
		} else if typ == "Depends" {
			m1.SetDependencies(test[1].([]string))
			m2.SetDependencies(test[2].([]string))
		} else if typ == "Extends" {
			for _, consumer := range test[1].([]string) {
				m1.AddExtension(consumer)
			}
			for _, consumer := range test[2].([]string) {
				m2.AddExtension(consumer)
			}
		}
		expected := test[3].(Change)
		expectedField := expected.Field
		changes := Diff(m1, m2)
		c.Assert(changes, HasLen, 1, Commentf(expectedField))
		c.Assert(changes[0].Field, DeepEquals, expectedField)
		c.Assert(changes[0].Removed, DeepEquals, expected.Removed)
		c.Assert(changes[0].Added, DeepEquals, expected.Added, Commentf(expectedField))
		if !expected.Added {
			c.Assert(changes[0].PreviousValue, DeepEquals, test[4], Commentf("Field %s", expectedField))
		}
		if !expected.Removed {
			c.Assert(changes[0].NewValue, DeepEquals, test[5], Commentf("Field %s", expectedField))
		}
	}
}

func (s *metadataSuite) Test_Diff_Templates(c *C) {
	v1 := map[interface{}]interface{}{
		"file":   "test.tpl",
		"target": "test.sh",
	}
	v2 := map[interface{}]interface{}{
		"file":   "test2.tpl",
		"target": "test.sh",
	}
	v3 := map[interface{}]interface{}{
		"file":   "test.tpl",
		"target": "/other",
	}
	testCases := [][]interface{}{
		[]interface{}{v1, v2, Change{Removed: false, Added: false, Field: `Templates[0].File`}, "test.tpl", "test2.tpl"},
		[]interface{}{v1, v3, Change{Removed: false, Added: false, Field: `Templates[0].Target`}, "test.sh", "/other"},
	}
	for _, test := range testCases {
		m1 := NewReleaseMetadata("test", "1.0")
		m2 := NewReleaseMetadata("test", "1.0")

		e1, err := templates.NewTemplateFromInterface(test[0].(map[interface{}]interface{}))
		e2, err2 := templates.NewTemplateFromInterface(test[1].(map[interface{}]interface{}))
		c.Assert(err, IsNil)
		c.Assert(err2, IsNil)
		m1.Templates = append(m1.Templates, e1)
		m2.Templates = append(m2.Templates, e2)

		expected := test[2].(Change)
		changes := Diff(m1, m2)
		c.Assert(changes, HasLen, 1, Commentf(expected.Field))
		c.Assert(changes[0].Field, DeepEquals, expected.Field)
		c.Assert(changes[0].Removed, DeepEquals, expected.Removed)
		c.Assert(changes[0].Added, DeepEquals, expected.Added, Commentf(expected.Field))
		if !expected.Added {
			c.Assert(changes[0].PreviousValue, DeepEquals, test[3], Commentf("Field %s", expected.Field))
		}
		if !expected.Removed {
			c.Assert(changes[0].NewValue, DeepEquals, test[4], Commentf("Field %s", expected.Field))
		}
	}
}
