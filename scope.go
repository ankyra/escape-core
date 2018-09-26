package core

type Scope string
type Scopes []string

const BuildScope = "build"
const DeployScope = "deploy"

var DeployScopes = Scopes{DeployScope}
var BuildScopes = Scopes{BuildScope}
var AllScopes = Scopes{BuildScope, DeployScope}

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
