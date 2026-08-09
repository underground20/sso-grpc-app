package auth

import (
	"app/internal/auth"
	"app/internal/storage"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ToGRPCErr(err error, internalMessage string) (error, bool) {
	if internalMessage == "" {
		internalMessage = "internal server error"
	}

	switch {
	case errors.Is(err, auth.ErrRolesDoesNotExists),
		errors.Is(err, auth.ErrInvalidCredentials):
		return status.Error(codes.InvalidArgument, err.Error()), false

	case errors.Is(err, storage.ErrUserExists):
		return status.Error(codes.AlreadyExists, "user already registered"), false

	case errors.Is(err, storage.ErrUserNotFound),
		errors.Is(err, storage.ErrTokenNotFound),
		errors.Is(err, storage.ErrAppNotFound):
		return status.Error(codes.NotFound, err.Error()), false

	case errors.Is(err, auth.ErrRefreshTokenIsRevoked):
		return status.Error(codes.PermissionDenied, err.Error()), false

	default:
		return status.Error(codes.Internal, internalMessage), true
	}
}
