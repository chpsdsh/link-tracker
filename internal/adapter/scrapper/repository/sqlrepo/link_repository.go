package sqlrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/pkg/database"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/scrapper"
)

type LinkRepository struct {
	db *pgxpool.Pool
}

func NewLinkRepository(db *pgxpool.Pool) *LinkRepository {
	return &LinkRepository{db: db}
}

func (r *LinkRepository) AddLink(ctx context.Context, chatID int64, link pkg.LinkInfo) error {
	q := database.GetQuerier(ctx, r.db)
	var linkID int64

	err := q.QueryRow(ctx, `
	insert into links (url)
	values($1)
	on conflict (url) do update set url = excluded.url
	returning id
	`, link.Link).Scan(&linkID)

	if err != nil {
		return fmt.Errorf("error insert link: %w", err)
	}

	commandTag, err := q.Exec(ctx, `
	insert into link_chat (chat_id,link_id)
	values($1,$2)
	on conflict do nothing
	`, chatID, linkID)

	if err != nil {
		if database.IsForeignKeyViolation(err) {
			return scrapper.ErrChatNotFound
		}
		return fmt.Errorf("error insert link: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return scrapper.ErrLinkExists
	}

	for _, tag := range link.Tags {
		var tagID int64
		err = q.QueryRow(ctx, `
		insert into tags (tag)
		values($1)
		on conflict (tag) do update set tag = excluded.tag
		returning id
		`, tag).Scan(&tagID)

		if err != nil {
			return fmt.Errorf("error insert link: %w", err)
		}

		_, err = q.Exec(ctx, `
		insert into link_tag (link_id, tag_id)
		values($1,$2)
		on conflict do nothing
		`, linkID, tagID)

		if err != nil {
			return fmt.Errorf("error insert link_tag: %w", err)
		}
	}
	return nil
}

func (r *LinkRepository) LinkExists(ctx context.Context, url string) (bool, error) {
	q := database.GetQuerier(ctx, r.db)

	var exists bool
	err := q.QueryRow(ctx, `
	select exists( 
	select 1 from links where url = $1 
	)`, url).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("error query link exists: %w", err)
	}
	return exists, nil
}

func (r *LinkRepository) DeleteLink(ctx context.Context, chatID int64, url string) (pkg.LinkInfo, error) {
	q := database.GetQuerier(ctx, r.db)

	var li pkg.LinkInfo
	var tags []string

	err := q.QueryRow(ctx, `
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
	`, chatID, url).Scan(&li.Link, &li.LastUpdateTime, &tags)

	if err != nil {
		return pkg.LinkInfo{}, fmt.Errorf("error delete link from link_chat: %w", err)
	}

	li.Tags = tags

	_, err = q.Exec(ctx, `
		delete from links
		where id = (
			select id from links where url = $1
		)
		and not exists (
			select 1 from link_chat lc
			join links l on l.id = lc.link_id
			where l.url = $1
		)
	`, url)

	if err != nil {
		return pkg.LinkInfo{}, fmt.Errorf("error deleting link link: %w", err)
	}

	return li, nil
}

func (r *LinkRepository) GetUserLinks(ctx context.Context, chatID int64) ([]pkg.LinkInfo, error) {
	q := database.GetQuerier(ctx, r.db)
	rows, err := q.Query(ctx, `
	select 
    l.url,
    l.updated_at,
    coalesce(array_agg(t.tag) filter (where t.tag is not null), '{}') as tags
	from link_chat lc
	join links l on l.id = lc.link_id
	left join link_tag lt on lt.link_id = l.id
	left join tags t on t.id = lt.tag_id
	where lc.chat_id = $1
	group by l.id, l.url, l.updated_at
	`, chatID)

	if err != nil {
		return nil, fmt.Errorf("error get user links: %w", err)
	}

	defer rows.Close()

	var links []pkg.LinkInfo

	for rows.Next() {
		var li pkg.LinkInfo

		if err = rows.Scan(&li.Link, &li.LastUpdateTime, &li.Tags); err != nil {
			return nil, fmt.Errorf("error get user links: %w", err)
		}
		links = append(links, li)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error get user links: %w", err)
	}

	return links, nil
}

func (r *LinkRepository) GetAllLinks(ctx context.Context, limit int, offset int) ([]pkg.LinkInfo, error) {
	q := database.GetQuerier(ctx, r.db)
	rows, err := q.Query(ctx, `
	select url, updated_at from links
	where id > $1
    order by id
    limit $2
	`, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("error query get all links: %w", err)
	}

	defer rows.Close()
	var links []pkg.LinkInfo
	for rows.Next() {
		var li pkg.LinkInfo
		if err = rows.Scan(&li.Link, &li.LastUpdateTime); err != nil {
			return nil, fmt.Errorf("error get links: %w", err)
		}
		links = append(links, li)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error get links: %w", err)
	}
	return links, nil
}

func (r *LinkRepository) UpdateLinksTime(ctx context.Context, newTime time.Time, url string) error {
	q := database.GetQuerier(ctx, r.db)
	_, err := q.Exec(ctx, `
	update links 
	set updated_at = $1
	where url = $2
	`, newTime, url)
	if err != nil {
		return fmt.Errorf("error update links: %w", err)
	}

	return nil
}

func (r *LinkRepository) GetChatIDsByLink(ctx context.Context, link string) ([]int64, error) {
	q := database.GetQuerier(ctx, r.db)

	var chatIDs []int64
	rows, err := q.Query(ctx, `
	select lc.chat_id from link_chat lc
	join links l on lc.link_id = l.id
	where l.url = $1
	`, link)
	if err != nil {
		return nil, fmt.Errorf("error get chatIDs by link: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var chatID int64

		if err = rows.Scan(&chatID); err != nil {
			return nil, fmt.Errorf("error get chatIDs by link: %w", err)
		}
		chatIDs = append(chatIDs, chatID)
	}
	return chatIDs, nil
}
