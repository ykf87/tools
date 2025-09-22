package mqs

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Mq struct {
	ID          int64     `gorm:"primaryKey"`
	Topic       string    `gorm:"index;not null"`
	Payload     string    `gorm:"type:text;not null"`
	Status      string    `gorm:"index;not null;default:pending"` // pending, processing, done, failed
	RetryCount  int       `gorm:"default:0"`                      // 重试次数
	MaxRetry    int       `gorm:"default:3"`                      // 最大重试次数
	AvailableAt time.Time `gorm:"index"`                          // 延迟消费
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Store struct {
	db *gorm.DB
}

// func init() {
// 	db.MQDB.AutoMigrate(&Mq{})
// }

func NewStore(db *gorm.DB) *Store {
	// 自动迁移
	db.AutoMigrate(&Mq{})
	return &Store{db: db}
}

func (s *Store) Enqueue(topic, payload string, delay time.Duration) (int64, error) {
	msg := &Mq{
		Topic:       topic,
		Payload:     payload,
		Status:      "pending",
		AvailableAt: time.Now().Add(delay),
	}
	if err := s.db.Create(msg).Error; err != nil {
		return 0, err
	}
	return msg.ID, nil
}

func (s *Store) FetchPending(topic string) (*Mq, error) {
	var msg Mq
	tx := s.db.Begin()
	if err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("topic = ? AND status = ? AND available_at <= ?", topic, "pending", time.Now()).
		Order("id ASC").
		First(&msg).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	msg.Status = "processing"
	if err := tx.Save(&msg).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return &msg, nil
}

func (s *Store) MarkDone(id int64) error {
	return s.db.Model(&Mq{}).Where("id = ?", id).Update("status", "done").Error
}

func (s *Store) MarkFailed(id int64) error {
	return s.db.Model(&Mq{}).Where("id = ?", id).Update("status", "failed").Error
}

func (s *Store) RetryMessage(id int64) error {
	var msg Mq
	if err := s.db.First(&msg, id).Error; err != nil {
		return err
	}
	if msg.RetryCount >= msg.MaxRetry {
		return s.MarkFailed(id)
	}
	msg.RetryCount++
	msg.Status = "pending"
	msg.AvailableAt = time.Now().Add(5 * time.Second) // 延迟 5 秒后重试
	return s.db.Save(&msg).Error
}

func (s *Store) FetchByID(id int64) (*Mq, error) {
	var msg Mq
	err := s.db.First(&msg, id).Error
	if err != nil {
		return nil, err
	}
	return &msg, nil
}
