package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"testing"

	"github.com/go-playground/validator/v10"

	"sistem-manajemen-armada/internal/entity"
	"sistem-manajemen-armada/internal/model"
)

type fakeLocationRepo struct {
	saved map[string]bool
	err   error
}

func (f *fakeLocationRepo) Create(_ context.Context, l *entity.VehicleLocation) (bool, error) {
	if f.err != nil {
		return false, f.err
	}
	key := fmt.Sprintf("%s@%d", l.VehicleID, l.Timestamp)
	if f.saved[key] {
		return false, nil
	}
	f.saved[key] = true
	return true, nil
}
func (f *fakeLocationRepo) FindLatest(context.Context, string) (*entity.VehicleLocation, error) {
	return nil, model.ErrNotFound
}
func (f *fakeLocationRepo) FindHistory(context.Context, string, int64, int64, int, int) ([]entity.VehicleLocation, error) {
	return nil, nil
}

type fakeProducer struct{ sent []model.GeofenceEvent }

func (f *fakeProducer) Send(_ context.Context, e *model.GeofenceEvent) error {
	f.sent = append(f.sent, *e)
	return nil
}

func newTestUseCase() (*VehicleLocationUseCase, *fakeLocationRepo, *fakeProducer) {
	repo := &fakeLocationRepo{saved: map[string]bool{}}
	producer := &fakeProducer{}
	uc := NewVehicleLocationUseCase(slog.New(slog.NewTextHandler(io.Discard, nil)),
		validator.New(validator.WithRequiredStructEnabled()), repo,
		NewGeofenceDetector([]entity.Geofence{defaultFence}), producer)
	return uc, repo, producer
}

func TestIngestPublishesOnceOnCrossingAndIgnoresDuplicates(t *testing.T) {
	uc, _, producer := newTestUseCase()
	msg := &model.LocationMessage{VehicleID: "B1", Latitude: defaultFence.Latitude, Longitude: defaultFence.Longitude, Timestamp: 10}

	for range 2 {
		if err := uc.Ingest(context.Background(), msg); err != nil {
			t.Fatal(err)
		}
	}
	if len(producer.sent) != 1 || producer.sent[0].Event != model.EventGeofenceEntry {
		t.Fatalf("want exactly one geofence_entry, got %+v", producer.sent)
	}
}

func TestIngestRejectsInvalidMessage(t *testing.T) {
	uc, repo, _ := newTestUseCase()
	err := uc.Ingest(context.Background(), &model.LocationMessage{VehicleID: "B1", Latitude: 99, Longitude: 1, Timestamp: 1})
	if !errors.Is(err, model.ErrBadRequest) {
		t.Fatalf("want ErrBadRequest, got %v", err)
	}
	if len(repo.saved) != 0 {
		t.Fatal("invalid message must not be stored")
	}
}

func TestIngestReturnsStorageErrorForRetry(t *testing.T) {
	uc, repo, producer := newTestUseCase()
	repo.err = errors.New("db down")
	err := uc.Ingest(context.Background(), &model.LocationMessage{VehicleID: "B1", Latitude: defaultFence.Latitude, Longitude: defaultFence.Longitude, Timestamp: 1})
	if err == nil || errors.Is(err, model.ErrBadRequest) {
		t.Fatalf("want retryable storage error, got %v", err)
	}
	if len(producer.sent) != 0 {
		t.Fatal("nothing must be published when the location was not stored")
	}
}
