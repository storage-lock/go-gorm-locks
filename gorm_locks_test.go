package gorm_locks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobalGormLockFactory_InitiallyNil(t *testing.T) {
	assert.Nil(t, GlobalGormLockFactory)
}
