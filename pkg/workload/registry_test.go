package workload

import (
	"fmt"
	"testing"
	"time"
)

func TestNewRegistry(t *testing.T) {
	r := NewRegistry(100)
	if r == nil {
		t.Fatal("NewRegistry returned nil")
	}
	if r.Count() != 0 {
		t.Errorf("expected count 0, got %d", r.Count())
	}
}

func TestNewRegistryDefaultSize(t *testing.T) {
	r := NewRegistry(0)
	if r == nil {
		t.Fatal("NewRegistry returned nil")
	}
	if r.maxSize != 1000 {
		t.Errorf("expected default maxSize 1000, got %d", r.maxSize)
	}
}

func TestRecordExec(t *testing.T) {
	r := NewRegistry(100)

	r.RecordExec(12345, "/test/path")

	if r.Count() != 1 {
		t.Errorf("expected count 1, got %d", r.Count())
	}

	m := r.Get(12345)
	if m == nil {
		t.Fatal("expected to find workload 12345")
	}
	if m.ExecCount != 1 {
		t.Errorf("expected ExecCount 1, got %d", m.ExecCount)
	}
	if m.FileCount != 0 {
		t.Errorf("expected FileCount 0, got %d", m.FileCount)
	}
	if m.ConnectCount != 0 {
		t.Errorf("expected ConnectCount 0, got %d", m.ConnectCount)
	}
	if m.CgroupPath != "/test/path" {
		t.Errorf("expected CgroupPath '/test/path', got %q", m.CgroupPath)
	}
}

func TestRecordFile(t *testing.T) {
	r := NewRegistry(100)

	r.RecordFile(12345, "/file/path")

	m := r.Get(12345)
	if m == nil {
		t.Fatal("expected to find workload 12345")
	}
	if m.FileCount != 1 {
		t.Errorf("expected FileCount 1, got %d", m.FileCount)
	}
}

func TestRecordConnect(t *testing.T) {
	r := NewRegistry(100)

	r.RecordConnect(12345, "/connect/path")

	m := r.Get(12345)
	if m == nil {
		t.Fatal("expected to find workload 12345")
	}
	if m.ConnectCount != 1 {
		t.Errorf("expected ConnectCount 1, got %d", m.ConnectCount)
	}
}

func TestRecordAlert(t *testing.T) {
	r := NewRegistry(100)

	// First create the workload
	r.RecordExec(12345, "/alert/path")

	// Then record an alert
	r.RecordAlert(12345)

	m := r.Get(12345)
	if m == nil {
		t.Fatal("expected to find workload 12345")
	}
	if m.AlertCount != 1 {
		t.Errorf("expected AlertCount 1, got %d", m.AlertCount)
	}
}

func TestRecordAlertNonExistent(t *testing.T) {
	r := NewRegistry(100)

	// Recording alert for non-existent workload should not create it
	r.RecordAlert(12345)

	if r.Count() != 0 {
		t.Errorf("expected count 0, got %d", r.Count())
	}
}

func TestMultipleRecords(t *testing.T) {
	r := NewRegistry(100)

	// Record multiple events for the same workload
	r.RecordExec(12345, "/test/path")
	r.RecordExec(12345, "/test/path")
	r.RecordFile(12345, "/test/path")
	r.RecordConnect(12345, "/test/path")
	r.RecordAlert(12345)

	if r.Count() != 1 {
		t.Errorf("expected count 1, got %d", r.Count())
	}

	m := r.Get(12345)
	if m == nil {
		t.Fatal("expected to find workload 12345")
	}
	if m.ExecCount != 2 {
		t.Errorf("expected ExecCount 2, got %d", m.ExecCount)
	}
	if m.FileCount != 1 {
		t.Errorf("expected FileCount 1, got %d", m.FileCount)
	}
	if m.ConnectCount != 1 {
		t.Errorf("expected ConnectCount 1, got %d", m.ConnectCount)
	}
	if m.AlertCount != 1 {
		t.Errorf("expected AlertCount 1, got %d", m.AlertCount)
	}
}

func TestMultipleWorkloads(t *testing.T) {
	r := NewRegistry(100)

	r.RecordExec(111, "/path/1")
	r.RecordExec(222, "/path/2")
	r.RecordExec(333, "/path/3")

	if r.Count() != 3 {
		t.Errorf("expected count 3, got %d", r.Count())
	}
}

func TestList(t *testing.T) {
	r := NewRegistry(100)

	r.RecordExec(111, "/path/1")
	r.RecordExec(222, "/path/2")

	list := r.List()
	if len(list) != 2 {
		t.Errorf("expected list length 2, got %d", len(list))
	}

	// Check that both workloads are in the list
	found := make(map[WorkloadID]bool)
	for _, m := range list {
		found[m.ID] = true
	}
	if !found[111] {
		t.Error("workload 111 not found in list")
	}
	if !found[222] {
		t.Error("workload 222 not found in list")
	}
}

func TestGetNonExistent(t *testing.T) {
	r := NewRegistry(100)

	m := r.Get(99999)
	if m != nil {
		t.Error("expected nil for non-existent workload")
	}
}

func TestGetReturnsCopy(t *testing.T) {
	r := NewRegistry(100)

	r.RecordExec(12345, "/test/path")

	m1 := r.Get(12345)
	m2 := r.Get(12345)

	// Modifying m1 should not affect m2
	m1.ExecCount = 999

	if m2.ExecCount != 1 {
		t.Error("Get should return a copy, not a reference")
	}
}

