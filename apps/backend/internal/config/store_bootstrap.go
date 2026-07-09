package config

type StoreBootstrap struct {
	SkipRuntimeSeed     bool
	SchemaPrepared      bool
	TestPartitionMonths int
	RuntimeSeed         bool
}
