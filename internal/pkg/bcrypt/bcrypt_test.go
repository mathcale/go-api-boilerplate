package bcrypt_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mathcale/go-api-boilerplate/internal/pkg/bcrypt"
)

func TestPassword_HashAndVerify(t *testing.T) {
	pw := bcrypt.NewPassword()

	hash, err := pw.Hash("s3cr3t-password")
	require.NoError(t, err)
	require.NotNil(t, hash)
	assert.NotEqual(t, "s3cr3t-password", *hash)

	require.NoError(t, pw.Verify("s3cr3t-password", *hash))
}

func TestPassword_VerifyRejectsWrongPassword(t *testing.T) {
	pw := bcrypt.NewPassword()

	hash, err := pw.Hash("s3cr3t-password")
	require.NoError(t, err)

	require.Error(t, pw.Verify("wrong-password", *hash))
}
