package helper

import "strings"

func ParseOS(ua string) string {
	lower := strings.ToLower(ua)

	switch {
	case strings.Contains(lower, "windows nt 10"):
		return "Windows 10"
	case strings.Contains(lower, "windows nt 6.3"):
		return "Windows 8.1"
	case strings.Contains(lower, "windows nt 6.2"):
		return "Windows 8"
	case strings.Contains(lower, "windows nt 6.1"):
		return "Windows 7"
	case strings.Contains(lower, "windows nt 6.0"):
		return "Windows Vista"
	case strings.Contains(lower, "windows nt 5.1"):
		return "Windows XP"
	case strings.Contains(lower, "windows"):
		return "Windows"
	case strings.Contains(lower, "mac os x") || strings.Contains(lower, "macintosh"):
		return "macOS"
	case strings.Contains(lower, "iphone os") || strings.Contains(lower, "cpu iphone os") || strings.Contains(lower, "cpu like mac os"):
		return "iOS"
	case strings.Contains(lower, "android"):
		return "Android"
	case strings.Contains(lower, "linux"):
		return "Linux"
	}

	return "Unknown"
}

func ParseDeviceName(ua string) string {
	lower := strings.ToLower(ua)

	if strings.Contains(lower, "iphone") {
		return "iPhone"
	}
	if strings.Contains(lower, "ipad") {
		return "iPad"
	}
	if strings.Contains(lower, "macintosh") || strings.Contains(lower, "mac os x") {
		return "Mac"
	}
	if strings.Contains(lower, "android") {
		return "Android Device"
	}
	if strings.Contains(lower, "windows phone") {
		return "Windows Phone"
	}
	if strings.Contains(lower, "windows") {
		return "Windows PC"
	}
	if strings.Contains(lower, "linux") {
		return "Linux Desktop"
	}

	return "Unknown"
}

func ParseBrowser(ua string) string {
	lower := strings.ToLower(ua)

	switch {
	case strings.Contains(lower, "edg/") || strings.Contains(lower, "edge"):
		return "Edge"
	case strings.Contains(lower, "opr/") || strings.Contains(lower, "opera"):
		return "Opera"
	case strings.Contains(lower, "chrome"):
		return "Chrome"
	case strings.Contains(lower, "firefox"):
		return "Firefox"
	case strings.Contains(lower, "safari"):
		return "Safari"
	}

	return "Unknown"
}
