package permission

import "github.com/tokenjoy/backend/internal/domain/grants"

type grantNormalizer struct{}

func NewGrantNormalizer() grants.Normalizer {
	return grantNormalizer{}
}

func (grantNormalizer) NormalizeGrantIDs(refs []string) ([]string, error) {
	return NormalizeGrantIDs(refs)
}

func (grantNormalizer) RoleGrantIDs(roleType, roleName string, refs []string) ([]string, error) {
	return RoleGrantIDs(roleType, roleName, refs)
}
