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

package core

import (
	. "gopkg.in/check.v1"
)

func (s *metadataSuite) Test_NewDependencyConfig_Validate_happy_path(c *C) {
	metadata := NewReleaseMetadata("name", "1.0")
	dep := NewDependencyConfig("my-dependency-v1.1")
	dep.BuildMapping = nil
	dep.DeployMapping = nil
	c.Assert(dep.Validate(metadata), IsNil)
	c.Assert(dep.VariableName, Equals, "my-dependency")
	c.Assert(dep.DeploymentName, Equals, "_/my-dependency")
	c.Assert(dep.Project, Equals, "_")
	c.Assert(dep.Name, Equals, "my-dependency")
	c.Assert(dep.Version, Equals, "1.1")
	c.Assert(dep.BuildMapping, Not(IsNil))
	c.Assert(dep.BuildMapping, HasLen, 0)
	c.Assert(dep.DeployMapping, Not(IsNil))
	c.Assert(dep.DeployMapping, HasLen, 0)
	c.Assert(dep.Scopes, DeepEquals, []string{"build", "deploy"})
	c.Assert(dep.Consumes, DeepEquals, map[string]string{})
}

func (s *metadataSuite) Test_NewDependencyConfig_Validate_happy_path_set_variable(c *C) {
	metadata := NewReleaseMetadata("name", "1.0")
	dep := NewDependencyConfig("my-dependency-v1.1 as my-variable")
	dep.BuildMapping = nil
	dep.DeployMapping = nil
	c.Assert(dep.Validate(metadata), IsNil)
	c.Assert(dep.VariableName, Equals, "my-variable")
}

func (s *metadataSuite) Test_NewDependencyConfig_Validate_fails_if_invalid_dependency_string(c *C) {
	cases := []string{
		"",
		"my",
		"my-dependency",
		"my-dependency-v£%%",
	}
	for _, test := range cases {
		metadata := NewReleaseMetadata("name", "1.0")
		dep := NewDependencyConfig(test)
		dep.BuildMapping = nil
		dep.DeployMapping = nil
		c.Assert(dep.Validate(metadata), NotNil)
	}
}

func (s *metadataSuite) Test_NewDependencyConfig_fails_if_version_needs_resolving(c *C) {
	cases := []string{
		"my-dependency-latest",
		"my-dependency-v1.0.@",
		"my-dependency-v0.@",
		"my-dependency-v@",
		"my-dependency-@",
	}
	for _, test := range cases {
		metadata := NewReleaseMetadata("name", "1.0")
		dep := NewDependencyConfig(test)
		dep.BuildMapping = nil
		dep.DeployMapping = nil
		c.Assert(dep.Validate(metadata), DeepEquals, DependencyNeedsResolvingError(test))
	}
}

func (s *metadataSuite) Test_NewDependencyConfigFromMap(c *C) {
	dep, err := NewDependencyConfigFromMap(map[interface{}]interface{}{
		"release_id":      "test-latest",
		"deployment_name": "my-deployment",
		"variable":        "my-variable",
		"build_mapping": map[interface{}]interface{}{
			"build": "building",
		},
		"deploy_mapping": map[interface{}]interface{}{
			"deploy": "deploying",
		},
		"mapping": map[interface{}]interface{}{
			"input_variable1": "test",
		},
		"scopes": []interface{}{"build"},
		"consumes": map[interface{}]interface{}{
			"test": "whatver",
		},
	})
	c.Assert(err, IsNil)
	c.Assert(dep.ReleaseId, Equals, "test-latest")
	c.Assert(dep.VariableName, Equals, "my-variable")
	c.Assert(dep.DeploymentName, Equals, "my-deployment")
	c.Assert(dep.BuildMapping, Not(IsNil))
	c.Assert(dep.BuildMapping, HasLen, 2)
	c.Assert(dep.BuildMapping["input_variable1"], Equals, "test")
	c.Assert(dep.BuildMapping["build"], Equals, "building")
	c.Assert(dep.DeployMapping, Not(IsNil))
	c.Assert(dep.DeployMapping, HasLen, 2)
	c.Assert(dep.DeployMapping["input_variable1"], Equals, "test")
	c.Assert(dep.DeployMapping["deploy"], Equals, "deploying")
	c.Assert(dep.Scopes, DeepEquals, []string{"build"})
	c.Assert(dep.Consumes, DeepEquals, map[string]string{"test": "whatver"})
}

func (s *metadataSuite) Test_NewDependencyConfig_normalises_release_id(c *C) {
	testCases := map[string]string{
		"  prj/test-v1.0  ":              "prj/test-v1.0",
		"test-v1.0  ":                    "_/test-v1.0",
		"prj/test-v0.1   as   var  ":     "prj/test-v0.1",
		"prj/test-v0.1 as var":           "prj/test-v0.1",
		"   prj/test-v0.1   as    var  ": "prj/test-v0.1",
	}
	for testCase, expected := range testCases {
		metadata := NewReleaseMetadata("name", "1.0")
		dep := NewDependencyConfig(testCase)
		c.Assert(dep.Validate(metadata), IsNil)
		c.Assert(dep.ReleaseId, Equals, expected)
	}
}
