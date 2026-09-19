package selection

import (
	"strings"
	"time"
)

func normalizeScope(scope Scope) (Scope, error) {
	switch scope.View {
	case "all":
		if scope.AlbumID != "" || scope.FolderPath != "" {
			return Scope{}, ErrInvalid
		}
	case "album":
		if scope.AlbumID == "" || scope.FolderPath != "" {
			return Scope{}, ErrInvalid
		}
	case "folder":
		scope.FolderPath = strings.TrimRight(scope.FolderPath, "/")
		if scope.FolderPath == "" || scope.AlbumID != "" {
			return Scope{}, ErrInvalid
		}
	default:
		return Scope{}, ErrInvalid
	}
	if scope.GPSFilter == "" {
		scope.GPSFilter = "no-gps"
	}
	if scope.HiddenFilter == "" {
		scope.HiddenFilter = "visible"
	}
	if scope.GPSFilter != "no-gps" && scope.GPSFilter != "with-gps" && scope.GPSFilter != "all" {
		return Scope{}, ErrInvalid
	}
	if scope.HiddenFilter != "visible" && scope.HiddenFilter != "hidden" && scope.HiddenFilter != "all" {
		return Scope{}, ErrInvalid
	}
	for _, day := range []string{scope.StartDate, scope.EndDate} {
		if day != "" {
			parsed, err := time.Parse("2006-01-02", day)
			if err != nil || parsed.Format("2006-01-02") != day {
				return Scope{}, ErrInvalid
			}
		}
	}
	if scope.StartDate != "" && scope.EndDate != "" && scope.StartDate > scope.EndDate {
		return Scope{}, ErrInvalid
	}
	return scope, nil
}
