package api

import "context"

type ctxKey int

const (
	ctxKeyAccountID ctxKey = iota + 1
	ctxKeyUserID
	ctxKeyUsername
)

func WithAuth(ctx context.Context, accountID, userID int64, username string) context.Context {
	ctx = context.WithValue(ctx, ctxKeyAccountID, accountID)
	ctx = context.WithValue(ctx, ctxKeyUserID, userID)
	ctx = context.WithValue(ctx, ctxKeyUsername, username)
	return ctx
}

func AccountIDFromContext(ctx context.Context) (int64, bool) {
	v, ok := ctx.Value(ctxKeyAccountID).(int64)
	return v, ok && v > 0
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	v, ok := ctx.Value(ctxKeyUserID).(int64)
	return v, ok && v > 0
}

func UsernameFromContext(ctx context.Context) string {
	v, _ := ctx.Value(ctxKeyUsername).(string)
	return v
}
