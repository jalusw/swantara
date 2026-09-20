package helper

import "strings"

func MaskEmail(s string) string {
	at := strings.LastIndexByte(s, '@')
	if at <= 0 {
		return MaskValue(s)
	}

	local := s[:at]
	if len([]rune(local)) > 1 {
		local = string([]rune(local)[0]) + "***"
	}

	domain := s[at+1:]
	labels := strings.Split(domain, ".")
	if len(labels) < 2 {
		domain = "***"
	} else {
		masked := make([]string, 0, len(labels))
		for i, label := range labels {
			if i == len(labels)-1 {
				masked = append(masked, label)
			} else {
				masked = append(masked, "***")
			}
		}
		domain = strings.Join(masked, ".")
	}

	return local + "@" + domain
}

func MaskValue(s string) string {
	runes := []rune(s)
	switch len(runes) {
	case 0:
		return ""
	case 1, 2:
		return strings.Repeat("*", len(runes))
	default:
		return string(runes[0]) + strings.Repeat("*", len(runes)-2) + string(runes[len(runes)-1])
	}
}
