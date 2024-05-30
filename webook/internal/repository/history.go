package repository

import (
	"context"
	"github.com/ClearloveHn/golangwebook/webook/internal/domain"
)

type HistoryRecordRepository interface {
	AddRecord(ctx context.Context, record domain.HistoryRecord) error
}
