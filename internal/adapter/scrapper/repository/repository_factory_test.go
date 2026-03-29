package repository

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/config"
)

func TestCreateRepositories(t *testing.T) {
	var db *pgxpool.Pool

	tests := []struct {
		name      string
		assetType config.AssetType
		wantErr   error
	}{
		{
			name:      "sql repositories",
			assetType: config.AssetTypeSQL,
			wantErr:   nil,
		},
		{
			name:      "builder repositories",
			assetType: config.AssetTypeBuilder,
			wantErr:   nil,
		},
		{
			name:      "unknown asset type",
			assetType: 23,
			wantErr:   config.ErrNoAssetType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := CreateRepositories(db, tt.assetType)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
