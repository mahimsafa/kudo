package state

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/hashicorp/raft"
)

func TestFSMApplySetApplication(t *testing.T) {
	fsm := NewFSM()

	app := Application{
		Name:     "test-app",
		Adapter:  "docker",
		Replicas: 3,
	}

	cmd := Command{
		Op:   OpSetApplication,
		Data: mustMarshal(t, app),
	}

	logEntry := &raft.Log{
		Data: mustMarshal(t, cmd),
	}

	result := fsm.Apply(logEntry)
	if result != nil {
		t.Fatalf("unexpected error: %v", result)
	}

	got, exists := fsm.GetApplication("test-app")
	if !exists {
		t.Fatal("expected application to exist")
	}
	if got.Replicas != 3 {
		t.Errorf("expected 3 replicas, got %d", got.Replicas)
	}
}

func TestFSMApplyDeleteApplication(t *testing.T) {
	fsm := NewFSM()

	app := Application{Name: "to-delete", Adapter: "docker", Replicas: 1}
	cmd := Command{Op: OpSetApplication, Data: mustMarshal(t, app)}
	fsm.Apply(&raft.Log{Data: mustMarshal(t, cmd)})

	instCmd := Command{Op: OpSetInstance, Data: mustMarshal(t, Instance{
		ID: "inst-1", AppName: "to-delete", NodeID: "node-1", Status: "running",
	})}
	fsm.Apply(&raft.Log{Data: mustMarshal(t, instCmd)})

	delCmd := Command{Op: OpDeleteApplication, Data: mustMarshal(t, "to-delete")}
	fsm.Apply(&raft.Log{Data: mustMarshal(t, delCmd)})

	_, exists := fsm.GetApplication("to-delete")
	if exists {
		t.Error("expected application to be deleted")
	}
	if len(fsm.GetInstancesForApp("to-delete")) != 0 {
		t.Error("expected instances to be deleted with application")
	}
}

func TestFSMSnapshot(t *testing.T) {
	fsm := NewFSM()

	app := Application{Name: "snap-app", Adapter: "docker", Replicas: 2}
	cmd := Command{Op: OpSetApplication, Data: mustMarshal(t, app)}
	fsm.Apply(&raft.Log{Data: mustMarshal(t, cmd)})

	snapshot, err := fsm.Snapshot()
	if err != nil {
		t.Fatalf("snapshot error: %v", err)
	}

	newFSM := NewFSM()
	sink := &mockSnapshotSink{}
	if err := snapshot.Persist(sink); err != nil {
		t.Fatalf("persist error: %v", err)
	}

	if err := newFSM.Restore(sink.Reader()); err != nil {
		t.Fatalf("restore error: %v", err)
	}

	got, exists := newFSM.GetApplication("snap-app")
	if !exists {
		t.Fatal("expected application after restore")
	}
	if got.Replicas != 2 {
		t.Errorf("expected 2 replicas, got %d", got.Replicas)
	}
}

func mustMarshal(t *testing.T, v interface{}) []byte {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	return data
}

type mockSnapshotSink struct {
	data []byte
}

func (s *mockSnapshotSink) Write(p []byte) (int, error) {
	s.data = append(s.data, p...)
	return len(p), nil
}

func (s *mockSnapshotSink) Close() error  { return nil }
func (s *mockSnapshotSink) ID() string    { return "mock" }
func (s *mockSnapshotSink) Cancel() error { return nil }
func (s *mockSnapshotSink) Reader() io.ReadCloser {
	return io.NopCloser(bytes.NewReader(s.data))
}
