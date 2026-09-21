package domain

import (
	"backend/internal/config"
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/repository"
	"backend/internal/infrastructure/repository/mocks"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_Unit_BackfillUsernames_DerivesFromEmailAndAvoidsCollisions(t *testing.T) {
	config.Reset()
	t.Cleanup(config.Reset)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepository := mocks.NewMockUserRepository(ctrl)
	repo := &repository.Repository{UserRepository: userRepository}

	taken := map[string]bool{"taken-name": true}
	assigned := map[string]string{}

	userRepository.EXPECT().ListWithoutUsername().Return([]model.User{
		{Id: "u1", UserBase: model.UserBase{Email: "john.smith@one.example"}},
		{Id: "u2", UserBase: model.UserBase{Email: "John.Smith@two.example"}},
		{Id: "u3", UserBase: model.UserBase{Email: "taken-name@three.example"}},
		{Id: "u4", UserBase: model.UserBase{Email: "admin@four.example"}},
	}, nil)
	userRepository.EXPECT().
		UsernameTaken(gomock.Any(), "").
		DoAndReturn(func(username, _ string) (bool, error) { return taken[username], nil }).
		AnyTimes()
	userRepository.EXPECT().
		SetUsername(gomock.Any(), gomock.Any()).
		DoAndReturn(func(userId, username string) error {
			taken[username] = true
			assigned[userId] = username
			return nil
		}).
		Times(4)

	require.NoError(t, BackfillUsernames(repo))

	require.Equal(t, map[string]string{
		"u1": "john-smith",
		"u2": "john-smith-2",
		"u3": "taken-name-2",
		// "admin" is reserved, so the derived name moves off it.
		"u4": "admin-2",
	}, assigned)
}

func Test_Unit_BackfillUsernames_NothingToDo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepository := mocks.NewMockUserRepository(ctrl)
	userRepository.EXPECT().ListWithoutUsername().Return(nil, nil)

	require.NoError(t, BackfillUsernames(&repository.Repository{UserRepository: userRepository}))
}

func Test_Unit_BackfillUsernames_PropagatesErrors(t *testing.T) {
	config.Reset()
	t.Cleanup(config.Reset)
	boom := errors.New("db unavailable")

	t.Run("list", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		userRepository := mocks.NewMockUserRepository(ctrl)
		userRepository.EXPECT().ListWithoutUsername().Return(nil, boom)

		require.ErrorIs(t, BackfillUsernames(&repository.Repository{UserRepository: userRepository}), boom)
	})

	t.Run("lookup", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		userRepository := mocks.NewMockUserRepository(ctrl)
		userRepository.EXPECT().ListWithoutUsername().Return([]model.User{{Id: "u1", UserBase: model.UserBase{Email: "a@b.c"}}}, nil)
		userRepository.EXPECT().UsernameTaken(gomock.Any(), "").Return(false, boom)

		require.ErrorIs(t, BackfillUsernames(&repository.Repository{UserRepository: userRepository}), boom)
	})

	t.Run("save", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		userRepository := mocks.NewMockUserRepository(ctrl)
		userRepository.EXPECT().ListWithoutUsername().Return([]model.User{{Id: "u1", UserBase: model.UserBase{Email: "alice@b.c"}}}, nil)
		userRepository.EXPECT().UsernameTaken("alice", "").Return(false, nil)
		userRepository.EXPECT().SetUsername("u1", "alice").Return(boom)

		require.ErrorIs(t, BackfillUsernames(&repository.Repository{UserRepository: userRepository}), boom)
	})
}

func Test_Unit_EnsureUniquePaths_SkippedWhileUserBasedPathsIsOn(t *testing.T) {
	config.Reset()
	t.Cleanup(config.Reset)
	config.Set("app.userBasedPaths", true)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// No expectations: with the feature on, the repository must not be asked.
	shelfRepository := mocks.NewMockShelfRepository(ctrl)

	require.NoError(t, EnsureUniquePaths(&repository.Repository{ShelfRepository: shelfRepository}))
}

func Test_Unit_EnsureUniquePaths_PassesWithoutCollisions(t *testing.T) {
	config.Reset()
	t.Cleanup(config.Reset)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shelfRepository := mocks.NewMockShelfRepository(ctrl)
	shelfRepository.EXPECT().ListPathCollisions().Return(nil, nil)

	require.NoError(t, EnsureUniquePaths(&repository.Repository{ShelfRepository: shelfRepository}))
}

func Test_Unit_EnsureUniquePaths_ListsEveryCollidingShelfAndItsOwner(t *testing.T) {
	config.Reset()
	t.Cleanup(config.Reset)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shelfRepository := mocks.NewMockShelfRepository(ctrl)
	shelfRepository.EXPECT().ListPathCollisions().Return([]repository.PathCollision{
		{Path: "profile", ShelfId: "shelf-1", UserId: "user-1", Username: "alice"},
		{Path: "profile", ShelfId: "shelf-2", UserId: "user-2", Username: "bob"},
	}, nil)

	err := EnsureUniquePaths(&repository.Repository{ShelfRepository: shelfRepository})

	require.Error(t, err)
	require.ErrorContains(t, err, "app.userBasedPaths is false")
	require.ErrorContains(t, err, "/profile  shelf shelf-1  owner alice (user-1)")
	require.ErrorContains(t, err, "/profile  shelf shelf-2  owner bob (user-2)")
	require.ErrorContains(t, err, "set app.userBasedPaths back to true")
}

func Test_Unit_EnsureUniquePaths_PropagatesRepositoryError(t *testing.T) {
	config.Reset()
	t.Cleanup(config.Reset)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	boom := errors.New("db unavailable")
	shelfRepository := mocks.NewMockShelfRepository(ctrl)
	shelfRepository.EXPECT().ListPathCollisions().Return(nil, boom)

	require.ErrorIs(t, EnsureUniquePaths(&repository.Repository{ShelfRepository: shelfRepository}), boom)
}
