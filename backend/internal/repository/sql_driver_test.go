package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

type sqlExpectation struct {
	kind      string
	match     string
	checkArgs func([]driver.NamedValue) error
	rows      *sqlRows
	result    driver.Result
	err       error
}

type sqlRows struct {
	columns []string
	values  [][]driver.Value
	index   int
}

func (r *sqlRows) Columns() []string { return r.columns }
func (r *sqlRows) Close() error      { return nil }
func (r *sqlRows) Next(dest []driver.Value) error {
	if r.index >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.index])
	r.index++
	return nil
}

type sqlTestState struct {
	expectations []sqlExpectation
	pos          int
	mu           sync.Mutex
}

func (s *sqlTestState) next(kind, query string, args []driver.NamedValue) (sqlExpectation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.pos >= len(s.expectations) {
		return sqlExpectation{}, fmt.Errorf("unexpected %s: %s", kind, query)
	}

	exp := s.expectations[s.pos]
	s.pos++

	if exp.kind != kind {
		return sqlExpectation{}, fmt.Errorf("got %s, want %s", kind, exp.kind)
	}
	if exp.match != "" && !strings.Contains(query, strings.TrimSpace(exp.match)) {
		return sqlExpectation{}, fmt.Errorf("query %q does not contain %q", query, strings.TrimSpace(exp.match))
	}
	if exp.checkArgs != nil {
		if err := exp.checkArgs(args); err != nil {
			return sqlExpectation{}, err
		}
	}
	return exp, nil
}

func (s *sqlTestState) verify(t *testing.T) {
	t.Helper()
	if s.pos != len(s.expectations) {
		t.Fatalf("processed %d expectations, want %d", s.pos, len(s.expectations))
	}
}

type sqlTestDriver struct{ state *sqlTestState }
type sqlTestConn struct{ state *sqlTestState }
type sqlTestTx struct{ state *sqlTestState }

func (d *sqlTestDriver) Open(string) (driver.Conn, error) { return &sqlTestConn{state: d.state}, nil }
func (c *sqlTestConn) Prepare(string) (driver.Stmt, error) { return nil, fmt.Errorf("not implemented") }
func (c *sqlTestConn) Close() error                        { return nil }
func (c *sqlTestConn) Begin() (driver.Tx, error)          { return c.BeginTx(context.Background(), driver.TxOptions{}) }

func (c *sqlTestConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	exp, err := c.state.next("begin", "", nil)
	if err != nil {
		return nil, err
	}
	if exp.err != nil {
		return nil, exp.err
	}
	return &sqlTestTx{state: c.state}, nil
}

func (c *sqlTestConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	exp, err := c.state.next("exec", query, args)
	if err != nil {
		return nil, err
	}
	if exp.err != nil {
		return nil, exp.err
	}
	if exp.result == nil {
		return driver.RowsAffected(0), nil
	}
	return exp.result, nil
}

func (c *sqlTestConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	exp, err := c.state.next("query", query, args)
	if err != nil {
		return nil, err
	}
	if exp.err != nil {
		return nil, exp.err
	}
	if exp.rows == nil {
		return &sqlRows{}, nil
	}
	return exp.rows, nil
}

func (t *sqlTestTx) Commit() error {
	exp, err := t.state.next("commit", "", nil)
	if err != nil {
		return err
	}
	return exp.err
}

func (t *sqlTestTx) Rollback() error {
	exp, err := t.state.next("rollback", "", nil)
	if err != nil {
		return err
	}
	return exp.err
}

var sqlDriverCounter uint64

func newTestDB(t *testing.T, expectations []sqlExpectation) *sql.DB {
	t.Helper()

	state := &sqlTestState{expectations: expectations}
	name := fmt.Sprintf("sqltest-%d", atomic.AddUint64(&sqlDriverCounter, 1))
	sql.Register(name, &sqlTestDriver{state: state})

	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
		state.verify(t)
	})

	return db
}

func namedValueAt(args []driver.NamedValue, idx int) any {
	return args[idx].Value
}
