package onedb

import (
	"database/sql"
	"errors"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
)

var errInvalidSqlQueryType = errors.New("Invalid query. Must be of type *SqlQuery")
var sqllibCreate sqllibCreator = &sqllibRealCreator{}

type sqllibCreator interface {
	Open(driverName, dataSourceName string) (sqlLibBackender, error)
}

type sqllibRealCreator struct{}

func (o *sqllibRealCreator) Open(driverName, dataSourceName string) (sqlLibBackender, error) {
	return sql.Open(driverName, dataSourceName)
}

type SqlQuery struct {
	query string
	args  []interface{}
}

func NewSqlQuery(query string, args ...interface{}) *SqlQuery {
	return &SqlQuery{query: query, args: args}
}

type sqllibBackend struct {
	db sqlLibBackender
	backender
}

type sqlLibBackender interface {
	Ping() error
	Close() error
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

func NewSqllib(driverName, connectionString string) (DBer, error) {
	sqlDb, err := sqllibCreate.Open(driverName, connectionString)
	if err != nil {
		return nil, err
	}
	err = sqlDb.Ping()
	if err != nil {
		return nil, err
	}
	return newBackendConverter(&sqllibBackend{db: sqlDb}), nil
}

func (b *sqllibBackend) Close() error {
	return b.db.Close()
}

func (b *sqllibBackend) Query(query interface{}) (rowsScanner, error) {
	q, ok := query.(*SqlQuery)
	if !ok {
		return nil, errInvalidSqlQueryType
	}
	return b.db.Query(q.query, q.args...)
}

func (b *sqllibBackend) Execute(command interface{}) error {
	c, ok := command.(*SqlQuery)
	if !ok {
		return errInvalidSqlQueryType
	}
	_, err := b.db.Exec(c.query, c.args...)
	return err
}
