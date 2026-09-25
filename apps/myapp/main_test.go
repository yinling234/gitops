package main

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mark3labs/mcp-go/mcp"
)

// ============================================================
// queryHandler 的测试
// ============================================================

// 用例 1：缺少 query 参数
func TestQueryHandler_MissingQueryArgument(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer mockDB.Close()
	db = mockDB

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{} // 空 map，没有 query

	result, err := queryHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Errorf("expected IsError=true, got false")
	}
}

// 用例 2：Arguments 不是 map 类型
func TestQueryHandler_InvalidArgumentsFormat(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer mockDB.Close()
	db = mockDB

	req := mcp.CallToolRequest{}
	req.Params.Arguments = "this is not a map" // 错误类型

	result, err := queryHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Errorf("expected IsError=true, got false")
	}
}

// 用例 3：正常执行 SQL，返回 JSON
func TestQueryHandler_ValidQuery(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer mockDB.Close()
	db = mockDB

	// 模拟 db.Query() 的返回
	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "Alice").
		AddRow(2, "Bob")
	mock.ExpectQuery("SELECT .* FROM users").WillReturnRows(rows)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"query": "SELECT id, name FROM users",
	}

	result, err := queryHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %v", result)
	}

	// 校验 mock 期望是否全部满足
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

// 用例 4：SQL 执行失败，返回错误
func TestQueryHandler_SQLError(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer mockDB.Close()
	db = mockDB

	mock.ExpectQuery("SELECT .* FROM nonexistent").
		WillReturnError(context.DeadlineExceeded)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"query": "SELECT * FROM nonexistent",
	}

	result, err := queryHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Errorf("expected IsError=true, got false")
	}
}

// ============================================================
// listSchemaHandler 的测试
// ============================================================

// 用例 5：正常列出 schema
func TestListSchemaHandler_Success(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer mockDB.Close()
	db = mockDB

	// SHOW TABLES 返回两个表
	mock.ExpectQuery("SHOW TABLES").WillReturnRows(
		sqlmock.NewRows([]string{"Tables_in_hr_db"}).
			AddRow("users").
			AddRow("orders"),
	)

	// DESCRIBE users
	mock.ExpectQuery("DESCRIBE users").WillReturnRows(
		sqlmock.NewRows([]string{"Field", "Type", "Null", "Key", "Default", "Extra"}).
			AddRow("id", "int", "NO", "PRI", nil, "").
			AddRow("name", "varchar(255)", "YES", "", nil, ""),
	)

	// DESCRIBE orders
	mock.ExpectQuery("DESCRIBE orders").WillReturnRows(
		sqlmock.NewRows([]string{"Field", "Type", "Null", "Key", "Default", "Extra"}).
			AddRow("id", "int", "NO", "PRI", nil, "").
			AddRow("amount", "decimal(10,2)", "YES", "", nil, ""),
	)

	req := mcp.CallToolRequest{}
	result, err := listSchemaHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

// 用例 6：SHOW TABLES 失败
func TestListSchemaHandler_ShowTablesError(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer mockDB.Close()
	db = mockDB

	mock.ExpectQuery("SHOW TABLES").
		WillReturnError(context.DeadlineExceeded)

	req := mcp.CallToolRequest{}
	result, err := listSchemaHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Errorf("expected IsError=true, got false")
	}
}
