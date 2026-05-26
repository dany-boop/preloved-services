// Package rabbitmq provides publish/consume helpers for RabbitMQ.
// Uses the official amqp091-go client (maintained by RabbitMQ team).
//
// PUBLISH USAGE (e.g. auth-service sends welcome email):
//   pub, err := rabbitmq.NewPublisher(cfg.RabbitMQURL)
//   err = pub.Publish("email.queue", types.EmailMessage{To: "user@example.com", Template: "welcome"})
//
// CONSUME USAGE (notification-service listens):
//   cons, err := rabbitmq.NewConsumer(cfg.RabbitMQURL)
//   err = cons.Consume("email.queue", func(msg []byte) error {
//       // process message
//       return nil
//   })

package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ──────────────────────────────────────────────
// Publisher
// ──────────────────────────────────────────────

// Publisher sends messages to queues.
type Publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewPublisher connects to RabbitMQ and returns a Publisher.
// url format: "amqp://user:pass@host:5672/vhost"
func NewPublisher(url string) (*Publisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("connect rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("open channel: %w", err)
	}

	return &Publisher{conn: conn, channel: ch}, nil
}

// Publish sends any struct as JSON to the named queue.
// The queue is declared if it doesn't exist (durable = survives restart).
func (p *Publisher) Publish(queueName string, message interface{}) error {
	// Declare queue (idempotent — safe to call every time)
	_, err := p.channel.QueueDeclare(
		queueName,
		true,  // durable: survives RabbitMQ restart
		false, // auto-delete: no
		false, // exclusive: no
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("declare queue %s: %w", queueName, err)
	}

	// Serialize the message to JSON
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return p.channel.PublishWithContext(ctx,
		"",        // exchange (empty = default direct exchange)
		queueName, // routing key = queue name
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // message survives RabbitMQ restart
			Body:         body,
		},
	)
}

// Close cleans up connections.
func (p *Publisher) Close() {
	p.channel.Close()
	p.conn.Close()
}

// ──────────────────────────────────────────────
// Consumer
// ──────────────────────────────────────────────

// Consumer reads messages from a queue.
type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewConsumer connects to RabbitMQ for consuming messages.
func NewConsumer(url string) (*Consumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("connect rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("open channel: %w", err)
	}

	// Process 1 message at a time per consumer (fair dispatch)
	// Prevents one slow consumer from hogging all messages
	if err := ch.Qos(1, 0, false); err != nil {
		return nil, fmt.Errorf("set qos: %w", err)
	}

	return &Consumer{conn: conn, channel: ch}, nil
}

// Consume starts listening to a queue. Calls handler for each message.
// handler should return nil on success, error to nack and requeue.
// This blocks — run it in a goroutine.
func (c *Consumer) Consume(queueName string, handler func([]byte) error) error {
	// Declare queue (idempotent)
	_, err := c.channel.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("declare queue %s: %w", queueName, err)
	}

	msgs, err := c.channel.Consume(
		queueName,
		"",    // consumer tag (auto-generated)
		false, // auto-ack: false (we manually ack after processing)
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("register consumer: %w", err)
	}

	for msg := range msgs {
		if err := handler(msg.Body); err != nil {
			// Nack = failed, requeue for retry
			msg.Nack(false, true)
		} else {
			// Ack = processed successfully, remove from queue
			msg.Ack(false)
		}
	}

	return nil
}

// Close cleans up connections.
func (c *Consumer) Close() {
	c.channel.Close()
	c.conn.Close()
}
