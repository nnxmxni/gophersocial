package store

import (
	"net/http"
	"strconv"
	"strings"
)

type PaginatedFeedQuery struct {
	Limit  int      `json:"limit" validate:"gte=1,lte=20"`
	Offset int      `json:"offset" validate:"gte=0"`
	Sort   string   `json:"sort" validate:"oneof=asc desc"`
	Tags   []string `json:"tags" validate:"max=5"`
	Search string   `json:"search" validate:"max=1000"`
}

func (fq PaginatedFeedQuery) Parse(r *http.Request) (PaginatedFeedQuery, error) {

	qs := r.URL.Query()

	limit := qs.Get("limit")
	offset := qs.Get("offset")
	sort := qs.Get("sort")
	tags := qs.Get("tags")
	search := qs.Get("search")

	if limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			return fq, err
		}

		fq.Limit = l
	}

	if offset != "" {
		l, err := strconv.Atoi(offset)
		if err != nil {
			return fq, err
		}

		fq.Offset = l
	}

	if sort != "" {
		fq.Sort = sort
	}

	if tags != "" {
		fq.Tags = strings.Split(tags, ",")
	}

	if search != "" {
		fq.Search = search
	}

	return fq, nil
}
