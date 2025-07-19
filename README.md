# go-uow

A generic Unit of Work (UOW) pattern implementation for Go with PostgreSQL support using pgx/v5.

## Overview

This library provides a clean and type-safe implementation of the Unit of Work pattern, designed to simplify database transaction management in Go applications. It supports PostgreSQL through the `pgx/v5` driver and offers a generic interface that can work with any repository registry.

## Features

- **Generic Design**: Type-safe implementation using Go generics
- **PostgreSQL Support**: Built on top of `pgx/v5` for high-performance PostgreSQL operations
- **Transaction Management**: Automatic transaction handling with commit/rollback
- **Isolation Levels**: Support for all PostgreSQL transaction isolation levels
- **Read-Only Transactions**: Built-in support for read-only transaction modes
- **Error Handling**: Comprehensive error handling with custom error types
- **Resource Management**: Automatic connection pool management

## Installation

```bash
go get github.com/callmemars1/go-uow
```

## Quick Start

### 1. Define Your Repository Registry

```go
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id string) (*User, error)
}

type OrderRepository interface {
    Create(ctx context.Context, order *Order) error
    GetByUserID(ctx context.Context, userID string) ([]*Order, error)
}

type RepoRegistry struct {
    Users  UserRepository
    Orders OrderRepository
}
```

### 2. Create Repository Factory

```go
func NewRepoRegistry(tx pgx.Tx) RepoRegistry {
    return RepoRegistry{
        Users:  NewUserRepository(tx),
        Orders: NewOrderRepository(tx),
    }
}
```

### 3. Initialize UOW Factory

```go
pool, err := pgxpool.New(ctx, "postgres://user:password@localhost:5432/dbname")
if err != nil {
    log.Fatal(err)
}

factory, err := pgxv5.NewFactory(pool, NewRepoRegistry)
if err != nil {
    log.Fatal(err)
}
// Note: In real applications, the factory should be a long-lived resource
// Only call factory.Release() when shutting down your application
```

### 4. Use the Unit of Work

```go
// Simple transaction
err := uow.RunTx(ctx, factory, func(uow uow.UOW[RepoRegistry]) error {
    repos := uow.MustRepoRegistry()
    
    user := &User{Name: "John Doe"}
    if err := repos.Users.Create(ctx, user); err != nil {
        return err
    }
    
    order := &Order{UserID: user.ID, Amount: 100}
    return repos.Orders.Create(ctx, order)
}, uow.DefaultTxOptions())

// Transaction with result
result, err := uow.RunTxWithResult(ctx, factory, func(uow uow.UOW[RepoRegistry]) (*User, error) {
    repos := uow.MustRepoRegistry()
    return repos.Users.GetByID(ctx, "user-123")
}, uow.SerializableTxOptions())
```

## API Reference

### Core Interfaces

#### UOW Interface
```go
type UOW[TRepoRegistry any] interface {
    MustRepoRegistry() TRepoRegistry
    Begin(ctx context.Context, options TxOptions) error
    Commit(ctx context.Context) error
    Rollback(ctx context.Context) error
}
```

#### Factory Interface
```go
type Factory[TRepoRegistry any] interface {
    NewUOW(ctx context.Context) (UOW[TRepoRegistry], error)
    Release() error
}
```

### Transaction Options

The library uses standard `database/sql.TxOptions` for transaction configuration:

```go
type TxOptions struct {
    Isolation IsolationLevel
    ReadOnly  bool
}
```

#### Predefined Options
- `DefaultTxOptions()`: Read committed, read-write
- `SerializableTxOptions()`: Serializable, read-write  
- `ReadOnlyTxOptions()`: Read committed, read-only

#### Standard Isolation Levels
- `sql.LevelDefault`
- `sql.LevelReadUncommitted`
- `sql.LevelReadCommitted`
- `sql.LevelRepeatableRead`
- `sql.LevelSerializable`
- `sql.LevelSnapshot`
- `sql.LevelLinearizable`

### Helper Functions

#### RunTx
Executes a transaction without returning a result:
```go
func RunTx[TRepoRegistry any](
    ctx context.Context,
    factory Factory[TRepoRegistry],
    action TxAction[TRepoRegistry],
    options TxOptions,
) error
```

