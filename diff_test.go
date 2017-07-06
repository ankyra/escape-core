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
		c.Assert(changes[0].Field, DeepEquals, test[0].(string)+" field 'newfile.txt'")
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
		c.Assert(changes[0].Field, DeepEquals, "Stages field 'test'")
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
		[]interface{}{errand1, errand2, Change{Removed: false, Added: false, Field: "Errands field 'test' field Script"}, "test.sh", "test2.sh"},
		[]interface{}{errand1, errand3, Change{Removed: false, Added: false, Field: "Errands field 'test' field Description"}, "", "Description"},
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
