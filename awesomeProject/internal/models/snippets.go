package models

import (
	"context"
	"strconv"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

type SnippetModel struct {
	DB *pgxpool.Pool
}

func (m *SnippetModel) Insert(title string, content string, expires int) (int, error) {
	conn, err := m.DB.Acquire(context.Background())
	if err != nil {
		return 0, err
	}
	expireString := "interval '" + strconv.Itoa(expires) + " day'"
	row := conn.QueryRow(context.Background(),
		"INSERT INTO snippets (title, content, created, expires) VALUES ($1, $2, now(), now() + "+expireString+") RETURNING id", title, content)
	var id uint64
	err = row.Scan(&id)
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (m *SnippetModel) Get(id int) (*Snippet, error) {
	conn, err := m.DB.Acquire(context.Background())
	if err != nil {
		return nil, err
	}
	row := conn.QueryRow(context.Background(),
		"select id, title, content, created, expires from snippets where expires > now() and id = $1", id)
	s := &Snippet{}

	err = row.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	return s, nil
}

func (m *SnippetModel) Latest() ([]*Snippet, error) {
	conn, err := m.DB.Acquire(context.Background())
	if err != nil {
		return nil, err
	}
	rows, err2 := conn.Query(context.Background(), "select id, title, content, created, expires from snippets where expires > now() order by id desc limit 10")
	defer conn.Release()

	if err2 != nil {
		return nil, err2
	}

	snippets := []*Snippet{}

	for rows.Next() {
		s := &Snippet{}
		err = rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
		if err != nil {
			return nil, err
		}
		snippets = append(snippets, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return snippets, nil
}
