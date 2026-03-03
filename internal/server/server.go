package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/oscar/messaging-playgrounds/internal/config"
	"github.com/oscar/messaging-playgrounds/internal/models"
	"github.com/oscar/messaging-playgrounds/internal/rabbitmq"
)

type Server struct {
	Router    *gin.Engine
	publisher *rabbitmq.Publisher
}

func NewServer(cfg *config.ServerConfig) (*Server, error) {
	publisher, err := rabbitmq.NewPublisher(cfg.RabbitMQURL, cfg.QueueName)
	if err != nil {
		return nil, err
	}

	router := gin.Default()

	server := &Server{
		Router:    router,
		publisher: publisher,
	}

	server.setupRoutes()

	return server, nil
}

func (s *Server) setupRoutes() {
	s.Router.GET("/health", s.healthCheck)
	s.Router.POST("/api/messages", s.publishMessage)
}

func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time":   time.Now().UTC(),
	})
}

func (s *Server) publishMessage(c *gin.Context) {
	var req models.MessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.MessageResponse{
			Success: false,
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	message := models.Message{
		ID:        uuid.New().String(),
		Content:   req.Content,
		Timestamp: time.Now().UTC(),
		Metadata:  req.Metadata,
	}

	if err := s.publisher.Publish(c.Request.Context(), message); err != nil {
		c.JSON(http.StatusInternalServerError, models.MessageResponse{
			Success: false,
			Message: "Failed to publish message: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{
		Success: true,
		Message: "Message published successfully",
		ID:      message.ID,
	})
}

func (s *Server) Close() error {
	return s.publisher.Close()
}
