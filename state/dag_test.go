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

package state

import (
	. "gopkg.in/check.v1"
)

func (s *suite) Test_GetDeploymentStateDAG_empty_env(c *C) {
	prj, _ := NewProjectState("my-project")
	env := prj.GetEnvironmentStateOrMakeNew("my-env")
	dag, err := env.GetDeploymentStateDAG("build")
	c.Assert(err, IsNil)
	c.Assert(dag, HasLen, 0)
}

func (s *suite) Test_GetDeploymentStateDAG_one_deployment(c *C) {
	stage := "deploy"
	prj, _ := NewProjectState("my-project")
	env := prj.GetEnvironmentStateOrMakeNew("my-env")
	depl1 := env.GetOrCreateDeploymentState("depl1")
	depl1.GetStageOrCreateNew(stage)
	dag, err := env.GetDeploymentStateDAG(stage)
	c.Assert(err, IsNil)
	c.Assert(dag, HasLen, 1)
	c.Assert(dag[0].Node, DeepEquals, depl1)
	c.Assert(dag[0].AndThen, HasLen, 0)
}

func (s *suite) Test_GetDeploymentStateDAG_two_deployments_one_provider(c *C) {
	stage := "deploy"
	prj, _ := NewProjectState("my-project")
	env := prj.GetEnvironmentStateOrMakeNew("my-env")
	depl1 := env.GetOrCreateDeploymentState("depl1")
	depl2 := env.GetOrCreateDeploymentState("depl2")
	st := depl1.GetStageOrCreateNew(stage)
	st.Providers["whatever"] = "depl2"
	depl2.GetStageOrCreateNew(stage)

	dag, err := env.GetDeploymentStateDAG(stage)
	c.Assert(err, IsNil)
	c.Assert(dag, HasLen, 1)
	c.Assert(dag[0].Node, DeepEquals, depl2)
	c.Assert(dag[0].AndThen, HasLen, 1)
	c.Assert(dag[0].AndThen[0].Node, DeepEquals, depl1)
	c.Assert(dag[0].AndThen[0].AndThen, HasLen, 0)
}

type hasItemChecker struct{}

var HasItem = &hasItemChecker{}

func (*hasItemChecker) Info() *CheckerInfo {
	return &CheckerInfo{Name: "HasItem", Params: []string{"obtained", "expected to have item"}}
}
func (*hasItemChecker) Check(params []interface{}, names []string) (bool, string) {
	obtained := params[0]
	expectedItem := params[1]
	switch obtained.(type) {
	case []interface{}:
		for _, v := range obtained.([]interface{}) {
			if v == expectedItem {
				return true, ""
			}
		}
	case []string:
		for _, v := range obtained.([]string) {
			if v == expectedItem {
				return true, ""
			}
		}
	default:
		return false, "Unexpected type."
	}
	return false, "Item not found"
}
