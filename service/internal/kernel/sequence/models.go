package sequence

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

type ResetPeriod string

const (
	ResetNever   ResetPeriod = "never"
	ResetYearly  ResetPeriod = "yearly"
	ResetMonthly ResetPeriod = "monthly"
)

var ErrInvalidResetPeriod = errors.New("invalid reset period")

type DocumentSequence struct {
	model.Base
	OrganizationID uint64 `gorm:"not null;uniqueIndex:idx_doc_seq_organization_code" json:"organization_id"`
	Code           string `gorm:"not null;uniqueIndex:idx_doc_seq_organization_code" json:"code"`
	Prefix         string `json:"prefix"`
	Suffix         string `json:"suffix"`
	NextNumber     int64  `gorm:"default:1;not null" json:"next_number"`
	Padding        int    `gorm:"default:5;not null" json:"padding"`
	ResetPeriod    string `json:"reset_period"`
}

func (DocumentSequence) TableName() string {
	return "doc_sequences"
}

func ParseResetPeriod(value string) (ResetPeriod, error) {
	period := ResetPeriod(strings.TrimSpace(value))
	switch period {
	case ResetNever, ResetYearly, ResetMonthly:
		return period, nil
	case "":
		return ResetNever, nil
	default:
		return "", ErrInvalidResetPeriod
	}
}

func (s DocumentSequence) Format(value int64) string {
	return s.Prefix + fmt.Sprintf("%0*d", s.Padding, value) + s.Suffix
}

func shouldReset(period string, lastUsed, now time.Time) bool {
	switch ResetPeriod(period) {
	case ResetYearly:
		return now.Year() != lastUsed.Year()
	case ResetMonthly:
		return now.Year() != lastUsed.Year() || now.Month() != lastUsed.Month()
	default:
		return false
	}
}
