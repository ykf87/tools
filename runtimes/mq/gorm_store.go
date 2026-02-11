package mq

import (
	"time"
	"tools/runtimes/db"

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
	db *db.SQLiteWriter
}

var MqClient *MQ

func init() {
	MqClient = New(NewGormStore(db.MQDB), 3)
	MqClient.Start()
}

func NewGormStore(db *db.SQLiteWriter) *GormStore {
	_ = db.DB().AutoMigrate(&MqModel{})
	return &GormStore{db: db}
}

func (s *GormStore) Save(topic, payload string) (int64, error) {
	msg := &MqModel{
		Topic:   topic,
		Payload: payload,
		Status:  "pending",
	}
	if err := s.db.DB().Create(msg).Error; err != nil {
		return 0, err
	}
	return msg.ID, nil
}

func (s *GormStore) LoadPending() ([]*Message, error) {
	var rows []MqModel
	if err := s.db.Write(func(tx *gorm.DB) error {
		return tx.Where("status = ?", "pending").Order("id asc").Find(&rows).Error
	}); err != nil {
		return nil, err
	}
	// if err := s.db.
	// 	Where("status = ?", "pending").
	// 	Order("id asc").
	// 	Find(&rows).Error; err != nil {
	// 	return nil, err
	// }

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
	return s.db.Write(func(tx *gorm.DB) error {
		return tx.Model(&MqModel{}).
			Where("id = ?", id).
			Update("status", "done").Error
	})
	// return s.db.Model(&MqModel{}).
	// 	Where("id = ?", id).
	// 	Update("status", "done").Error
}

func (s *GormStore) MarkFailed(id int64) error {
	return s.db.Write(func(tx *gorm.DB) error {
		return tx.Model(&MqModel{}).
			Where("id = ?", id).
			Update("status", "failed").Error
	})
	// return s.db.Model(&MqModel{}).
	// 	Where("id = ?", id).
	// 	Update("status", "failed").Error
}
