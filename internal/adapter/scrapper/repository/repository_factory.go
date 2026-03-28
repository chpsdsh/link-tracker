package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/config"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/repository/builderrepo"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/repository/sqlrepo"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/service"
)

func CreateRepositories(db *pgxpool.Pool, assetType config.AssetType) (service.ChatRepository, service.LinkRepository, error) {
	switch assetType {
	case config.AssetTypeSQL:
		return sqlrepo.NewChatRepository(db), sqlrepo.NewLinkRepository(db), nil
	case config.AssetTypeBuilder:
		return builderrepo.NewChatRepository(db), sqlrepo.NewLinkRepository(db), nil
	default:
		return nil, nil, config.ErrNoAssetType
	}
}
