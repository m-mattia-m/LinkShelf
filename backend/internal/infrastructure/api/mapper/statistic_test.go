package mapper

import (
	"backend/internal/infrastructure/api/model"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_MapStatisticToStatisticResponse(t *testing.T) {
	statistic := model.Statistic{
		ShelfNumber:   3,
		SectionNumber: 5,
		LinkNumber:    8,
	}

	resp := MapStatisticToStatisticResponse(statistic)

	require.NotNil(t, resp)
	require.Equal(t, 3, resp.Body.ShelfNumber)
	require.Equal(t, 5, resp.Body.SectionNumber)
	require.Equal(t, 8, resp.Body.LinkNumber)
}
