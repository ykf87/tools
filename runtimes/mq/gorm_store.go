package mq

import (
	"time"

	"gorm.io/gorm"
)

type MqModel struct {
	ID        int64  `gorm:"primaryKey"`
	Topic     string `gorm:"index;not null"`
	Payload   string `gorm:"type:text;not null"`
	Status    string `gorm:"index;not null"` // pending / done / failed
	CreatedAt time.Time
	UpdatedAt time.Time
}

type GormStore struct {
	db *gorm.DB
}

func NewGormStore(db *gorm.DB) *GormStore {
	_ = db.AutoMigrate(&MqModel{})
	return &GormStore{db: db}
}

func (s *GormStore) Save(topic, payload string) (int64, error) {
	msg := &MqModel{
		Topic:   topic,
		Payload: payload,
		Status:  "pending",
	}
	if err := s.db.Create(msg).Error; err != nil {
		return 0, err
	}
	return msg.ID, nil
}

func (s *GormStore) LoadPending() ([]*Message, error) {
	var rows []MqModel
	if err := s.db.
		Where("status = ?", "pending").
		Order("id asc").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	msgs := make([]*Message, 0, len(rows))
	for _, r := range rows {
		msgs = append(msgs, &Message{
			ID:      r.ID,
			Topic:   r.Topic,
			Payload: r.Payload,
		})
	}
	return msgs, nil
}

func (s *GormStore) MarkDone(id int64) error {
	return s.db.Model(&MqModel{}).
		Where("id = ?", id).
		Update("status", "done").Error
}

func (s *GormStore) MarkFailed(id int64) error {
	return s.db.Model(&MqModel{}).
		Where("id = ?", id).
		Update("status", "failed").Error
}