func TestLRUEviction(t *testing.T) {
	r := NewRegistry(3) // Small max size for testing

	// Add 3 workloads
	r.RecordExec(111, "/path/1")
	time.Sleep(time.Millisecond)
	r.RecordExec(222, "/path/2")
	time.Sleep(time.Millisecond)
	r.RecordExec(333, "/path/3")

	if r.Count() != 3 {
		t.Errorf("expected count 3, got %d", r.Count())
	}

	// Add a 4th workload - should evict the oldest (111)
	r.RecordExec(444, "/path/4")

	if r.Count() != 3 {
		t.Errorf("expected count 3 after eviction, got %d", r.Count())
	}

	// 111 should be evicted
	if r.Get(111) != nil {
		t.Error("workload 111 should have been evicted")
	}

	// Others should still exist
	if r.Get(222) == nil {
		t.Error("workload 222 should still exist")
	}
	if r.Get(333) == nil {
		t.Error("workload 333 should still exist")
	}
	if r.Get(444) == nil {
		t.Error("workload 444 should exist")
	}
}

func TestLRUTouchUpdatesOrder(t *testing.T) {
	r := NewRegistry(3)

	// Add 3 workloads
	r.RecordExec(111, "/path/1")
	time.Sleep(time.Millisecond)
	r.RecordExec(222, "/path/2")
	time.Sleep(time.Millisecond)
	r.RecordExec(333, "/path/3")

	// Touch the oldest (111) to make it recent
	r.RecordExec(111, "/path/1")

	// Add a 4th workload - should now evict 222 (the new oldest)
	r.RecordExec(444, "/path/4")

	// 222 should be evicted
	if r.Get(222) != nil {
		t.Error("workload 222 should have been evicted")
	}

	// 111 should still exist because we touched it
	if r.Get(111) == nil {
		t.Error("workload 111 should still exist after being touched")
	}
}

func TestTimestamps(t *testing.T) {
	r := NewRegistry(100)

	beforeFirst := time.Now()
	r.RecordExec(12345, "/test/path")
	afterFirst := time.Now()

	m := r.Get(12345)
	if m == nil {
		t.Fatal("expected to find workload")
	}

	// FirstSeen should be set within the expected range
	if m.FirstSeen.Before(beforeFirst) || m.FirstSeen.After(afterFirst) {
		t.Error("FirstSeen timestamp is not in expected range")
	}

	// LastSeen should be >= FirstSeen (they may differ by nanoseconds due to the update in Record*)
	if m.LastSeen.Before(m.FirstSeen) {
		t.Error("LastSeen should not be before FirstSeen")
	}

	firstSeenOriginal := m.FirstSeen

	// Wait a bit and record another event
	time.Sleep(10 * time.Millisecond)
	r.RecordExec(12345, "/test/path")

	m = r.Get(12345)

	// FirstSeen should not change
	if !m.FirstSeen.Equal(firstSeenOriginal) {
		t.Error("FirstSeen should not change after subsequent records")
	}

	// LastSeen should be updated (after the original FirstSeen)
	if !m.LastSeen.After(firstSeenOriginal) {
		t.Error("LastSeen should be updated after second record")
	}
}

func TestWorkloadID(t *testing.T) {
	r := NewRegistry(100)

	r.RecordExec(12345, "/test/path")

	m := r.Get(12345)
	if m == nil {
		t.Fatal("expected to find workload")
	}

	if m.ID != 12345 {
		t.Errorf("expected ID 12345, got %d", m.ID)
	}
}

func TestConcurrentAccess(t *testing.T) {
	r := NewRegistry(1000)

	done := make(chan bool)

	// Spawn multiple goroutines recording events
	for i := 0; i < 10; i++ {
		go func(id int) {
			path := fmt.Sprintf("/path/%d", id)
			for j := 0; j < 100; j++ {
				r.RecordExec(uint64(id), path)
				r.RecordFile(uint64(id), path)
				r.RecordConnect(uint64(id), path)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have 10 workloads
	if r.Count() != 10 {
		t.Errorf("expected count 10, got %d", r.Count())
	}

	// Each workload should have 100 of each event type
	for i := 0; i < 10; i++ {
		m := r.Get(uint64(i))
		if m == nil {
			t.Errorf("workload %d not found", i)
			continue
		}
		if m.ExecCount != 100 {
			t.Errorf("workload %d: expected ExecCount 100, got %d", i, m.ExecCount)
		}
		if m.FileCount != 100 {
			t.Errorf("workload %d: expected FileCount 100, got %d", i, m.FileCount)
		}
		if m.ConnectCount != 100 {
			t.Errorf("workload %d: expected ConnectCount 100, got %d", i, m.ConnectCount)
		}
	}
}

func TestCgroupPathUpdate(t *testing.T) {
	r := NewRegistry(100)

	// First record with empty path
	r.RecordExec(12345, "")

	m := r.Get(12345)
	if m.CgroupPath != "" {
		t.Errorf("expected empty path, got %q", m.CgroupPath)
	}

	// Second record with actual path - should update
	r.RecordExec(12345, "/updated/path")

	m = r.Get(12345)
	if m.CgroupPath != "/updated/path" {
		t.Errorf("expected '/updated/path', got %q", m.CgroupPath)
	}

	// Third record with different path - should NOT update (we keep the first non-empty)
	r.RecordExec(12345, "/another/path")

	m = r.Get(12345)
	if m.CgroupPath != "/updated/path" {
		t.Errorf("path should not change once set, got %q", m.CgroupPath)
	}
}
