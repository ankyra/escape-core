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

import "fmt"

type DAG struct {
	Node    *DeploymentState
	AndThen []*DAG
}

func (e *EnvironmentState) GetDeploymentStateDAG(stage string) ([]*DAG, error) {
	result := []*DAG{}
	dependsOn := map[*DeploymentState][]*DeploymentState{}
	roots := []*DeploymentState{}

	for _, depl := range e.Deployments {
		stage, found := depl.Stages[stage]
		if !found {
			continue
		}
		depsFound := 0
		for _, deplName := range stage.Providers {
			d, found := e.Deployments[deplName]
			if !found {
				return nil, fmt.Errorf("Referencing unknown provider deployment '%s'", deplName)
			}
			if d == depl {
				continue
			}
			deps, found := dependsOn[d]
			if !found {
				deps = []*DeploymentState{}
			}
			deps = append(deps, depl)
			dependsOn[d] = deps
			depsFound += 1
		}
		if depsFound == 0 {
			roots = append(roots, depl)
		}
	}

	dagMap := map[*DeploymentState]*DAG{}
	seen := map[*DeploymentState]bool{}
	queue := roots
	for len(queue) > 0 {
		q := queue[0]
		queue = queue[1:]

		// mark as seen; only process this node once
		_, alreadySeen := seen[q]
		if alreadySeen {
			continue
		}
		seen[q] = true

		// get the DAG for this Node, or create a new one
		dag, found := dagMap[q]
		if !found {
			dag = &DAG{
				Node:    q,
				AndThen: []*DAG{},
			}
		}

		// add downstream dependencies to DAG
		for _, dep := range dependsOn[q] {
			depDag, found := dagMap[dep]
			if !found {
				depDag = &DAG{
					Node:    dep,
					AndThen: []*DAG{},
				}
			}
			dag.AndThen = append(dag.AndThen, depDag)
			queue = append(queue, dep)
		}
		dagMap[q] = dag
	}
	for _, root := range roots {
		result = append(result, dagMap[root])
	}
	return result, nil
}
