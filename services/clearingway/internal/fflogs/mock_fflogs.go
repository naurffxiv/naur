package fflogs

import (
	"context"

	"github.com/hasura/go-graphql-client"
)

type MockGraphQLClient struct {
	QueryFunc func(ctx context.Context, q interface{}, variables map[string]interface{}, opts ...graphql.Option) error
}

func (m *MockGraphQLClient) Query(ctx context.Context, q interface{}, variables map[string]interface{}, opts ...graphql.Option) error {
	if m.QueryFunc != nil {
		return m.QueryFunc(ctx, q, variables, opts...)
	}
	return nil
}
