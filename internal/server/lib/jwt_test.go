package lib

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SmoothWay/gophkeeper/pkg/jwt"
	"github.com/SmoothWay/gophkeeper/pkg/models"
)

func TestParseToken(t *testing.T) {
	user := models.User{ID: 10, Email: "name@example.com", PassHash: []byte("hash")}
	app := models.App{ID: 1, Name: "gophkeeper", Secret: "test-secret"}
	token, err := jwt.NewToken(user, app, time.Hour)
	require.NoError(t, err)

	userID, err := ParseToken(token)
	require.NoError(t, err)
	assert.Equal(t, user.ID, userID)
}
