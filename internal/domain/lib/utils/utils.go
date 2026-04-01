package utils

import (
	"context"
	"fmt"
	"strings"
)

// NormalizeString removes spaces and leads to lowercase
func NormalizeString(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// ExtractUserIdFromCtx returns user_id from context
func ExtractUserIdFromCtx(ctx context.Context) (string, error) {
	userId, ok := ctx.Value("user_id").(string)
	if !ok {
		return "", fmt.Errorf("user id not found in context")
	}

	return userId, nil
}
