package postgres

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/furadx/iam-go/internal/apiserver/model"
)

func TestDeleteUnscopedDoesNotMutateStoreDB(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("migrate users: %v", err)
	}

	repo := &users{db: db}
	if err := repo.Delete(context.Background(), "missing", model.DeleteOptions{Unscoped: true}); err != nil {
		t.Fatalf("delete missing user: %v", err)
	}

	if repo.db.Statement != nil && repo.db.Statement.Unscoped {
		t.Fatal("expected unscoped delete to use a local DB session")
	}
}

func TestDeleteSoftDeletesByDefault(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("migrate users: %v", err)
	}
	if err := db.Create(&model.User{
		Name:     "alice",
		Password: "x",
		Email:    "alice@example.com",
		Status:   1,
	}).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	repo := &users{db: db}
	if err := repo.Delete(context.Background(), "alice", model.DeleteOptions{}); err != nil {
		t.Fatalf("soft delete user: %v", err)
	}
	if _, err := repo.Get(context.Background(), "alice", model.GetOptions{}); err == nil {
		t.Fatal("expected soft-deleted user to be hidden from scoped get")
	}

	var count int64
	if err := db.Unscoped().Model(&model.User{}).Where("name = ?", "alice").Count(&count).Error; err != nil {
		t.Fatalf("count unscoped user: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected soft-deleted row to remain, got %d", count)
	}
}

func TestDeleteUnscopedHardDeletes(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("migrate users: %v", err)
	}
	if err := db.Create(&model.User{
		Name:     "alice",
		Password: "x",
		Email:    "alice@example.com",
		Status:   1,
	}).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	repo := &users{db: db}
	if err := repo.Delete(context.Background(), "alice", model.DeleteOptions{Unscoped: true}); err != nil {
		t.Fatalf("hard delete user: %v", err)
	}

	var count int64
	if err := db.Unscoped().Model(&model.User{}).Where("name = ?", "alice").Count(&count).Error; err != nil {
		t.Fatalf("count unscoped user: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected hard-deleted row to be removed, got %d", count)
	}
}
