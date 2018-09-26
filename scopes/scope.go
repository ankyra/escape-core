package scopes

import "fmt"

type Scope string
type Scopes []string

const BuildScope = "build"
const DeployScope = "deploy"

var DeployScopes = Scopes{DeployScope}
var BuildScopes = Scopes{BuildScope}
var AllScopes = Scopes{BuildScope, DeployScope}

func NewScopesFromInterface(val interface{}) ([]string, error) {
	valList, ok := val.([]interface{})
	if !ok {
		return nil, fmt.Errorf("Expecting string in scopes, got '%v' (%T)", val, val)
	}
	scopes := []string{}
	for _, val := range valList {
		kStr, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("Expecting string in scopes, got '%v' (%T)", val, val)
		}
		scopes = append(scopes, kStr)
	}
	return scopes, nil
}

func (s Scopes) Copy() Scopes {
	result := make(Scopes, len(s))
	for i := 0; i < len(s); i++ {
		result[i] = s[i]
	}
	return result
}

func (s Scopes) InScope(scope string) bool {
	for _, sc := range s {
		if sc == scope {
			return true
		}
	}
	return false
}
