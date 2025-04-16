package kafka

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"time"
	"todo-app/internal/domain/requests"

	"github.com/IBM/sarama"
)

type Producer struct {
	producer sarama.SyncProducer
	topic    string
	logger   *slog.Logger
}

func NewProducer(brokers []string, topic string, logger *slog.Logger) (*Producer, error) {
	config := sarama.NewConfig()

	// 1. Критические настройки таймаутов
	config.Net.DialTimeout = 5 * time.Second   // Таймаут подключения к брокеру
	config.Net.WriteTimeout = 10 * time.Second // Таймаут отправки сообщения
	config.Net.ReadTimeout = 10 * time.Second  // Таймаут чтения ответа
	config.Metadata.Retry.Max = 10             // Повторы для получения метаданных
	config.Metadata.Retry.Backoff = 1 * time.Second

	// 2. Гарантия доставки (усиленная версия)
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 15                  // Больше попыток
	config.Producer.Retry.Backoff = 2 * time.Second // Увеличенный backoff
	config.Producer.Idempotent = true
	config.Producer.MaxMessageBytes = 1000000 // 1MB max
	config.Producer.Return.Successes = true

	// 3. Оптимизация для медленных сетей
	config.Net.MaxOpenRequests = 1
	config.Producer.Timeout = 30 * time.Second    // Общий таймаут продюсера
	config.ClientID = "high-reliability-producer" // Идентификация в логах Kafka

	// Проверка доступности брокеров перед созданием продюсера
	if err := validateBrokers(brokers, config, logger); err != nil {
		return nil, fmt.Errorf("broker validation failed: %w", err)
	}

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer (check broker availability): %w", err)
	}

	logger.Info("Producer initialized",
		"brokers", brokers,
		"timeouts", fmt.Sprintf("dial=%v, write=%v", config.Net.DialTimeout, config.Net.WriteTimeout))

	return &Producer{producer: producer, topic: topic, logger: logger}, nil
}
func validateBrokers(brokers []string, config *sarama.Config, logger *slog.Logger) error {
	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		return fmt.Errorf("can't connect to any broker: %w", err)
	}
	defer client.Close()

	// Проверяем доступность каждого брокера
	for _, broker := range brokers {
		connected := client.Brokers()
		if len(connected) == 0 {
			logger.Warn("Broker not reachable", "address", broker)
		}
	}

	return nil
}

func (p *Producer) SendApiEvent(apiRequest *requests.ApiRequest) error {
	jsonData, err := json.Marshal(apiRequest)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(apiRequest.Action),
		Value: sarama.ByteEncoder(jsonData),
	}

	partition, offset, err := p.producer.SendMessage(msg)

	if err != nil {
		p.logger.Error("failed to send message", "partition", partition, "offset", offset)
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (p *Producer) Close() error {
	return p.producer.Close()
}
