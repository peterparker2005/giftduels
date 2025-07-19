package proto

import (
	domain "github.com/peterparker2005/giftduels/apps/service-identity/internal/domain/user"
	"github.com/peterparker2005/giftduels/apps/service-identity/pkg/telegram"
	identityv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/identity/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToPBUser(u *domain.User) *identityv1.User {
	return &identityv1.User{
		UserId:          &sharedv1.UserId{Value: u.ID},
		TelegramId:      &sharedv1.TelegramUserId{Value: u.TelegramID},
		Username:        u.Username,
		FirstName:       u.FirstName,
		LastName:        u.LastName,
		PhotoUrl:        u.PhotoUrl,
		LanguageCode:    u.LanguageCode,
		AllowsWriteToPm: u.AllowsWriteToPm,
		IsPremium:       u.IsPremium,
		CreatedAt:       timestamppb.New(u.CreatedAt),
		UpdatedAt:       timestamppb.New(u.UpdatedAt),
	}
}

func ToPBProfile(u *domain.User) *identityv1.UserProfile {
	return &identityv1.UserProfile{
		UserId:     &sharedv1.UserId{Value: u.ID},
		TelegramId: &sharedv1.TelegramUserId{Value: u.TelegramID},
		Username:   u.Username,
		DisplayName: telegram.GetDisplayName(
			u.FirstName, u.LastName, u.Username),
		PhotoUrl:  u.PhotoUrl,
		IsPremium: u.IsPremium,
	}
}
