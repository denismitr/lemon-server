package database

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_createPath(t *testing.T) {
	validNames := []struct {
		in  string
		exp string
	}{
		{in: "foo", exp: "./data/foo.ldb"},
		{in: "foo-ldb", exp: "./data/foo-ldb.ldb"},
		{in: "foo_ldb", exp: "./data/foo_ldb.ldb"},
	}

	for i, tc := range validNames {
		t.Run(fmt.Sprintf("Valid DB names test case: %d", i), func(t *testing.T) {
			exp, err := createFullDBPath(tc.in)
			require.NoErrorf(t, err, "should be no error")
			assert.Equal(t, tc.exp, exp)
		})
	}

	invalidNames := []struct {
		in string
	}{
		{in: "foo/fpp"},
		{in: "./data/foo.bar.ldb"},
		{in: "foo.ldb"},
		{in: "./data/foo.ldb"},
	}

	for i, tc := range invalidNames {
		t.Run(fmt.Sprintf("Invalid DB names test case: %d", i), func(t *testing.T) {
			exp, err := createFullDBPath(tc.in)
			assert.Error(t, err)
			assert.Truef(t, errors.Is(err, ErrInvalidDatabaseName), "should be ErrInvalidDatabaseName")
			assert.Equal(t, "", exp)
		})
	}
}
