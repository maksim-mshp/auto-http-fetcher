package service

import (
	coreHttp "auto-http-fetcher/internal/core/http"
	"context"
)

func (u *UserService) Delete(ctx context.Context, userID int) error {

	if userID < 0 {
		return coreHttp.ErrInvalidUserID
	}

	if err := u.userRepo.Delete(ctx, userID); err != nil {
		return err
	}

	return nil
}
