package main

import "context"

type logContext string

func withLoggedValue(ctx context.Context, key string, val any) context.Context {
	return context.WithValue(ctx, logContext(key), val)
}
