package manager

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Manager manages RabbitMQ connections and channels
type Manager struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	mu      sync.RWMutex
	config  Config
}

// Config holds RabbitMQ connection configuration
type Config struct {
	User           string
	Password       string
	Host           string
	Port           string
	VHost          string
	ConnectionName string // Optional: label for the connection in RabbitMQ management console
}

// NewManager creates a new RabbitMQ manager
func NewManager(config Config) *Manager {
	return &Manager{
		config: config,
	}
}

// Connect establishes connection to RabbitMQ
func (r *Manager) Connect(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	url := fmt.Sprintf("amqp://%s:%s@%s:%s%s",
		r.config.User,
		r.config.Password,
		r.config.Host,
		r.config.Port,
		r.config.VHost)

	// Configure client properties for connection labeling
	config := amqp.Config{
		Properties: amqp.Table{},
	}

	// Set connection name if provided (visible in RabbitMQ management console)
	if r.config.ConnectionName != "" {
		config.Properties["connection_name"] = r.config.ConnectionName
	}

	conn, err := amqp.DialConfig(url, config)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	r.conn = conn
	r.channel = channel

	log.Println("Successfully connected to RabbitMQ")
	return nil
}

// GetChannel returns the RabbitMQ channel
func (r *Manager) GetChannel() *amqp.Channel {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.channel
}

// GetConnection returns the RabbitMQ connection
func (r *Manager) GetConnection() *amqp.Connection {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.conn
}

// IsConnected checks if the connection is active
func (r *Manager) IsConnected() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.conn != nil && !r.conn.IsClosed()
}

// Reconnect attempts to reconnect to RabbitMQ
func (r *Manager) Reconnect(ctx context.Context) error {
	r.Close()
	return r.Connect(ctx)
}

// Close closes the RabbitMQ connection and channel
func (r *Manager) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var err error
	if r.channel != nil {
		if closeErr := r.channel.Close(); closeErr != nil {
			err = closeErr
		}
		r.channel = nil
	}

	if r.conn != nil {
		if closeErr := r.conn.Close(); closeErr != nil {
			err = closeErr
		}
		r.conn = nil
	}

	log.Println("RabbitMQ connection closed")
	return err
}

// StartHeartbeat starts a goroutine to monitor connection health
func (r *Manager) StartHeartbeat(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if !r.IsConnected() {
					log.Println("RabbitMQ connection lost, attempting to reconnect...")
					if err := r.Reconnect(ctx); err != nil {
						log.Printf("Failed to reconnect to RabbitMQ: %v", err)
					}
				}
			}
		}
	}()
}

// PublishMessage publishes a message to a queue
func (r *Manager) PublishMessage(ctx context.Context, exchange, routingKey string, message []byte) error {
	channel := r.GetChannel()
	if channel == nil {
		return fmt.Errorf("no active channel available")
	}

	return channel.PublishWithContext(
		ctx,
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        message,
		},
	)
}

// DeclareQueue declares a queue
func (r *Manager) DeclareQueue(name string, durable, autoDelete, exclusive, noWait bool) (amqp.Queue, error) {
	channel := r.GetChannel()
	if channel == nil {
		return amqp.Queue{}, fmt.Errorf("no active channel available")
	}

	return channel.QueueDeclare(
		name,       // name
		durable,    // durable
		autoDelete, // delete when unused
		exclusive,  // exclusive
		noWait,     // no-wait
		nil,        // arguments
	)
}
