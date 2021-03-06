package sqlite

import (
  "database/sql"
  "github.com/manabie-com/togo/internal/context"
  "github.com/manabie-com/togo/internal/core"
  "github.com/mattn/go-sqlite3"
  "log"
)


type UserRepo struct {
  DB             *sql.DB
}

func (repo *UserRepo) Hash(password string) (string, error) {
  return password, nil
}

func (repo *UserRepo) Compare(password, hash string) bool {
  return password == hash
}

func (repo *UserRepo) Create(ctx context.Context, user *core.User, password string) (err error) {
  user.Hash, err = repo.Hash(password)
  if err != nil {
    return
  }
  rs, err := repo.DB.ExecContext(ctx, "insert into users(id, password, max_todo) values (?,?,?)", user.ID, user.Hash,
    user.MaxTodo)
  if err != nil {
    if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode != sqlite3.ErrConstraintPrimaryKey {
      log.Printf("[sqlite::UserRepo::Create - exec error: %v (code: %v, extCode: %v)]\n", err, int(sqliteErr.Code),
        int(sqliteErr.ExtendedCode))
      return err
    } else if !ok {
      log.Printf("[sqlite::UserRepo::Create - exec error: %v]\n", err)
      return err
    }
    return core.ErrUserAlreadyExists
  }
  if affected, _ := rs.RowsAffected(); affected == 0 {
    return core.ErrUserAlreadyExists
  }
  return err
}

func (repo *UserRepo) ById(ctx context.Context, id string) (*core.User, error) {
  var user core.User
  user.ID = id
  row := repo.DB.QueryRowContext(ctx, "select password, max_todo from users where id=?", id)
  err := row.Scan(&user.Hash, &user.MaxTodo)
  if err != nil {
    if err != sql.ErrNoRows {
      log.Printf("[sqlite::UserRepo::ByUser - row scan error: %v]\n", err)
      return nil, err
    }
    return nil, core.ErrUserNotFound
  }
  return &user, nil
}

func (repo *UserRepo) Validate(ctx context.Context, userId string, password string) (*core.User, error) {
  var user core.User
  user.ID = userId
  row := repo.DB.QueryRowContext(ctx, "select password, max_todo from users where id=?", userId)
  err := row.Scan(&user.Hash, &user.MaxTodo)
  if err != nil {
    if err != sql.ErrNoRows {
      log.Printf("[sqlite::UserRepo::ValidateUser - row scan error: %v]\n", err)
      return nil, err
    }
    return nil, core.ErrWrongIdPassword
  }

  if !repo.Compare(password, user.Hash) {
    return nil, core.ErrWrongIdPassword
  }
  return &user, nil
}
