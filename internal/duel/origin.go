package duel

import (
	"errors"
	"net/url"
	"strings"
)

// ParseAllowedOrigins accepts the original single value or an explicit CSV list.
// Entries are origin tuples, never URL paths, patterns or opaque origins.
func ParseAllowedOrigins(raw string) ([]string, error) {
	origins := []string{}
	if strings.TrimSpace(raw) == "" {
		return origins, nil
	}
	seen := map[string]bool{}
	for _, entry := range strings.Split(raw, ",") {
		origin := strings.TrimSpace(entry)
		u, err := url.Parse(origin)
		if err != nil || origin == "" || strings.Contains(origin, "*") ||
			(u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" ||
			u.User != nil || u.Path != "" || u.RawQuery != "" || u.ForceQuery ||
			u.Fragment != "" || u.RawFragment != "" || u.Opaque != "" ||
			strings.HasSuffix(u.Host, ":") || strings.ContainsAny(origin, "?#") {
			return nil, errors.New("invalid exact origin configuration")
		}
		if !seen[origin] {
			origins = append(origins, origin)
			seen[origin] = true
		}
	}
	return origins, nil
}

func (h *Handler) allowsOrigin(origin string) bool {
	origins, err := ParseAllowedOrigins(h.Origin)
	if err != nil {
		return false
	}
	for _, allowed := range origins {
		if origin == allowed {
			return true
		}
	}
	return false
}
