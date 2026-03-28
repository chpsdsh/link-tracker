package builderrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/database"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/scrapper"
)

type LinkRepository struct {
	db      *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

func NewLinkRepository(db *pgxpool.Pool) *LinkRepository {
	b := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	return &LinkRepository{db: db, builder: b}
}

func (r *LinkRepository) AddLink(ctx context.Context, chatID int64, link pkg.LinkInfo) error {
	q := database.GetQuerier(ctx, r.db)

	var linkID int64

	query, args, err := r.builder.
		Insert("links").
		Columns("url").
		Values(link.Link).
		Suffix(`
			on conflict (url) do update set url = excluded.url
			returning id
		`).
		ToSql()

	if err != nil {
		return fmt.Errorf("build insert link: %w", err)
	}

	err = q.QueryRow(ctx, query, args...).Scan(&linkID)
	if err != nil {
		return fmt.Errorf("insert link: %w", err)
	}

	query, args, err = r.builder.
		Insert("link_chat").
		Columns("chat_id", "link_id").
		Values(chatID, linkID).
		Suffix("on conflict do nothing").
		ToSql()

	if err != nil {
		return fmt.Errorf("build insert link_chat: %w", err)
	}

	commandTag, err := q.Exec(ctx, query, args...)
	if err != nil {
		if database.IsForeignKeyViolation(err) {
			return scrapper.ErrChatNotFound
		}
		return fmt.Errorf("insert link_chat: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return scrapper.ErrLinkExists
	}

	for _, tag := range link.Tags {
		var tagID int64

		query, args, err = r.builder.
			Insert("tags").
			Columns("tag").
			Values(tag).
			Suffix(`
				on conflict (tag) do update set tag = excluded.tag
				returning id
			`).
			ToSql()

		if err != nil {
			return fmt.Errorf("build insert tag: %w", err)
		}

		err = q.QueryRow(ctx, query, args...).Scan(&tagID)
		if err != nil {
			return fmt.Errorf("insert tag: %w", err)
		}

		query, args, err = r.builder.
			Insert("link_tag").
			Columns("link_id", "tag_id").
			Values(linkID, tagID).
			Suffix("on conflict do nothing").
			ToSql()

		if err != nil {
			return fmt.Errorf("build insert link_tag: %w", err)
		}

		_, err = q.Exec(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("insert link_tag: %w", err)
		}
	}

	return nil
}

func (r *LinkRepository) LinkExists(ctx context.Context, url string) (bool, error) {
	q := database.GetQuerier(ctx, r.db)

	var exists bool

	subquery := r.builder.
		Select("1").
		From("links").
		Where(squirrel.Eq{"url": url})

	query, args, err := r.builder.
		Select().
		Column(squirrel.Expr("exists (?)", subquery)).
		ToSql()

	if err != nil {
		return false, fmt.Errorf("build link exists query: %w", err)
	}

	err = q.QueryRow(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("query link exists: %w", err)
	}

	return exists, nil
}

func (r *LinkRepository) DeleteLink(ctx context.Context, chatID int64, url string) (pkg.LinkInfo, error) {
	q := database.GetQuerier(ctx, r.db)

	var li pkg.LinkInfo
	var tags []string

	query := `
		with deleted as (
			delete from link_chat lc
			using links l
			where lc.link_id = l.id
			  and lc.chat_id = $1
			  and l.url = $2
			returning l.id, l.url, l.updated_at
		)
		select 
			d.url,
			d.updated_at,
			coalesce(array_agg(t.tag), '{}')
		from deleted d
		left join link_tag lt on lt.link_id = d.id
		left join tags t on t.id = lt.tag_id
		group by d.id, d.url, d.updated_at
	`

	err := q.QueryRow(ctx, query, chatID, url).
		Scan(&li.Link, &li.LastUpdateTime, &tags)

	if err != nil {
		return pkg.LinkInfo{}, err
	}

	li.Tags = tags

	builder := r.builder

	subQuery, subArgs, err := builder.
		Select("id").
		From("links").
		Where(squirrel.Eq{"url": url}).
		ToSql()

	if err != nil {
		return pkg.LinkInfo{}, fmt.Errorf("build subquery: %w", err)
	}

	notExistsQuery, notExistsArgs, err := builder.
		Select("1").
		From("link_chat lc").
		Join("links l on l.id = lc.link_id").
		Where(squirrel.Eq{"l.url": url}).
		ToSql()

	if err != nil {
		return pkg.LinkInfo{}, fmt.Errorf("build exists: %w", err)
	}

	query, args, err := builder.
		Delete("links").
		Where(squirrel.Expr("id = ("+subQuery+")", subArgs...)).
		Where(squirrel.Expr("not exists ("+notExistsQuery+")", notExistsArgs...)).
		ToSql()

	if err != nil {
		return pkg.LinkInfo{}, fmt.Errorf("build delete links: %w", err)
	}

	_, err = q.Exec(ctx, query, args...)
	if err != nil {
		return pkg.LinkInfo{}, fmt.Errorf("delete links: %w", err)
	}

	return li, nil
}

func (r *LinkRepository) GetUserLinks(ctx context.Context, chatID int64) ([]pkg.LinkInfo, error) {
	q := database.GetQuerier(ctx, r.db)

	query, args, err := r.builder.
		Select("l.url", "l.updated_at").
		Column(squirrel.Expr("coalesce(array_agg(t.tag) filter (where t.tag is not null), '{}')")).
		From("link_chat lc").
		Join("links l on l.id = lc.link_id").
		LeftJoin("link_tag lt on lt.link_id = l.id").
		LeftJoin("tags t on t.id = lt.tag_id").
		Where(squirrel.Eq{"lc.chat_id": chatID}).
		GroupBy("l.id", "l.url", "l.updated_at").
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build get user links: %w", err)
	}

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query get user links: %w", err)
	}
	defer rows.Close()

	var links []pkg.LinkInfo

	for rows.Next() {
		var li pkg.LinkInfo

		if err = rows.Scan(&li.Link, &li.LastUpdateTime, &li.Tags); err != nil {
			return nil, fmt.Errorf("scan get user links: %w", err)
		}
		links = append(links, li)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return links, nil
}

func (r *LinkRepository) GetAllLinks(ctx context.Context, host string, limit int, offset int) ([]pkg.LinkInfo, error) {
	q := database.GetQuerier(ctx, r.db)

	query, args, err := r.builder.
		Select("url", "updated_at").
		From("links").
		Where(squirrel.Expr("url like '%' || ? || '%'", host)).
		OrderBy("id").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build get all links: %w", err)
	}

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query get all links: %w", err)
	}
	defer rows.Close()

	var links []pkg.LinkInfo

	for rows.Next() {
		var li pkg.LinkInfo

		if err = rows.Scan(&li.Link, &li.LastUpdateTime); err != nil {
			return nil, fmt.Errorf("scan links: %w", err)
		}

		links = append(links, li)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return links, nil
}

func (r *LinkRepository) UpdateLinksTime(ctx context.Context, newTime time.Time, url string) error {
	q := database.GetQuerier(ctx, r.db)

	query, args, err := r.builder.
		Update("links").
		Set("updated_at", newTime).
		Where(squirrel.Eq{"url": url}).
		ToSql()

	if err != nil {
		return fmt.Errorf("build update links: %w", err)
	}

	_, err = q.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update links: %w", err)
	}

	return nil
}

func (r *LinkRepository) GetChatIDsByLink(ctx context.Context, link string) ([]int64, error) {
	q := database.GetQuerier(ctx, r.db)

	query, args, err := r.builder.
		Select("lc.chat_id").
		From("link_chat lc").
		Join("links l on lc.link_id = l.id").
		Where(squirrel.Eq{"l.url": link}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build get chatIDs by link: %w", err)
	}

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query get chatIDs by link: %w", err)
	}
	defer rows.Close()

	var chatIDs []int64

	for rows.Next() {
		var chatID int64

		if err = rows.Scan(&chatID); err != nil {
			return nil, fmt.Errorf("scan chatIDs: %w", err)
		}

		chatIDs = append(chatIDs, chatID)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return chatIDs, nil
}
