//go:build testhook

package relayfix

import "context"

// StubRelay implements relay.OverrunRelayControl for testing.
type StubRelay struct {
	enabled      bool
	DisabledKeys []string
}

func NewStubRelay() *StubRelay {
	return &StubRelay{enabled: true}
}

func (s *StubRelay) Enabled() bool { return s.enabled }

func (s *StubRelay) DisablePlatformKey(_ context.Context, platformKeyID string) error {
	s.DisabledKeys = append(s.DisabledKeys, platformKeyID)
	return nil
}
