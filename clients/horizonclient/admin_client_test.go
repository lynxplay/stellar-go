package horizonclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultAdminHostPort(t *testing.T) {
	horizonAdminClient := AdminClient{}

	fullAdminURL, err := horizonAdminClient.getIngestionFiltersURL("test")
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:4200/ingestion/filters/test", fullAdminURL)
}

func TestOverrideAdminHostPort(t *testing.T) {
	horizonAdminClient := AdminClient{
		AdminHost: "127.0.0.1",
		AdminPort: 1234,
	}

	fullAdminURL, err := horizonAdminClient.getIngestionFiltersURL("test")
	require.NoError(t, err)
	assert.Equal(t, "http://127.0.0.1:1234/ingestion/filters/test", fullAdminURL)
}