#### RunTxWithResult
Executes a transaction and returns a result:
```go
func RunTxWithResult[TRepoRegistry any, TReturn any](
    ctx context.Context,
    factory Factory[TRepoRegistry],
    action TxActionWithResult[TRepoRegistry, TReturn],
    options TxOptions,
) (res *TReturn, err error)
```

## Error Handling

The library provides custom error types:
- `ErrTransactionNotStarted`: Returned when attempting to access repositories before starting a transaction

## Best Practices

1. **Always use the helper functions**: `RunTx` and `RunTxWithResult` handle transaction lifecycle automatically
2. **Define repository interfaces**: This ensures type safety and testability
3. **Use appropriate isolation levels**: Choose based on your application's consistency requirements
4. **Handle errors properly**: Always check returned errors from UOW operations
5. **Release resources**: Call `factory.Release()` only when shutting down your application, not immediately after creation

## Factory Lifecycle Management

The factory should be a long-lived resource in your application. Here are common patterns:

### Global Variable Pattern
```go
var factory uow.Factory[RepoRegistry]

func init() {
    pool, err := pgxpool.New(context.Background(), "postgres://...")
    if err != nil {
        log.Fatal(err)
    }
    
    factory, err = pgxv5.NewFactory(pool, NewRepoRegistry)
    if err != nil {
        log.Fatal(err)
    }
}

func shutdown() {
    if factory != nil {
        factory.Release()
    }
}
```

### Dependency Injection Pattern
```go
type App struct {
    Factory uow.Factory[RepoRegistry]
}

func NewApp() *App {
    pool, err := pgxpool.New(context.Background(), "postgres://...")
    if err != nil {
        log.Fatal(err)
    }
    
    factory, err := pgxv5.NewFactory(pool, NewRepoRegistry)
    if err != nil {
        log.Fatal(err)
    }
    
    return &App{Factory: factory}
}

func (app *App) Shutdown() {
    if app.Factory != nil {
        app.Factory.Release()
    }
}
```

## Example: Complete Application

```go
package main

import (
    "context"
    "log"
    
    "github.com/callmemars1/go-uow"
    "github.com/callmemars1/go-uow/pgxv5"
    "github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
    ID   string
    Name string
}

type UserRepository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id string) (*User, error)
}

type userRepository struct {
    tx pgx.Tx
}

func NewUserRepository(tx pgx.Tx) UserRepository {
    return &userRepository{tx: tx}
}

func (r *userRepository) Create(ctx context.Context, user *User) error {
    // Implementation using r.tx
    return nil
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*User, error) {
    // Implementation using r.tx
    return nil, nil
}

type RepoRegistry struct {
    Users UserRepository
}

func NewRepoRegistry(tx pgx.Tx) RepoRegistry {
    return RepoRegistry{
        Users: NewUserRepository(tx),
    }
}

func main() {
    ctx := context.Background()
    
    pool, err := pgxpool.New(ctx, "postgres://user:password@localhost:5432/dbname")
    if err != nil {
        log.Fatal(err)
    }
    
    factory, err := pgxv5.NewFactory(pool, NewRepoRegistry)
    if err != nil {
        log.Fatal(err)
    }
    // In a real application, you would typically:
    // 1. Store the factory as a global variable, or
    // 2. Pass it through dependency injection, or
    // 3. Use it as a singleton
    // Only call factory.Release() when shutting down your application
    
    // Create a user
    err = uow.RunTx(ctx, factory, func(uow uow.UOW[RepoRegistry]) error {
        repos := uow.MustRepoRegistry()
        user := &User{Name: "John Doe"}
        return repos.Users.Create(ctx, user)
    }, uow.DefaultTxOptions())
    
    if err != nil {
        log.Fatal(err)
    }
}
```

## Custom Database Implementations

The library provides base interfaces that allow you to implement the UOW pattern for any database. The core interfaces are database-agnostic:

### Core Interfaces
```go
type UOW[TRepoRegistry any] interface {
    MustRepoRegistry() TRepoRegistry
    Begin(ctx context.Context, options TxOptions) error
    Commit(ctx context.Context) error
    Rollback(ctx context.Context) error
}

type Factory[TRepoRegistry any] interface {
    NewUOW(ctx context.Context) (UOW[TRepoRegistry], error)
    Release() error
}
```

### Implementing for Other Databases

To implement UOW for a different database, you need to:

