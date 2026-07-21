package postgres

import "github.com/google/uuid"

const modelSelectColumns = `
	model_id, company_id, provider, type, name, description, endpoint,
	api_key, endpoint_model_name,
	max_context, max_tokens, enabled, capabilities`

type modelCatalog struct {
	db                dbQuerier
	tokenJoyCompanyID uuid.UUID
}

func newModelCatalog(db dbQuerier, tokenJoyCompanyID uuid.UUID) modelCatalog {
	return modelCatalog{db: db, tokenJoyCompanyID: tokenJoyCompanyID}
}

func (c modelCatalog) globalCompanyID() uuid.UUID {
	return c.tokenJoyCompanyID
}
