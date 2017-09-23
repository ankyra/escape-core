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

type DAG []*DAGNode

type DAGNode struct {
	Node    *DeploymentState
	AndThen []*DAGNode
}

func NewDAGNode(d *DeploymentState) *DAGNode {
	return &DAGNode{
		Node:    d,
		AndThen: []*DAGNode{},
	}
}

func (roots DAG) DepthFirstWalk(withFunc func(*DeploymentState)) {
	queue := roots
	seen := map[*DAGNode]bool{}
	for len(queue) > 0 {
		q := queue[0]
		queue = queue[1:]

		// mark as seen; only process this node once
		_, alreadySeen := seen[q]
		if alreadySeen {
			continue
		}
		seen[q] = true
		withFunc(q.Node)

		for _, d := range q.AndThen {
			queue = append(queue, nil)
			copy(queue[1:], queue)
			queue[0] = d
		}
	}
}

func (e *EnvironmentState) GetDeploymentStateTopologicalSort(stage string) ([]*DeploymentState, error) {
	dag, err := e.GetDeploymentStateDAG(stage)
	if err != nil {
		return nil, err
	}
	result := []*DeploymentState{}
	dag.DepthFirstWalk(func(d *DeploymentState) {
		result = append(result, d)
	})
	return result, nil
}

func (e *EnvironmentState) GetDeploymentStateDAG(stage string) (DAG, error) {
	result := DAG{}
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
				return nil, fmt.Errorf("'%s' name is trying to consume itself", deplName)
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

	// Walk the dependency graph
	dagMap := map[*DeploymentState]*DAGNode{}
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
			dag = NewDAGNode(q)
		}

		// add downstream dependencies to DAG
		for _, dep := range dependsOn[q] {
			depDag, found := dagMap[dep]
			if !found {
				depDag = NewDAGNode(dep)
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
