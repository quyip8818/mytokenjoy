package grants

type Normalizer interface {
	NormalizeGrantIDs(refs []string) ([]string, error)
	RoleGrantIDs(roleType, roleName string, refs []string) ([]string, error)
}
