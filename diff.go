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
	"fmt"
	"reflect"
)

type Change struct {
	Field         string
	PreviousValue interface{}
	NewValue      interface{}
	Added         bool
	Removed       bool
}

type Changes []Change

func Diff(this *ReleaseMetadata, other *ReleaseMetadata) Changes {
	result := []Change{}
	thisVal := reflect.Indirect(reflect.ValueOf(this))
	otherVal := reflect.Indirect(reflect.ValueOf(other))
	fields := thisVal.Type().NumField()
	for i := 0; i < fields; i++ {
		name := thisVal.Type().Field(i).Name
		oldValue := thisVal.Field(i).Interface()
		newValue := otherVal.FieldByName(name).Interface()
		for _, change := range diff(name, oldValue, newValue) {
			result = append(result, change)
		}
	}
	return result
}

func diff(name string, oldValue, newValue interface{}) Changes {
	thisVal := reflect.ValueOf(oldValue)
	typ := thisVal.Type().String()
	kind := thisVal.Type().Kind().String()
	if typ == "int" || typ == "string" {
		if r := diffSimpleType(name, oldValue, newValue); r != nil {
			return []Change{*r}
		}
	} else if typ == "*core.ExecStage" {
		return diffExecStage(name, oldValue, newValue)
	} else if kind == "map" {
		return diffMap(name, oldValue, newValue)
	} else if kind == "slice" {
		return diffSlice(name, oldValue, newValue)
	} else {
		fmt.Printf("WARN: Undiffable type '%s' (%s) for field '%s'\n", typ, kind, name)
	}
	return []Change{}
}

func diffSimpleType(name string, oldValue, newValue interface{}) *Change {
	if !reflect.DeepEqual(oldValue, newValue) {
		return &Change{name, oldValue, newValue, false, false}
	}
	return nil
}

func diffMap(name string, oldValue, newValue interface{}) []Change {
	changes := []Change{}
	if reflect.DeepEqual(oldValue, newValue) {
		return changes
	}
	oldMap := reflect.ValueOf(oldValue)
	newMap := reflect.ValueOf(newValue)

	for _, key := range oldMap.MapKeys() {
		oldVal := oldMap.MapIndex(key).Interface()
		newVal := newMap.MapIndex(key)
		field := fmt.Sprintf("%s field '%s'", name, key)
		if !newVal.IsValid() {
			changes = append(changes, Change{field, diffValue(oldVal), nil, false, true})
			continue
		}
		newValI := newVal.Interface()
		if reflect.DeepEqual(oldVal, newValI) {
			continue
		}
		for _, c := range diff(field, oldVal, newValI) {
			changes = append(changes, c)
		}
	}
	for _, key := range newMap.MapKeys() {
		oldVal := oldMap.MapIndex(key)
		newVal := newMap.MapIndex(key).Interface()
		if !oldVal.IsValid() {
			field := fmt.Sprintf("%s field '%s'", name, key)
			changes = append(changes, Change{field, nil, diffValue(newVal), true, false})
		} else {
			fmt.Printf("%s didn't change %v\n", key, oldVal)
		}
	}
	return changes
}
func diffSlice(name string, oldValue, newValue interface{}) []Change {
	if reflect.DeepEqual(oldValue, newValue) {
		return nil
	}
	return nil
}

func diffExecStage(name string, oldValue, newValue interface{}) []Change {
	if reflect.DeepEqual(oldValue, newValue) {
		return nil
	}
	oldStage := oldValue.(*ExecStage)
	newStage := newValue.(*ExecStage)
	return diff(name, oldStage.Script, newStage.Script)
}

func diffValue(v interface{}) interface{} {
	switch v.(type) {
	case string, int:
		return v
	case *ExecStage:
		return v.(*ExecStage).Script

	}
	return v
}
