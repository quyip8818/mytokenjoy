package config

type StoreBootstrap struct {
	// SkipSchema skips DDL apply in postgres.New (caller already applied schema externally).
	SkipSchema bool
	// SkipSeed skips seed.Init in postgres.New (used by tests with pre-seeded template DBs).
	SkipSeed            bool
	TestPartitionMonths int
}
