package ingest

import "github.com/tokenjoy/backend/internal/store"

// groupJobsByCompany groups jobs by resolved company ID.
// Jobs without a known company each become their own group so they can fail fast in parallel.
func groupJobsByCompany(jobs []store.IngestJob, companyByLogID map[int64]int64) [][]store.IngestJob {
	if len(jobs) == 0 {
		return nil
	}
	groups := make(map[int64][]store.IngestJob)
	order := make([]int64, 0)
	orphans := make([]store.IngestJob, 0)
	for _, job := range jobs {
		companyID, ok := companyByLogID[job.LogID]
		if !ok {
			orphans = append(orphans, job)
			continue
		}
		if _, exists := groups[companyID]; !exists {
			order = append(order, companyID)
		}
		groups[companyID] = append(groups[companyID], job)
	}
	out := make([][]store.IngestJob, 0, len(order)+len(orphans))
	for _, companyID := range order {
		out = append(out, groups[companyID])
	}
	for _, job := range orphans {
		out = append(out, []store.IngestJob{job})
	}
	return out
}
