package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	_ "github.com/joho/godotenv/autoload"
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
	Pipelines   []Pipeline
}

// Pipeline model
type Pipeline struct {
	gorm.Model
	ExternalID uint `gorm:"default:0"`
	Project    Project
	ProjectID  *uint
	Ref        string
	Status     string `gorm:"default:'failed'"`
	URL        string
	SpawnedAt  *time.Time
	FinishedAt *time.Time
}

// WebhookData json bind model
type WebhookData struct {
	ObjectAttributes struct {
		ID         uint `json:"id"`
		Ref        string
		Status     string
		CreatedAt  string `json:"created_at"`
		FinishedAt string `json:"finished_at"`
	} `json:"object_attributes"`
	Project struct {
		Name   string
		WebURL string `json:"web_url"`
	}
}

var dbFileName string
var listenAddress string
var mergeRefs []string

var db *gorm.DB
var wsConnections = []*websocket.Conn{}
var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	initEnv()
	initDB()
	defer db.Close()

	wsupgrader.CheckOrigin = func(_ *http.Request) bool {
		return true
	}

	router := gin.Default()
	router.Use(cors.Default())

	router.GET("/namespaces", listNamespaces)
	router.GET("/namespaces/:namespace", listProjectsForNamespace)
	router.POST("/namespaces/:namespace", webhook)
	router.DELETE("/namespaces/:namespace/pipelines/:id", deletePipeline)
	router.GET("/ws", webSocketUpgrade)

	router.Run(listenAddress)
}

func initEnv() {
	dbFileName = os.Getenv("DB_FILE")
	if dbFileName == "" {
		dbFileName = "data.db"
	}

	listenAddress = os.Getenv("LISTEN_ADDRESS")
	if listenAddress == "" {
		listenAddress = "0.0.0.0:8081"
	}

	mergeRefsInput := os.Getenv("MERGE_REFS")
	mergeRefs = strings.Split(mergeRefsInput, ",")
}

func initDB() {
	var err error
	db, err = gorm.Open("sqlite3", dbFileName)

	if err != nil {
		panic("failed to connect database")
	}

	db.LogMode(true)
	db.AutoMigrate(
		&Namespace{},
		&Project{},
		&Pipeline{},
	)
}

func listNamespaces(c *gin.Context) {
	var namespaces []Namespace

	db.Find(&namespaces)

	c.JSON(200, namespaces)
}

func listProjectsForNamespace(c *gin.Context) {
	var namespace Namespace

	db.Preload("Projects").Preload("Projects.Pipelines").First(&namespace, Namespace{Name: c.Param("namespace")})

	c.JSON(200, namespace)
}

func webhook(c *gin.Context) {
	var project Project
	var webhookData WebhookData
	var namespace Namespace
	var pipeline Pipeline
	mergeRefs := strings.Split(c.Query("mergeRefs"), ",")

	db.FirstOrCreate(&namespace, Namespace{Name: c.Param("namespace")})

	err := c.BindJSON(&webhookData)
	if err != nil {
		c.Error(err)
	}

	if c.Query("ignoreRefs") != "" {
		ignoreRefs := strings.Split(c.Query("ignoreRefs"), ",")
		for _, ignoreRef := range ignoreRefs {
			matched, _ := regexp.Match(ignoreRef, []byte(webhookData.ObjectAttributes.Ref))
			if matched {
				log.Println("Ignored ref: %v", webhookData.ObjectAttributes.Ref)
				return
			}
		}
	}

	db.FirstOrCreate(&project, Project{
		Name:        webhookData.Project.Name,
		NamespaceID: &namespace.ID,
	})

	refSpec := webhookData.ObjectAttributes.Ref
	for _, mergeRef := range mergeRefs {
		matched, _ := regexp.Match(mergeRef, []byte(webhookData.ObjectAttributes.Ref))
		if matched {
			refSpec = strings.ReplaceAll(mergeRef, ".*", "%")
			break
		}
	}

	db.Where("ref like ? collate nocase", refSpec).FirstOrCreate(&pipeline, Pipeline{
		ProjectID: &project.ID,
	})

	if pipeline.SpawnedAt == nil || webhookData.ObjectAttributes.ID >= pipeline.ExternalID {
		createdAt, _ := dateparse.ParseAny(webhookData.ObjectAttributes.CreatedAt)
		finishedAt, _ := dateparse.ParseAny(webhookData.ObjectAttributes.FinishedAt)
		db.Model(&pipeline).UpdateColumn(Pipeline{
			ExternalID: webhookData.ObjectAttributes.ID,
			Ref:        webhookData.ObjectAttributes.Ref,
			Status:     webhookData.ObjectAttributes.Status,
			URL:        webhookData.Project.WebURL + "/pipelines/" + strconv.FormatUint(uint64(webhookData.ObjectAttributes.ID), 10),
			SpawnedAt:  &createdAt,
			FinishedAt: &finishedAt,
		})

		db.Preload("Projects").Preload("Projects.Pipelines").First(&namespace, &Namespace{Name: c.Param("namespace")})

		for _, conn := range wsConnections {
			websocket.WriteJSON(conn, &namespace)
		}
	}

	c.JSON(200, gin.H{"success": true})
}

func deletePipeline(c *gin.Context) {
	idParam := c.Param("id")

	if idParam == "" {
		c.AbortWithStatus(422)
		return
	}

	id, _ := strconv.Atoi(idParam)
	var pipeline Pipeline
	db.Find(&pipeline, id)
	db.Delete(&pipeline)

	var namespace Namespace
	db.Preload("Projects").Preload("Projects.Pipelines").First(&namespace, Namespace{Name: c.Param("namespace")})

	for _, conn := range wsConnections {
		websocket.WriteJSON(conn, &namespace)
	}

	c.JSON(200, namespace)
}

func webSocketUpgrade(c *gin.Context) {
	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Failed to set websocket upgrade: %+v", err)
		return
	}

	wsConnections = append(wsConnections, conn)
}
