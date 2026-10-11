package gorm_locks

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func testGormDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:gorm_test.db?mode=memory"), &gorm.Config{})
	assert.Nil(t, err)
	return db
}

func TestNewGormConnectionManager(t *testing.T) {
	db := testGormDB(t)
	m := NewGormConnectionManager(db)
	assert.NotNil(t, m)
	assert.Equal(t, "gorm-connection-manager", m.Name())

	ctx := context.Background()
	got, err := m.Take(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, got)

	assert.Nil(t, m.Return(ctx, got))
	assert.Nil(t, m.Shutdown(ctx))
}

func TestNewGormLockFactory(t *testing.T) {
	db := testGormDB(t)
	factory, err := NewGormLockFactory(db)
	assert.Nil(t, err)
	assert.NotNil(t, factory)

	// 端到端加解锁
	ctx := context.Background()
	lock, err := factory.CreateLock("test-gorm-lock")
	assert.Nil(t, err)
	assert.NotNil(t, lock)
	assert.Nil(t, lock.Lock(ctx, "owner-1"))
	assert.Nil(t, lock.UnLock(ctx, "owner-1"))
}

func TestCreateStorageForGormDb(t *testing.T) {
	db := testGormDB(t)
	m := NewGormConnectionManager(db)
	s, err := CreateStorageForGormDb(db, m)
	assert.Nil(t, err)
	assert.NotNil(t, s)
}

func TestInitGormLockFactory(t *testing.T) {
	db := testGormDB(t)
	assert.Nil(t, InitGormLockFactory(db))
	assert.NotNil(t, GlobalGormLockFactory)
	// 还原全局，避免污染其它测试
	GlobalGormLockFactory = nil
}
