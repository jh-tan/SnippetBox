package models

import (
	"context"
  "errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Snippet struct { 
  ID int
  Title string
  Content string
  Created time.Time
  Expires time.Time
 }

type SnippetModel struct { 
  DB *pgxpool.Pool
  CONTEXT context.Context
 }

func (m *SnippetModel) Insert(title string, content string, expires int) (int, error) {
  stmt := `INSERT INTO snippets (title, content, created, expires)
  VALUES(@title, @content, NOW(), NOW() + INTERVAL '1 DAY' * @expires) RETURNING id;`
  args := pgx.NamedArgs{
    "title" : title,
    "content": content,
    "expires": expires, 
  }

  var id int
  err:= m.DB.QueryRow(m.CONTEXT, stmt, args).Scan(&id)
  
  if err != nil {
    return 0, err
  }

  return id, nil
}

func (m *SnippetModel) Get(id int) (*Snippet, error) {
  stmt := `SELECT id, title, content, created, expires FROM snippets WHERE expires > NOW() AND id = $1`
  s := &Snippet{}

  err := m.DB.QueryRow(m.CONTEXT, stmt, id).Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)

  if err != nil {
    if errors.Is(err, pgx.ErrNoRows) {
      return nil, ErrNoRecord
    } else {
      return nil, err
    }
  }

  return s, nil
}

func (m *SnippetModel) Latest() ([]Snippet, error) {
  stmt := `SELECT id, title, content, created, expires FROM snippets WHERE expires > NOW() order by id DESC LIMIT 10`

  rows, err := m.DB.Query(m.CONTEXT, stmt)
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  return pgx.CollectRows(rows, pgx.RowToStructByName[Snippet])
}
