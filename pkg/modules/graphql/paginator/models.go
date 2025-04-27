package paginator

import "entgo.io/contrib/entgql"

type Noder interface {
	IsNode()
}

type CursorPaginatedEntity interface {
	IsCursorPaginatedEntity()
}

type EntityCursorEdge interface {
	IsEntityCursorEdge()
}

type EntityOffsetEdge interface {
	IsEntityOffsetEdge()
}

type OffsetPaginatedEntity interface {
	IsOffsetPaginatedEntity()
}

type CursorPageInfoEntity struct {
	StartCursor     *entgql.Cursor[string] `json:"startCursor,omitempty"`
	EndCursor       *entgql.Cursor[string] `json:"endCursor,omitempty"`
	TotalCount      *int                   `json:"totalCount,omitempty"`
	HasPreviousPage bool                   `json:"hasPreviousPage"`
	HasNextPage     bool                   `json:"hasNextPage"`
}

type CursorPaginationArgs struct {
	Cursor *entgql.Cursor[string] `json:"cursor,omitempty"`
	After  *entgql.Cursor[string] `json:"after,omitempty"`
	Before *entgql.Cursor[string] `json:"before,omitempty"`
	Take   *int                   `json:"take,omitempty"`
}

type OffsetPageInfoEntity struct {
	CurrentPage *int `json:"currentPage,omitempty"`
	TotalPages  *int `json:"totalPages,omitempty"`
	TotalCount  *int `json:"totalCount,omitempty"`
}

type OffsetPaginationArgs struct {
	Page *int `json:"page,omitempty"`
	Take *int `json:"take,omitempty"`
}
