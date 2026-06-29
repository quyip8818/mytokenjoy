package org

import (
	"encoding/json"
	"net/http"

	"github.com/tokenjoy/backend/internal/domain"
	domainorg "github.com/tokenjoy/backend/internal/domain/org"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/http/httputil"
)

func decodeCredential(r *http.Request) (domainorg.Credential, error) {
	cred, err := types.DecodeCredential(json.NewDecoder(r.Body))
	if err != nil {
		return domainorg.Credential{}, domain.BadRequest(httputil.MsgBadBody)
	}
	return cred, nil
}
