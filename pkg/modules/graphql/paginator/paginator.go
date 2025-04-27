package paginator

import (
	"context"
	"math"
	"reflect"

	"entgo.io/contrib/entgql"
	"github.com/99designs/gqlgen/graphql"
	"github.com/gianglt2198/graphql-gateway-go/pkg/utils/reflection"
	"github.com/samber/lo"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type DefaultPaginationParams struct {
	Page *int
	Take *int
}

const (
	DefaultPage = 1
	DefaultTake = 20
)

func OffsetPaginationParams(pagination *OffsetPaginationArgs, defaults ...DefaultPaginationParams) (page, take int) {
	// Default values
	page = DefaultPage
	take = DefaultTake

	if len(defaults) > 0 {
		page = lo.FromPtrOr(defaults[0].Page, DefaultPage)
		take = lo.FromPtrOr(defaults[0].Take, DefaultTake)
	}

	if pagination != nil {
		page = lo.FromPtrOr(pagination.Page, page)
		take = lo.FromPtrOr(pagination.Take, take)
	}

	return
}

func CursorPaginationParams(pagination *CursorPaginationArgs, defaults ...DefaultPaginationParams) (
	after *entgql.Cursor[string],
	first *int,
	before *entgql.Cursor[string],
	last *int,
) {
	if pagination == nil {
		first = lo.ToPtr(DefaultTake)
		return
	}

	if pagination.After == nil && pagination.Before == nil {
		first = pagination.Take
	}
	if pagination.After != nil {
		after = pagination.After
		first = pagination.Take
	}
	if pagination.Before != nil {
		before = pagination.Before
		last = pagination.Take
	}
	return
}

func GetPaginationContext(ctx context.Context) (context.Context, error) {
	fc := graphql.GetFieldContext(ctx)
	opCtx := graphql.GetOperationContext(ctx)
	edgeFields := graphql.CollectFields(opCtx, fc.Field.Selections, []string{})
	var (
		edge *graphql.FieldContext
		err  error
	)

	for _, field := range edgeFields {
		if field.Field.Name == "edges" {
			edge, err = fc.Child(ctx, field)
			break
		}
	}
	if err != nil {
		return nil, gqlerror.Errorf("failed to get pagination edge context: %v", err)
	}

	if edge == nil {
		return nil, gqlerror.Errorf("failed to get pagination edge context")
	}

	nodeFields := graphql.CollectFields(opCtx, edge.Field.Selections, []string{})
	node, err := edge.Child(ctx, nodeFields[0])
	if err != nil {
		return nil, gqlerror.Errorf("failed to get pagination node context: %v", err)
	}
	return graphql.WithFieldContext(ctx, node), nil
}

type OffsetPaginatedOptions[Q any] func(*Q) *Q

type OffsetPaginatedParams[Q, R, E any] struct {
	Pagination *OffsetPaginationArgs
	Query      *Q
	Options    []OffsetPaginatedOptions[Q]
	Factory    func(*R) *E
}

type CursorPaginatedParams[Q, R, E any] struct {
	Pagination *CursorPaginationArgs
	Query      *Q
	Options    []any
	Factory    func(*R) *E
}

func OffsetPaginate[Q, R, E any, O OffsetPaginatedEntity](ctx context.Context, params OffsetPaginatedParams[Q, R, E]) (*O, error) {
	page, take := OffsetPaginationParams(params.Pagination)

	counter := reflection.CallMethodWithValue[*Q]("Clone", params.Query)

	paginatedCtx, err := GetPaginationContext(ctx)
	if err != nil {
		return nil, err
	}

	pager, err := reflection.CallMethodWithError[*Q]("CollectFields", params.Query, paginatedCtx)
	if err != nil {
		return nil, err
	}

	for _, option := range params.Options {
		pager = option(pager)
		counter = option(counter)
	}

	offset := reflection.CallMethodWithValue[*Q]("Offset", pager, (page-1)*take)
	limit := reflection.CallMethodWithValue[*Q]("Limit", offset, take)
	items, err := reflection.CallMethodWithError[[]*R]("All", limit, ctx)
	if err != nil {
		return nil, err
	}

	count, err := reflection.CallMethodWithError[int]("Count", counter, ctx)
	if err != nil {
		return nil, err
	}

	pages := int(math.Ceil(float64(count) / float64(take)))
	obj := new(O)

	err = reflection.SetField(obj, "PageInfo", &OffsetPageInfoEntity{
		CurrentPage: &page,
		TotalPages:  &pages,
		TotalCount:  &count,
	})
	if err != nil {
		return nil, err
	}

	edgesField := reflect.ValueOf(obj).Elem().FieldByName("Edges")
	edgeItemType := edgesField.Type().Elem().Elem()
	edgeSlice := reflect.MakeSlice(edgesField.Type(), 0, len(items))
	for _, i := range items {
		e := reflect.New(edgeItemType)
		e.Elem().FieldByName("Node").Set(reflect.ValueOf(*params.Factory(i)))
		edgeSlice = reflect.Append(edgeSlice, e)
	}

	_ = reflection.SetField(obj, "Edges", edgeSlice.Interface())

	return obj, nil
}

func CursorPaginate[Q, R, E any, O CursorPaginatedEntity](ctx context.Context, params CursorPaginatedParams[Q, R, E]) (*O, error) {
	after, first, before, last := CursorPaginationParams(params.Pagination)
	paginateParams := []any{
		ctx,
		after,
		first,
		before,
		last,
	}
	if params.Options != nil {
		paginateParams = append(paginateParams, params.Options...)
	}
	connections, err := reflection.CallMethodWithError[any](
		"Paginate",
		params.Query,
		paginateParams...,
	)
	if err != nil {
		return nil, err
	}

	obj := new(O)
	rawConnections := reflect.ValueOf(connections).Elem()
	rawEdges := rawConnections.FieldByName("Edges")
	rawPageInfo := rawConnections.FieldByName("PageInfo")
	rawTotalCount := rawConnections.FieldByName("TotalCount")

	err = reflection.SetField(obj, "PageInfo", &CursorPageInfoEntity{
		HasPreviousPage: rawPageInfo.FieldByName("HasPreviousPage").Bool(),
		HasNextPage:     rawPageInfo.FieldByName("HasNextPage").Bool(),
		StartCursor:     rawPageInfo.FieldByName("StartCursor").Interface().(*entgql.Cursor[string]),
		EndCursor:       rawPageInfo.FieldByName("EndCursor").Interface().(*entgql.Cursor[string]),
		TotalCount:      rawTotalCount.Addr().Interface().(*int),
	})
	if err != nil {
		return nil, err
	}

	edgesField := reflect.ValueOf(obj).Elem().FieldByName("Edges")
	edgeItemType := edgesField.Type().Elem().Elem()
	edgeSlice := reflect.MakeSlice(edgesField.Type(), 0, rawEdges.Len())

	for i := 0; i < rawEdges.Len(); i++ {
		rawEdge := rawEdges.Index(i).Elem()
		node := rawEdge.FieldByName("Node")
		edge := reflect.New(edgeItemType)
		cursor := rawEdge.FieldByName("Cursor")
		edge.Elem().FieldByName("Node").Set(
			reflect.ValueOf(
				*params.Factory(node.Interface().(*R)),
			),
		)
		edge.Elem().FieldByName("Cursor").Set(cursor)
		edgeSlice = reflect.Append(edgeSlice, edge)
	}

	_ = reflection.SetField(obj, "Edges", edgeSlice.Interface())

	return obj, nil
}
