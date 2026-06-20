package testsupport

import (
	"context"
	"testing"
	"time"

	"github.com/maedathiago/eatinpeace/internal/application"
	"github.com/maedathiago/eatinpeace/internal/storage/memory"
)

type Fixture struct {
	Service *application.Service
	Store   *memory.Store
	Now     *time.Time
}

func NewFixture(t *testing.T) Fixture {
	t.Helper()
	now := time.Date(2026, 6, 20, 18, 0, 0, 0, time.UTC)
	store := memory.NewStore()
	service := application.NewService(store)
	service.SetClock(func() time.Time { return now })
	if err := service.SeedPilotFixtures(context.Background()); err != nil {
		t.Fatalf("seed fixtures: %v", err)
	}
	return Fixture{Service: service, Store: store, Now: &now}
}

func (f Fixture) Advance(d time.Duration) {
	*f.Now = f.Now.Add(d)
}
