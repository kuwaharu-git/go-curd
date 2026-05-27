package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type Todo struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

type createTodoRequest struct {
	Title string `json:"title"`
}

func main() {
	host := getEnv("DB_HOST", "db")
	port := getEnvAsInt("DB_PORT", 3306)
	user := getEnv("DB_USER", "todo")
	password := getEnv("DB_PASSWORD", "todo")
	name := getEnv("DB_NAME", "todo")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4", user, password, host, port, name)
	db, err := connectWithRetry(dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := ensureSchema(db); err != nil {
		log.Fatalf("failed to initialize schema: %v", err)
	}

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./static")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	api := r.Group("/api")
	{
		api.GET("/todos", func(c *gin.Context) {
			todos, err := listTodos(db)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch todos"})
				return
			}
			c.JSON(http.StatusOK, todos)
		})

		api.POST("/todos", func(c *gin.Context) {
			var req createTodoRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
				return
			}
			req.Title = trimSpace(req.Title)
			if req.Title == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
				return
			}

			todo, err := createTodo(db, req.Title)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create todo"})
				return
			}
			c.JSON(http.StatusCreated, todo)
		})

		api.PUT("/todos/:id/toggle", func(c *gin.Context) {
			id, err := strconv.ParseInt(c.Param("id"), 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
				return
			}
			todo, err := toggleTodo(db, id)
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "todo not found"})
				return
			}
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update todo"})
				return
			}
			c.JSON(http.StatusOK, todo)
		})

		api.DELETE("/todos/:id", func(c *gin.Context) {
			id, err := strconv.ParseInt(c.Param("id"), 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
				return
			}
			if err := deleteTodo(db, id); err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "todo not found"})
				return
			} else if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete todo"})
				return
			}
			c.Status(http.StatusNoContent)
		})
	}

	serverPort := getEnv("PORT", "8080")
	if err := r.Run(":" + serverPort); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvAsInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func connectWithRetry(dsn string) (*sql.DB, error) {
	var (
		db  *sql.DB
		err error
	)
	for i := 0; i < 20; i++ {
		db, err = sql.Open("mysql", dsn)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}
		err = db.Ping()
		if err == nil {
			return db, nil
		}
		time.Sleep(2 * time.Second)
	}
	return nil, err
}

func ensureSchema(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS todos (
id BIGINT AUTO_INCREMENT PRIMARY KEY,
title VARCHAR(255) NOT NULL,
completed BOOLEAN NOT NULL DEFAULT FALSE,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)
`)
	return err
}

func listTodos(db *sql.DB) ([]Todo, error) {
	rows, err := db.Query(`SELECT id, title, completed FROM todos ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	todos := []Todo{}
	for rows.Next() {
		var todo Todo
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Completed); err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return todos, nil
}

func createTodo(db *sql.DB, title string) (*Todo, error) {
	result, err := db.Exec(`INSERT INTO todos (title) VALUES (?)`, title)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &Todo{ID: id, Title: title, Completed: false}, nil
}

func toggleTodo(db *sql.DB, id int64) (*Todo, error) {
	if _, err := db.Exec(`UPDATE todos SET completed = NOT completed WHERE id = ?`, id); err != nil {
		return nil, err
	}
	var todo Todo
	row := db.QueryRow(`SELECT id, title, completed FROM todos WHERE id = ?`, id)
	if err := row.Scan(&todo.ID, &todo.Title, &todo.Completed); err != nil {
		return nil, err
	}
	return &todo, nil
}

func deleteTodo(db *sql.DB, id int64) error {
	result, err := db.Exec(`DELETE FROM todos WHERE id = ?`, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func trimSpace(text string) string {
	return strings.TrimSpace(text)
}
