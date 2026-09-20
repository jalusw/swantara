package helper

import "time"

func ParseBirthday(value *string) (*time.Time, error) {
	return ParseDate(value)
}

func ParseDate(value *string) (*time.Time, error) {
	if value == nil || *value == "" {
		return nil, nil
	}

	birthday, err := time.Parse("2006-01-02", *value)
	if err != nil {
		return nil, err
	}

	return &birthday, nil
}

func ParseDateOrToday(value string) (time.Time, error) {
	if value == "" {
		return time.Now(), nil
	}
	return time.Parse("2006-01-02", value)
}

func FormatDatePtr(value *time.Time) *string {
	if value == nil || value.IsZero() {
		return nil
	}
	formatted := value.Format("2006-01-02")
	return &formatted
}

func ParseTimestamp(value *string) (*time.Time, error) {
	if value == nil || *value == "" {
		return nil, nil
	}

	timestamp, err := time.Parse(time.RFC3339, *value)
	if err != nil {
		return nil, err
	}

	return &timestamp, nil
}
