package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
)

var session *gocql.Session

func init() {
	var err error

	
	cluster.Timeout = 30 * time.Second // Adjust the timeout duration as needed

	
	cluster.Keyspace = "todo"
	session, err = cluster.CreateSession()
	if err != nil {
    	panic("Failed to connect to Cassandra/ScyllaDB: " + err.Error())
	}


    var query = session.Query("SELECT * FROM system.clients")

    if rows, err := query.Iter().SliceMap(); err == nil {
        for _, row := range rows {
            fmt.Printf("%v\n", row)
        }
    } else {
        panic("Query error: " + err.Error())
    }

	// Create keyspace and table if not exists
}

func createKeyspaceAndTable() {
	if err := session.Query(`
		CREATE KEYSPACE IF NOT EXISTS todo
		WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };
	`).Exec(); err != nil {
		log.Fatal(err)
	}

	if err := session.Query(`
		CREATE TABLE IF NOT EXISTS todo.todos (
			todo_id TEXT,
			user_id TEXT,
			title TEXT,
			description TEXT,
			status TEXT,
			created TIMESTAMP,
			updated TIMESTAMP,
			created_formatted TEXT,
			updated_formatted TEXT,
			PRIMARY KEY (user_id, todo_id)
		);
	`).Exec(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	r := gin.Default()
	createKeyspaceAndTable()


	r.POST("/todos", createTodo)
	r.GET("/todos/:userId", getTodos)
	r.GET("/todos/:userId/:todoId", getTodo)
	r.PUT("/todos/:userId/:todoId", updateTodo)
	r.DELETE("/todos/:userId/:todoId", deleteTodo)
	r.GET("/todoStatus/:status", filterTodos)
	r.GET("/todos", listTodos)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	err := r.Run(":" + port)
	if err != nil {
		log.Fatal(err)
	}
}

// TODO struct
type Todo struct {
 	TodoId       string `json:"todo_id"`
	UserID      string `json:"user_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Created     int64     `json:"created"`
	Updated     int64      `json:"updated"`
	CreatedFormatted   string     `json:"created_formatted"`
    UpdatedFormatted   string     `json:"updated_formatted"`
}

// Create a new TODO
func createTodo(c *gin.Context) {
	var todo Todo
	if err := c.ShouldBindJSON(&todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// todo todoId = gocql.RandomUUID().String()
	todo_id, err := gocql.RandomUUID()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    todo.TodoId = todo_id.String()
    // todo.UserID = gocql.TimeUUID()
	userID, err := gocql.RandomUUID()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    todo.UserID = userID.String()
	// todo.Created =time.Now().Unix() 
	todo.Updated = todo.Created

	todo.CreatedFormatted = time.Unix(time.Now().Unix(), 0).Format("01/02/2006 15:04:05")
    todo.UpdatedFormatted = time.Unix(time.Now().Unix(),0).Format("01/02/2006 15:04:05")

	if err := session.Query(`
		INSERT INTO todo.todos (todo_id, user_id, title, description, status, updated, created_formatted, updated_formatted)
		VALUES (?, ?, ?, ?, ?, ?, ?,?)
	`, todo.TodoId, todo.UserID, todo.Title, todo.Description, todo.Status, todo.Updated, todo.CreatedFormatted, todo.UpdatedFormatted).Exec(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, todo)
}

// Get all TODOs for a user
func getTodos(c *gin.Context) {
    var todos []Todo

    userID := c.Param("userId")
    // parsedUserID, err := gocql.ParseUUID(userID)
    // if err != nil {
    //     c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user todoId"})
    //     return
    // }
	log.Println(userID, "parsedUserId")
    query := session.Query(`SELECT * FROM todo.todos WHERE user_id = ?`, userID)
    if rows, err := query.Iter().SliceMap(); err == nil {
        for _, row := range rows {
            // Convert each row to Todo and append to the slice
            todo := Todo{
             TodoId:          row["todo_id"].(string),
                UserID:      row["user_id"].(string),
                Title:       row["title"].(string),
                Description: row["description"].(string),
                Status:      row["status"].(string),
                CreatedFormatted:      row["created_formatted"].(string),
                UpdatedFormatted:      row["updated_formatted"].(string),
                // Created:     row["created"].(int64),
                // Updated:     row["updated"].(int64),
            }
            todos = append(todos, todo)
        }
        c.JSON(http.StatusOK, todos)
    } else {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    }
}

// Get a specific TODO for a user
func getTodo(c *gin.Context) {
    userID := c.Param("userId")
    todoID := c.Param("todoId")

    // parsedUserID, err := gocql.ParseUUID(userID)
    // if err != nil {
    //     c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user todoId"})
    //     return
    // }
    // parsedTodoId, err := gocql.ParseUUID(todoID)
    // if err != nil {
    //     c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user todoId"})
    //     return
    // }

    query := session.Query(`SELECT * FROM todo.todos WHERE user_id = ? AND todo_id = ?`, userID, todoID)
    var todo Todo
    if err := query.MapScan(map[string]interface{}{
        "todo_id":          &todo.TodoId,
        "user_id":     &todo.UserID,
        "title":       &todo.Title,
        "description": &todo.Description,
        "status":      &todo.Status,
        "created":     &todo.Created,
        "updated":     &todo.Updated,
		"CreatedFormatted":      &todo.CreatedFormatted,
        "UpdatedFormatted":     &todo.UpdatedFormatted,
    }); err == nil {
        c.JSON(http.StatusOK, todo)
    } else if err == gocql.ErrNotFound {
        c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
    } else {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    }
}


// Update a TODO for a user
func updateTodo(c *gin.Context) {
	var updatedTodo Todo
    if err := c.ShouldBindJSON(&updatedTodo); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    userID := c.Param("userId")
    todoID := c.Param("todoId")

	var existingTodo Todo
    if err := session.Query(`
        SELECT * FROM todo.todos
        WHERE user_id = ? AND todo_id = ?
    `, userID, todoID).MapScan(map[string]interface{}{
        "todo_id":          &existingTodo.TodoId,
        "user_id":          &existingTodo.UserID,
        "title":            &existingTodo.Title,
        "description":      &existingTodo.Description,
        "status":           &existingTodo.Status,
        "created":          &existingTodo.Created,
        "updated":          &existingTodo.Updated,
        "created_formatted": &existingTodo.CreatedFormatted,
        "updated_formatted": &existingTodo.UpdatedFormatted,
    }); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

   // Apply changes from updatedTodo
    existingTodo.Title = updatedTodo.Title
    existingTodo.Description = updatedTodo.Description
    existingTodo.Status = strings.ToLower(updatedTodo.Status)
    existingTodo.UpdatedFormatted = time.Unix(time.Now().Unix(), 0).Format("01/02/2006 15:04:05")

    // Update the Todo in the database
    if err := session.Query(`
        UPDATE todo.todos
        SET title = ?, description = ?, status = ?, updated_formatted = ?
        WHERE user_id = ? AND todo_id = ?
    `, existingTodo.Title, existingTodo.Description, existingTodo.Status, existingTodo.UpdatedFormatted, userID, todoID).Exec(); err == nil {
        c.JSON(http.StatusOK, existingTodo)
    } else {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    }

}

// Delete a TODO for a user
func deleteTodo(c *gin.Context) {
    userID := c.Param("userId")
    todoID := c.Param("todoId")

	log.Println(userID,"userId")
	log.Println(todoID)

    // parsedUserID, err := gocql.ParseUUID(userID)
    // if err != nil {
    //     c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
    //     return
    // }

    if err := session.Query(`
        DELETE FROM todo.todos
        WHERE user_id = ? 
    `, userID).Exec(); err == nil {
        c.JSON(http.StatusNoContent, nil)
    } else {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    }
}


func filterTodos(c *gin.Context) {
    var todos []Todo

    status := strings.ToLower(c.Param("status"))
    query := session.Query(`SELECT * FROM todo.todos WHERE status = ? ALLOW FILTERING`, status)
    if rows, err := query.Iter().SliceMap(); err == nil {
        for _, row := range rows {
            todo := Todo{
                TodoId:             row["todo_id"].(string),
                UserID:             row["user_id"].(string),
                Title:              row["title"].(string),
                Description:        row["description"].(string),
                Status:             row["status"].(string),
                CreatedFormatted:  row["created_formatted"].(string),
                UpdatedFormatted:  row["updated_formatted"].(string),
            }

            todos = append(todos, todo)
        }
        c.JSON(http.StatusOK, todos)
    } else {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    }
}

func listTodos(c *gin.Context) {
	pageStr := c.Query("page")
	sizeStr := c.Query("size")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil || size <= 0 {
		size = 10
	}

	offset := (page - 1) * size

	status := c.Query("status")

	var todos []Todo

	query := session.Query(`
		SELECT * FROM todo.todos
		WHERE status = ? 
		ALLOW FILTERING
		LIMIT ? OFFSET ?
	`, status, size, offset)

	if rows, err := query.Iter().SliceMap(); err == nil {
		for _, row := range rows {
			todo := Todo{
				TodoId:            row["todo_id"].(string),
				UserID:            row["user_id"].(string),
				Title:             row["title"].(string),
				Description:       row["description"].(string),
				Status:            row["status"].(string),
				CreatedFormatted:  row["created_formatted"].(string),
				UpdatedFormatted:  row["updated_formatted"].(string),
			}
			todos = append(todos, todo)
		}
		c.JSON(http.StatusOK, todos)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}