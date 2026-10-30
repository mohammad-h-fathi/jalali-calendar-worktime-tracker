package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"persian_cal/internal/models"
)

func TestStoreCorruptRecovery(t *testing.T) {
	tempDir := t.TempDir()
	storeFile := filepath.Join(tempDir, "data.json")

	// Write corrupt JSON
	if err := os.WriteFile(storeFile, []byte("{corrupted json contents"), 0644); err != nil {
		t.Fatalf("Failed to write corrupt file: %v", err)
	}

	// Loading store should not fail, should quarantine and initialize defaults
	store, err := NewStore(storeFile)
	if err != nil {
		t.Fatalf("Expected NewStore to succeed despite corrupt file, got: %v", err)
	}

	// Verify defaults
	projects := store.GetProjects()
	if len(projects) == 0 {
		t.Fatalf("Expected default projects, got 0")
	}

	// Verify corrupt file was moved aside
	matches, _ := filepath.Glob(filepath.Join(tempDir, "data.json.corrupt-*"))
	if len(matches) == 0 {
		t.Errorf("Expected data.json.corrupt-* file to exist")
	}
}

func TestStoreTaskPrioritySorting(t *testing.T) {
	tempDir := t.TempDir()
	storeFile := filepath.Join(tempDir, "data.json")

	store, err := NewStore(storeFile)
	if err != nil {
		t.Fatalf("Failed to init store: %v", err)
	}

	now := time.Now()
	_ = store.SaveTask(models.Task{Title: "Low Task", Priority: models.PriorityLow, CreatedAt: now})
	_ = store.SaveTask(models.Task{Title: "Urgent Task", Priority: models.PriorityUrgent, CreatedAt: now.Add(1 * time.Second)})
	_ = store.SaveTask(models.Task{Title: "Medium Task", Priority: models.PriorityMedium, CreatedAt: now.Add(2 * time.Second)})
	_ = store.SaveTask(models.Task{Title: "High Task", Priority: models.PriorityHigh, CreatedAt: now.Add(3 * time.Second)})

	tasks := store.GetTasks("", nil)
	if len(tasks) != 4 {
		t.Fatalf("Expected 4 tasks, got %d", len(tasks))
	}

	// Expected order: Urgent, High, Medium, Low
	if tasks[0].Priority != models.PriorityUrgent {
		t.Errorf("Expected tasks[0] to be urgent, got %s", tasks[0].Priority)
	}
	if tasks[1].Priority != models.PriorityHigh {
		t.Errorf("Expected tasks[1] to be high, got %s", tasks[1].Priority)
	}
	if tasks[2].Priority != models.PriorityMedium {
		t.Errorf("Expected tasks[2] to be medium, got %s", tasks[2].Priority)
	}
	if tasks[3].Priority != models.PriorityLow {
		t.Errorf("Expected tasks[3] to be low, got %s", tasks[3].Priority)
	}
}
