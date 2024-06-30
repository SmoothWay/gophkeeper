package encrypt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeMsg(t *testing.T) {
	sourceMsg := "some text"
	key := "testpassword"
	encodeHash := EncodeMsg([]byte(sourceMsg), key)
	require.NotEmpty(t, sourceMsg)
	resultMsg := DecodeMsg(encodeHash, key)
	require.NotEmpty(t, resultMsg)
	require.Equal(t, sourceMsg, resultMsg)
}
