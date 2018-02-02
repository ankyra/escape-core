/*
Copyright 2017, 2018 Ankyra

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
	"github.com/ankyra/escape-core"
	"github.com/ankyra/escape-core/validate"
	. "gopkg.in/check.v1"
)

func (s *suite) Test_NewEnvironmentState(c *C) {
	e, err := NewEnvironmentState("ci", nil)
	c.Assert(err, IsNil)
	c.Assert(e.Name, Equals, "ci")
	c.Assert(e.Inputs, Not(IsNil))
	c.Assert(e.Deployments, Not(IsNil))
	c.Assert(e.Project, IsNil)
}

func (s *suite) Test_Environment_ValidateAndFix_fixes_nils(c *C) {
	e, err := NewEnvironmentState("ci", nil)
	c.Assert(err, IsNil)
	e.Inputs = nil
	e.Deployments = nil
	e.Project = nil
	c.Assert(e.ValidateAndFix("ci", nil), IsNil)
	c.Assert(e.Inputs, Not(IsNil))
	c.Assert(e.Deployments, Not(IsNil))
	c.Assert(e.Project, IsNil)
}

func (s *suite) Test_Environment_ValidateAndFix_fails_on_invalid_name(c *C) {
	cases := []string{
		"",
		".../../",
		"$",
		"@",
		":",
	}
	for _, test := range cases {
		e, err := NewEnvironmentState("ci", nil)
		c.Assert(err, IsNil)
		c.Assert(e.ValidateAndFix(test, nil), DeepEquals, validate.InvalidEnvironmentNameError(test))
	}
}

func (s *suite) Test_Environment_ValidateAndFix_valid_names(c *C) {
	cases := []string{
		"ci",
		"dev",
		"prod",
		"a",
		"a1",
		"a-1",
		"a-_2",
		"a________3",
	}
	for _, test := range cases {
		e, err := NewEnvironmentState("ci", nil)
		c.Assert(err, IsNil)
		c.Assert(e.ValidateAndFix(test, nil), IsNil)
		c.Assert(e.Name, Equals, test)
	}
}

func (s *suite) Test_Project_GetEnvironmentStateOrMakeNew_Env_Name_Field_Is_Set(c *C) {
	p, err := NewProjectStateFromFile("prj", "testdata/project.json", nil)
	c.Assert(err, IsNil)
	env, err := p.GetEnvironmentStateOrMakeNew("incomplete_env")
	c.Assert(err, IsNil)
	c.Assert(env.Name, Equals, "incomplete_env")
}

func (s *suite) Test_Environment_LookupDeploymentState(c *C) {
	p, err := NewProjectStateFromFile("prj", "testdata/project.json", nil)
	c.Assert(err, IsNil)
	env, err := p.GetEnvironmentStateOrMakeNew("dev")
	c.Assert(err, IsNil)
	depl, err := env.LookupDeploymentState("archive-release")
	c.Assert(err, IsNil)
	c.Assert(depl.Name, Equals, "archive-release")
	c.Assert(depl.Inputs["input_variable"], DeepEquals, "depl_override")
	c.Assert(depl.Inputs["list_input"], DeepEquals, []interface{}{"depl_override"})
}

func (s *suite) Test_Environment_LookupDeploymentState_doesnt_exist(c *C) {
	p, err := NewProjectStateFromFile("prj", "testdata/project.json", nil)
	c.Assert(err, IsNil)
	env, err := p.GetEnvironmentStateOrMakeNew("dev")
	c.Assert(err, IsNil)
	_, err = env.LookupDeploymentState("doesnt-exist")
	c.Assert(err, Not(IsNil))
	c.Assert(err.Error(), Equals, "Deployment 'doesnt-exist' does not exist")
}

func (s *suite) Test_Environment_GetOrCreateDeploymentState_no_deps(c *C) {
	p, err := NewProjectStateFromFile("prj", "testdata/project.json", nil)
	c.Assert(err, IsNil)
	env, err := p.GetEnvironmentStateOrMakeNew("dev")
	c.Assert(err, IsNil)
	depl := env.GetOrCreateDeploymentState("archive-release")
	c.Assert(depl.Name, Equals, "archive-release")
	c.Assert(depl.Inputs["input_variable"], DeepEquals, "depl_override")
	c.Assert(depl.Inputs["list_input"], DeepEquals, []interface{}{"depl_override"})
}

func (s *suite) Test_Environment_GetOrCreateDeploymentState_doesnt_exist_no_deps_returns_new(c *C) {
	p, err := NewProjectStateFromFile("prj", "testdata/project.json", nil)
	c.Assert(err, IsNil)
	env, err := p.GetEnvironmentStateOrMakeNew("dev")
	c.Assert(err, IsNil)
	depl := env.GetOrCreateDeploymentState("doesnt-exist")
	c.Assert(depl.Name, Equals, "doesnt-exist")
	c.Assert(depl.Inputs, HasLen, 0)
}

func (s *suite) Test_Environment_GetProviders(c *C) {
	p, err := NewProjectStateFromFile("prj", "testdata/project.json", nil)
	c.Assert(err, IsNil)
	env, err := p.GetEnvironmentStateOrMakeNew("dev")
	c.Assert(err, IsNil)
	depl := env.GetOrCreateDeploymentState("provider")
	metadata := core.NewReleaseMetadata("test", "1")
	metadata.SetProvides([]string{"test-provider"})
	depl.CommitVersion("deploy", metadata)
	providers := env.GetProviders()
	c.Assert(providers, HasLen, 1)
	c.Assert(providers["test-provider"], DeepEquals, []string{"provider"})
}

func (s *suite) Test_Environment_GetProvidersOfType(c *C) {
	p, err := NewProjectStateFromFile("prj", "testdata/project.json", nil)
	c.Assert(err, IsNil)
	env, err := p.GetEnvironmentStateOrMakeNew("dev")
	c.Assert(err, IsNil)
	depl := env.GetOrCreateDeploymentState("provider")
	metadata := core.NewReleaseMetadata("test", "1")
	metadata.SetProvides([]string{"test-provider"})
	depl.CommitVersion("deploy", metadata)
	providers := env.GetProvidersOfType("test-provider")
	c.Assert(providers, HasLen, 1)
	c.Assert(providers, DeepEquals, []string{"provider"})

	providers = env.GetProvidersOfType("no-test-provider")
	c.Assert(providers, HasLen, 0)
}

func (s *suite) Test_Environment_ResolveDeploymentPath(c *C) {
	proj, _ := NewProjectState("project")
	env, err := proj.GetEnvironmentStateOrMakeNew("env")
	c.Assert(err, IsNil)

	_, err = env.ResolveDeploymentPath("deploy", "test")
	c.Assert(err, DeepEquals, DeploymentDoesNotExistError("test"))
	_, err = env.ResolveDeploymentPath("build", "test")
	c.Assert(err, DeepEquals, DeploymentDoesNotExistError("test"))

	depl := env.GetOrCreateDeploymentState("test")
	returnedDepl, err := env.ResolveDeploymentPath("deploy", "test")
	c.Assert(err, IsNil)
	c.Assert(returnedDepl, DeepEquals, depl)

	deplDep := depl.GetDeploymentOrMakeNew("deploy", "test-dependency")
	returnedDepl, err = env.ResolveDeploymentPath("deploy", "test:test-dependency")
	c.Assert(err, IsNil)
	c.Assert(returnedDepl, DeepEquals, deplDep)
	_, err = env.ResolveDeploymentPath("build", "test:test-dependency")
	c.Assert(err, DeepEquals, DeploymentPathResolveError("build", "test:test-dependency", "test-dependency"))

	deplDep2 := deplDep.GetDeploymentOrMakeNew("deploy", "test-dependency2")
	returnedDepl, err = env.ResolveDeploymentPath("deploy", "test:test-dependency:test-dependency2")
	c.Assert(err, IsNil)
	c.Assert(returnedDepl, DeepEquals, deplDep2)
}

func (s *suite) Test_Environment_ResolveDeploymentPath_with_build_stage(c *C) {
	proj, _ := NewProjectState("project")
	env, err := proj.GetEnvironmentStateOrMakeNew("env")
	c.Assert(err, IsNil)

	depl := env.GetOrCreateDeploymentState("test")
	returnedDepl, err := env.ResolveDeploymentPath("build", "test")
	c.Assert(err, IsNil)
	c.Assert(returnedDepl, DeepEquals, depl)

	deplDep := depl.GetDeploymentOrMakeNew("build", "test-dependency")
	returnedDepl, err = env.ResolveDeploymentPath("build", "test:test-dependency")
	c.Assert(err, IsNil)
	c.Assert(returnedDepl, DeepEquals, deplDep)
	_, err = env.ResolveDeploymentPath("deploy", "test:test-dependency")
	c.Assert(err, DeepEquals, DeploymentPathResolveError("deploy", "test:test-dependency", "test-dependency"))
}
