package httpx

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
)

type Link struct {
	Rel    string `json:"rel"`
	Href   string `json:"href"`
	Method string `json:"method,omitempty"`
}

var linkBaseURL string

func SetLinkBaseURL(base string) {
	linkBaseURL = strings.TrimSuffix(base, "/")
}

func LinkBaseFor(c fiber.Ctx) string {
	if linkBaseURL != "" {
		return linkBaseURL
	}
	return c.BaseURL()
}

func getLink(rel, href string) Link {
	return Link{Rel: rel, Href: href}
}

func paginationLinks(c fiber.Ctx, pagination *PaginationMeta) []Link {
	if pagination == nil || pagination.TotalPages <= 1 {
		return nil
	}

	u, err := url.ParseRequestURI(c.OriginalURL())
	if err != nil {
		return nil
	}

	pageCandidates := []struct {
		rel  string
		page int
	}{
		{rel: "first", page: 1},
		{rel: "prev", page: pagination.Page - 1},
		{rel: "next", page: pagination.Page + 1},
		{rel: "last", page: pagination.TotalPages},
	}

	links := make([]Link, 0, len(pageCandidates))
	for _, candidate := range pageCandidates {
		if candidate.page < 1 || candidate.page > pagination.TotalPages {
			continue
		}
		query := u.Query()
		query.Set("page", strconv.Itoa(candidate.page))
		query.Set("size", strconv.Itoa(pagination.PerPage))
		u.RawQuery = query.Encode()
		links = append(links, getLink(candidate.rel, LinkBaseFor(c)+u.String()))
	}
	return links
}
