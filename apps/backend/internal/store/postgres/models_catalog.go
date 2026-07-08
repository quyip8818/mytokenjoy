package postgres

const modelSelectColumns = `
	model_id, company_id, provider, type, name, description, endpoint,
	input_price, output_price, max_context, enabled`

type modelCatalog struct {
	db                dbQuerier
	tokenJoyCompanyID int64
}

func newModelCatalog(db dbQuerier, tokenJoyCompanyID int64) modelCatalog {
	return modelCatalog{db: db, tokenJoyCompanyID: tokenJoyCompanyID}
}

func (c modelCatalog) globalCompanyID() int64 {
	return c.tokenJoyCompanyID
}
