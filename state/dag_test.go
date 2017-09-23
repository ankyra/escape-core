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

	tsort, err := env.GetDeploymentStateTopologicalSort(stage)
	c.Assert(err, IsNil)
	c.Assert(tsort, HasLen, 2)
	c.Assert(tsort[0], DeepEquals, depl2)
	c.Assert(tsort[1], DeepEquals, depl1)
}

func (s *suite) Test_GetDeploymentStateDAG(c *C) {
	// For deployment graph:
	//
	// A -> B, E
	// B -> C, D
	// C -> D
	// D
	// E

	stage := "deploy"
	prj, _ := NewProjectState("my-project")
	env := prj.GetEnvironmentStateOrMakeNew("my-env")
	deplA := env.GetOrCreateDeploymentState("deplA")
	deplB := env.GetOrCreateDeploymentState("deplB")
	deplC := env.GetOrCreateDeploymentState("deplC")
	deplD := env.GetOrCreateDeploymentState("deplD")
	deplE := env.GetOrCreateDeploymentState("deplE")

	stA := deplA.GetStageOrCreateNew(stage)
	stA.Providers["b"] = "deplB"
	stA.Providers["e"] = "deplE"

	stB := deplB.GetStageOrCreateNew(stage)
	stB.Providers["c"] = "deplC"
	stB.Providers["d"] = "deplD"

	stC := deplC.GetStageOrCreateNew(stage)
	stC.Providers["d"] = "deplD"

	deplD.GetStageOrCreateNew(stage)
	deplE.GetStageOrCreateNew(stage)

	dag, err := env.GetDeploymentStateDAG(stage)
	c.Assert(err, IsNil)
	c.Assert(dag, HasLen, 2)
	c.Assert(dag[0].Node, DeepEquals, deplD)
	c.Assert(dag[1].Node, DeepEquals, deplE)
	c.Assert(dag[1].AndThen, HasLen, 1)

	tsort, err := env.GetDeploymentStateTopologicalSort(stage)
	c.Assert(err, IsNil)
	c.Assert(tsort, HasLen, 5)
	if tsort[0] == deplD {
		c.Assert(tsort[0], DeepEquals, deplD)
		c.Assert(tsort[1], DeepEquals, deplC)
		c.Assert(tsort[2], DeepEquals, deplB)
		c.Assert(tsort[3], DeepEquals, deplE)
		c.Assert(tsort[4], DeepEquals, deplA)
	} else if tsort[0] == deplE {
		c.Assert(tsort[0], DeepEquals, deplE)
		c.Assert(tsort[1], DeepEquals, deplD)
		c.Assert(tsort[2], DeepEquals, deplC)
		c.Assert(tsort[3], DeepEquals, deplB)
		c.Assert(tsort[4], DeepEquals, deplA)
	} else {
		c.Assert(false, Equals, true)
	}
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