1. **Implement the `UOW` interface** for your database's transaction type
2. **Implement the `Factory` interface** to create UOW instances
3. **Use the same helper functions** (`RunTx`, `RunTxWithResult`) with your implementation

#### Example: MySQL Implementation
```go
import (
    "database/sql"
    "github.com/go-sql-driver/mysql"
)

type MySQLUOW[TRepoRegistry any] struct {
    tx           *sql.Tx
    repoRegistry TRepoRegistry
}

func (u *MySQLUOW[TRepoRegistry]) MustRepoRegistry() TRepoRegistry {
    return u.repoRegistry
}

func (u *MySQLUOW[TRepoRegistry]) Begin(ctx context.Context, options TxOptions) error {
    // MySQL doesn't support all PostgreSQL isolation levels
    // Map to supported levels or use defaults
    return nil // Transaction already started
}

func (u *MySQLUOW[TRepoRegistry]) Commit(ctx context.Context) error {
    return u.tx.Commit()
}

func (u *MySQLUOW[TRepoRegistry]) Rollback(ctx context.Context) error {
    return u.tx.Rollback()
}

type MySQLFactory[TRepoRegistry any] struct {
    db           *sql.DB
    repoFactory  func(*sql.Tx) TRepoRegistry
}

func NewMySQLFactory(db *sql.DB, repoFactory func(*sql.Tx) TRepoRegistry) *MySQLFactory[TRepoRegistry] {
    return &MySQLFactory[TRepoRegistry]{
        db:          db,
        repoFactory: repoFactory,
    }
}

func (f *MySQLFactory[TRepoRegistry]) NewUOW(ctx context.Context) (uow.UOW[TRepoRegistry], error) {
    tx, err := f.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, err
    }
    
    return &MySQLUOW[TRepoRegistry]{
        tx:           tx,
        repoRegistry: f.repoFactory(tx),
    }, nil
}

func (f *MySQLFactory[TRepoRegistry]) Release() error {
    return f.db.Close()
}
```

#### Example: SQLite Implementation
```go
import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)

type SQLiteUOW[TRepoRegistry any] struct {
    tx           *sql.Tx
    repoRegistry TRepoRegistry
}

func (u *SQLiteUOW[TRepoRegistry]) MustRepoRegistry() TRepoRegistry {
    return u.repoRegistry
}

func (u *SQLiteUOW[TRepoRegistry]) Begin(ctx context.Context, options TxOptions) error {
    return nil // Transaction already started
}

func (u *SQLiteUOW[TRepoRegistry]) Commit(ctx context.Context) error {
    return u.tx.Commit()
}

func (u *SQLiteUOW[TRepoRegistry]) Rollback(ctx context.Context) error {
    return u.tx.Rollback()
}

type SQLiteFactory[TRepoRegistry any] struct {
    db           *sql.DB
    repoFactory  func(*sql.Tx) TRepoRegistry
}

func NewSQLiteFactory(db *sql.DB, repoFactory func(*sql.Tx) TRepoRegistry) *SQLiteFactory[TRepoRegistry] {
    return &SQLiteFactory[TRepoRegistry]{
        db:          db,
        repoFactory: repoFactory,
    }
}

func (f *SQLiteFactory[TRepoRegistry]) NewUOW(ctx context.Context) (uow.UOW[TRepoRegistry], error) {
    tx, err := f.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, err
    }
    
    return &SQLiteUOW[TRepoRegistry]{
        tx:           tx,
        repoRegistry: f.repoFactory(tx),
    }, nil
}

func (f *SQLiteFactory[TRepoRegistry]) Release() error {
    return f.db.Close()
}
```

### Usage with Custom Implementations

Once you have your custom implementation, you can use it with the same helper functions:

```go
// MySQL example
db, err := sql.Open("mysql", "user:password@/dbname")
if err != nil {
    log.Fatal(err)
}

factory := NewMySQLFactory(db, NewRepoRegistry)
defer factory.Release()

err = uow.RunTx(ctx, factory, func(uow uow.UOW[RepoRegistry]) error {
    repos := uow.MustRepoRegistry()
    // Your transaction logic here
    return nil
}, uow.DefaultTxOptions())
```

## Requirements

- Go 1.24.4 or later
- PostgreSQL database (for the provided pgx/v5 implementation)
- `github.com/jackc/pgx/v5` driver (for the provided pgx/v5 implementation)