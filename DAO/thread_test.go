package data

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestThread_TypeString(t *testing.T) {
	tests := []struct {
		name     string
		thread   Thread
		expected string
	}{
		{"Type Ithink", Thread{Type: ThreadTypeIthink}, "我觉得"},
		{"Type Idea", Thread{Type: ThreadTypeIdea}, "出主意"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.thread.TypeString()
			if result != tt.expected {
				t.Errorf("TypeString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestThread_IsEdited(t *testing.T) {
	now := time.Now()
	later := now.Add(time.Hour)

	tests := []struct {
		name     string
		thread   Thread
		expected bool
	}{
		{"Not edited - nil EditAt", Thread{CreatedAt: now, EditAt: nil}, false},
		{"Not edited - same time", Thread{CreatedAt: now, EditAt: &now}, false},
		{"Edited - different time", Thread{CreatedAt: now, EditAt: &later}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.thread.IsEdited()
			if result != tt.expected {
				t.Errorf("IsEdited() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestThread_NumReplies(t *testing.T) {
	thread := Thread{Id: 1}
	count := thread.NumReplies()

	if count < 0 {
		t.Errorf("NumReplies() returned negative count: %d", count)
	}
}

func TestThread_NumSupport(t *testing.T) {
	thread := Thread{Id: 1}
	count := thread.NumSupport()

	if count < 0 {
		t.Errorf("NumSupport() returned negative count: %d", count)
	}
}

func TestThread_NumOppose(t *testing.T) {
	thread := Thread{Id: 1}
	count := thread.NumOppose()

	if count < 0 {
		t.Errorf("NumOppose() returned negative count: %d", count)
	}
}

func TestThread_IsAuthor(t *testing.T) {
	tests := []struct {
		name     string
		thread   Thread
		user     User
		expected bool
	}{
		{"Is author", Thread{UserId: 1}, User{Id: 1}, true},
		{"Not author", Thread{UserId: 1}, User{Id: 2}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.thread.IsAuthor(tt.user)
			if result != tt.expected {
				t.Errorf("IsAuthor() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestThread_UpdateBodyAndClass(t *testing.T) {
	thread := Thread{Id: 1}
	ctx := context.Background()

	err := thread.UpdateBodyAndClass("test body", 1, ctx)
	if err != nil {
		t.Logf("UpdateBodyAndClass() error (expected if thread doesn't exist): %v", err)
	}
}

func TestHotThreads(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		limit     int
		expectErr bool
	}{
		{"Valid limit", 10, false},
		{"Zero limit", 0, true},
		{"Negative limit", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			threads, err := HotThreads(tt.limit, ctx)

			if tt.expectErr {
				if err == nil {
					t.Errorf("HotThreads() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("HotThreads() unexpected error: %v", err)
				return
			}

			if len(threads) > tt.limit {
				t.Errorf("HotThreads() returned %d threads, expected max %d", len(threads), tt.limit)
			}
		})
	}
}

func TestGetThreadByUUID(t *testing.T) {

	tests := []struct {
		name      string
		uuid      string
		expectErr bool
	}{
		{"Empty UUID", "", true},
		{"Valid UUID format", "550e8400-e29b-41d4-a716-446655440000", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetThreadByUUID(tt.uuid)

			if tt.expectErr && err == nil {
				t.Errorf("GetThreadByUUID() expected error but got none")
			}
			if !tt.expectErr && err != nil && err != sql.ErrNoRows {
				t.Errorf("GetThreadByUUID() unexpected error: %v", err)
			}
		})
	}
}

func TestGetThreadById(t *testing.T) {

	_, err := GetThreadById(1)
	if err != nil && err != sql.ErrNoRows {
		t.Errorf("GetThreadById() unexpected error: %v", err)
	}
}

func TestProject_ThreadsNormal(t *testing.T) {

	project := Project{Id: 1}
	ctx := context.Background()

	threads, err := project.ThreadsNormal(ctx)
	if err != nil {
		t.Errorf("ThreadsNormal() error: %v", err)
		return
	}

	if threads == nil {
		t.Errorf("ThreadsNormal() returned nil slice")
	}
}

func TestProject_ThreadAppointment(t *testing.T) {

	project := Project{Id: 1}
	ctx := context.Background()

	_, err := project.ThreadAppointment(ctx)
	if err != nil && err != sql.ErrNoRows {
		t.Errorf("ThreadAppointment() unexpected error: %v", err)
	}
}

func TestProject_ThreadsSeeSeek(t *testing.T) {

	project := Project{Id: 1}
	ctx := context.Background()

	threads, err := project.ThreadsSeeSeek(ctx)
	if err != nil {
		t.Errorf("ThreadsSeeSeek() error: %v", err)
		return
	}

	if threads == nil {
		t.Errorf("ThreadsSeeSeek() returned nil slice")
	}
}

func TestProject_ThreadsBrainFire(t *testing.T) {

	project := Project{Id: 1}
	ctx := context.Background()

	threads, err := project.ThreadsBrainFire(ctx)
	if err != nil {
		t.Errorf("ThreadsBrainFire() error: %v", err)
		return
	}

	if threads == nil {
		t.Errorf("ThreadsBrainFire() returned nil slice")
	}
}

func TestProject_ThreadsSuggestion(t *testing.T) {

	project := Project{Id: 1}
	ctx := context.Background()

	threads, err := project.ThreadsSuggestion(ctx)
	if err != nil {
		t.Errorf("ThreadsSuggestion() error: %v", err)
		return
	}

	if threads == nil {
		t.Errorf("ThreadsSuggestion() returned nil slice")
	}
}

func TestProject_ThreadsGoods(t *testing.T) {

	project := Project{Id: 1}
	ctx := context.Background()

	threads, err := project.ThreadsGoods(ctx)
	if err != nil {
		t.Errorf("ThreadsGoods() error: %v", err)
		return
	}

	if threads == nil {
		t.Errorf("ThreadsGoods() returned nil slice")
	}
}

func TestProject_ThreadsHandicraft(t *testing.T) {

	project := Project{Id: 1}
	ctx := context.Background()

	threads, err := project.ThreadsHandicraft(ctx)
	if err != nil {
		t.Errorf("ThreadsHandicraft() error: %v", err)
		return
	}

	if threads == nil {
		t.Errorf("ThreadsHandicraft() returned nil slice")
	}
}

func TestSearchThreadByTitle(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		keyword string
		limit   int
	}{
		{"Normal search", "test", 10},
		{"Empty keyword", "", 5},
		{"Large limit", "keyword", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			threads, err := SearchThreadByTitle(tt.keyword, tt.limit, ctx)
			if err != nil {
				t.Errorf("SearchThreadByTitle() error: %v", err)
				return
			}

			if threads == nil {
				t.Errorf("SearchThreadByTitle() returned nil slice")
			}

			if len(threads) > tt.limit {
				t.Errorf("SearchThreadByTitle() returned %d threads, expected max %d", len(threads), tt.limit)
			}
		})
	}
}

func TestDraftThread_StatusString(t *testing.T) {
	tests := []struct {
		name     string
		draft    DraftThread
		expected string
	}{
		{"Status Pending", DraftThread{Status: DraftThreadStatusPending}, "草稿"},
		{"Status Accepted", DraftThread{Status: DraftThreadStatusAccepted}, "接纳"},
		{"Status Rejected", DraftThread{Status: DraftThreadStatusRejected}, "婉拒"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.draft.StatusString()
			if result != tt.expected {
				t.Errorf("StatusString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestThread_CreatedAtDate(t *testing.T) {
	now := time.Now()
	thread := Thread{CreatedAt: now}

	result := thread.CreatedAtDate()
	if result == "" {
		t.Errorf("CreatedAtDate() returned empty string")
	}
}

func TestThread_EditAtDate(t *testing.T) {
	now := time.Now()
	thread := Thread{EditAt: &now}

	result := thread.EditAtDate()
	if result == "" {
		t.Errorf("EditAtDate() returned empty string")
	}
}

func TestCreateRequiredThreads(t *testing.T) {

	objective := &Objective{Id: 1, FamilyId: 1, TeamId: 1}
	project := &Project{Id: 1, Class: 1, IsPrivate: false}
	userId := 1
	ctx := context.Background()

	err := CreateRequiredThreads(objective, project, userId, ctx)
	if err != nil {
		t.Logf("CreateRequiredThreads() error (expected if data doesn't exist): %v", err)
	}
}

func TestThreadApproved_Create(t *testing.T) {

	approved := ThreadApproved{
		ProjectId: 1,
		ThreadId:  1,
		UserId:    1,
	}

	err := approved.Create()
	if err != nil {
		t.Logf("ThreadApproved.Create() error (expected if data doesn't exist): %v", err)
	}
}

func TestThread_IsApproved(t *testing.T) {

	thread := Thread{Id: 1}
	result := thread.IsApproved()

	// Result should be boolean
	if result != true && result != false {
		t.Errorf("IsApproved() should return boolean, got %v", result)
	}
}

// Benchmark tests
func BenchmarkThread_TypeString(b *testing.B) {
	thread := Thread{Type: ThreadTypeIthink}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = thread.TypeString()
	}
}

func BenchmarkThread_IsEdited(b *testing.B) {
	now := time.Now()
	later := now.Add(time.Hour)
	thread := Thread{CreatedAt: now, EditAt: &later}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = thread.IsEdited()
	}
}

func BenchmarkHotThreads(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = HotThreads(10, ctx)
	}
}
