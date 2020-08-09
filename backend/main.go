package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// Namespace model
type Namespace struct {
	gorm.Model
	Name     string
	Projects []Project
}

// Project model
type Project struct {
	gorm.Model
	Namespace   Namespace
	NamespaceID *uint
	Name        string
	URL         string
	Pipelines   []Pipeline
}

// Pipeline model
type Pipeline struct {
	gorm.Model
	Project    Project
	ProjectID  *uint
	Ref        string
	Status     string `gorm:"default:'failed'"`
	FinishedAt *time.Time
}

// WebhookData json bind model
type WebhookData struct {
	Object_attributes struct {
		Ref         string
		Status      string
		Finished_at time.Time
	}
	Project struct {
		Name    string
		Web_url string
	}
}

func main() {
	db, err := gorm.Open("sqlite3", "file::memory:?cache=shared")
	defer db.Close()

	if err != nil {
		panic("failed to connect database")
	}

	db.LogMode(true)
	db.AutoMigrate(
		&Namespace{},
		&Project{},
		&Pipeline{},
	)

	var wsConnections = []*websocket.Conn{}
	var wsupgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	wsupgrader.CheckOrigin = func(_ *http.Request) bool {
		return true
	}

	router := gin.Default()
	router.Use(cors.Default())

	router.GET("/namespaces", func(c *gin.Context) {
		var namespaces []Namespace

		db.Find(&namespaces)

		c.JSON(200, namespaces)
	})

	router.GET("/namespaces/:namespace", func(c *gin.Context) {
		var namespace Namespace

		db.Preload("Projects").Preload("Projects.Pipelines").First(&namespace, &Namespace{Name: c.Param("namespace")})

		c.JSON(200, namespace)
	})

	router.POST("/namespaces/:namespace", func(c *gin.Context) {
		var project Project
		var webhookData WebhookData
		var namespace Namespace
		var pipeline Pipeline

		db.FirstOrCreate(&namespace, &Namespace{Name: c.Param("namespace")})

		c.BindJSON(&webhookData)
		fmt.Println(webhookData)

		db.FirstOrCreate(&project, &Project{
			Name:        webhookData.Project.Name,
			URL:         webhookData.Project.Web_url,
			NamespaceID: &namespace.ID,
		})

		db.FirstOrCreate(&pipeline, &Pipeline{
			Ref:       webhookData.Object_attributes.Ref,
			ProjectID: &project.ID,
		}).UpdateColumn(&Pipeline{
			Status:     webhookData.Object_attributes.Status,
			FinishedAt: &webhookData.Object_attributes.Finished_at,
		})

		db.Preload("Projects").Preload("Projects.Pipelines").First(&namespace, &Namespace{Name: c.Param("namespace")})

		for _, conn := range wsConnections {
			websocket.WriteJSON(conn, &namespace)
		}

		c.JSON(200, gin.H{"success": true})
	})

	router.GET("/ws", func(c *gin.Context) {
		conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			fmt.Println("Failed to set websocket upgrade: %+v", err)
			return
		}

		wsConnections = append(wsConnections, conn)
	})

	router.Run("0.0.0.0:8081")
}
