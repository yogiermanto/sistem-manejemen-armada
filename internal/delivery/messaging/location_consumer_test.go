package messaging

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/go-playground/validator/v10"

	"sistem-manajemen-armada/internal/entity"
	"sistem-manajemen-armada/internal/usecase"
)

type flakyRepo struct {
	fails  int
	stored int
}

func (r *flakyRepo) Create(context.Context, *entity.VehicleLocation) (bool, error) {
	if r.fails > 0 {
		r.fails--
		return false, errors.New("db down")
	}
	r.stored++
	return true, nil
}
func (r *flakyRepo) FindLatest(context.Context, string) (*entity.VehicleLocation, error) {
	return nil, nil
}
func (r *flakyRepo) FindHistory(context.Context, string, int64, int64, int, int) ([]entity.VehicleLocation, error) {
	return nil, nil
}

type fakeMessage struct {
	topic   string
	payload string
	acked   bool
}

func (m *fakeMessage) Duplicate() bool   { return false }
func (m *fakeMessage) Qos() byte         { return 1 }
func (m *fakeMessage) Retained() bool    { return false }
func (m *fakeMessage) Topic() string     { return m.topic }
func (m *fakeMessage) MessageID() uint16 { return 1 }
func (m *fakeMessage) Payload() []byte   { return []byte(m.payload) }
func (m *fakeMessage) Ack()              { m.acked = true }

func newTestConsumer(ctx context.Context, repo *flakyRepo) *LocationConsumer {
	log := slog.New(slog.DiscardHandler)
	uc := usecase.NewVehicleLocationUseCase(log, validator.New(validator.WithRequiredStructEnabled()), repo, usecase.NewGeofenceDetector(nil), nil)
	return NewLocationConsumer(ctx, log, uc)
}

const validPing = `{"vehicle_id":"B1","latitude":-6.2,"longitude":106.8,"timestamp":1700000000}`

func TestHandleAcksOnlyAfterStore(t *testing.T) {
	cases := []struct {
		name      string
		topic     string
		payload   string
		fails     int
		wantStore int
	}{
		{"valid", "/fleet/vehicle/B1/location", validPing, 0, 1},
		{"db down once, retried", "/fleet/vehicle/B1/location", validPing, 1, 1},
		{"invalid json dropped", "/fleet/vehicle/B1/location", `not json`, 0, 0},
		{"topic mismatch dropped", "/fleet/vehicle/OTHER/location", validPing, 0, 0},
		{"out of range dropped", "/fleet/vehicle/B1/location", `{"vehicle_id":"B1","latitude":99,"longitude":106.8,"timestamp":1700000000}`, 0, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &flakyRepo{fails: tc.fails}
			msg := &fakeMessage{topic: tc.topic, payload: tc.payload}
			newTestConsumer(context.Background(), repo).Handle(nil, msg)
			if repo.stored != tc.wantStore || !msg.acked {
				t.Fatalf("stored=%d acked=%v, want stored=%d acked=true", repo.stored, msg.acked, tc.wantStore)
			}
		})
	}
}

func TestHandleShutdownDuringRetryDoesNotAck(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	repo := &flakyRepo{fails: 1000}
	msg := &fakeMessage{topic: "/fleet/vehicle/B1/location", payload: validPing}
	newTestConsumer(ctx, repo).Handle(nil, msg)
	if msg.acked || repo.stored != 0 {
		t.Fatalf("acked=%v stored=%d, want unacked so the broker redelivers", msg.acked, repo.stored)
	}
}
