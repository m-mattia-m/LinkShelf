package domain

import (
	"errors"
	"testing"

	"backend/internal/config"
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/repository"
	"backend/internal/infrastructure/repository/mocks"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func setupBootstrapTestConfig(t *testing.T, email, password string) {
	t.Helper()
	config.Reset()
	config.Set("authentication.bootstrapAdmin.email", email)
	config.Set("authentication.bootstrapAdmin.password", password)
}

func Test_Unit_EnsureBootstrapAdmin_NoopWhenNotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepository := mocks.NewMockUserRepository(ctrl)
	repo := &repository.Repository{UserRepository: userRepository}

	setupBootstrapTestConfig(t, "", "")
	require.NoError(t, EnsureBootstrapAdmin(repo))

	setupBootstrapTestConfig(t, "admin@example.com", "")
	require.NoError(t, EnsureBootstrapAdmin(repo))

	setupBootstrapTestConfig(t, "", "some-password")
	require.NoError(t, EnsureBootstrapAdmin(repo))
}

func Test_Unit_EnsureBootstrapAdmin_CreatesNewAdmin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepository := mocks.NewMockUserRepository(ctrl)
	repo := &repository.Repository{UserRepository: userRepository}
	setupBootstrapTestConfig(t, "admin@example.com", "super-secret")

	userRepository.EXPECT().
		FindByEmail("admin@example.com").
		Return(nil, nil)

	userRepository.EXPECT().
		Create(gomock.Any(), gomock.Any(), model.RoleAdmin).
		DoAndReturn(func(u model.UserBase, hashedPassword, role string) (string, error) {
			require.Equal(t, "admin@example.com", u.Email)
			require.NotEmpty(t, hashedPassword)
			return "new-admin-id", nil
		})

	userRepository.EXPECT().
		SetPasswordAndRole("new-admin-id", gomock.Any(), model.RoleAdmin).
		Return(nil)

	userRepository.EXPECT().
		MarkVerified("new-admin-id").
		Return(nil)

	require.NoError(t, EnsureBootstrapAdmin(repo))
}

func Test_Unit_EnsureBootstrapAdmin_RefreshesExistingAdmin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepository := mocks.NewMockUserRepository(ctrl)
	repo := &repository.Repository{UserRepository: userRepository}
	setupBootstrapTestConfig(t, "admin@example.com", "super-secret")

	userRepository.EXPECT().
		FindByEmail("admin@example.com").
		Return(&repository.AuthRecord{Id: "existing-admin-id"}, nil)

	userRepository.EXPECT().
		SetPasswordAndRole("existing-admin-id", gomock.Any(), model.RoleAdmin).
		Return(nil)

	userRepository.EXPECT().
		MarkVerified("existing-admin-id").
		Return(nil)

	require.NoError(t, EnsureBootstrapAdmin(repo))
}

func Test_Unit_EnsureBootstrapAdmin_PropagatesFindByEmailError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepository := mocks.NewMockUserRepository(ctrl)
	repo := &repository.Repository{UserRepository: userRepository}
	setupBootstrapTestConfig(t, "admin@example.com", "super-secret")

	findErr := errors.New("db unavailable")
	userRepository.EXPECT().
		FindByEmail("admin@example.com").
		Return(nil, findErr)

	require.ErrorIs(t, EnsureBootstrapAdmin(repo), findErr)
}

func Test_Unit_EnsureBootstrapAdmin_PropagatesCreateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepository := mocks.NewMockUserRepository(ctrl)
	repo := &repository.Repository{UserRepository: userRepository}
	setupBootstrapTestConfig(t, "admin@example.com", "super-secret")

	createErr := errors.New("insert failed")
	userRepository.EXPECT().
		FindByEmail("admin@example.com").
		Return(nil, nil)
	userRepository.EXPECT().
		Create(gomock.Any(), gomock.Any(), model.RoleAdmin).
		Return("", createErr)

	require.ErrorIs(t, EnsureBootstrapAdmin(repo), createErr)
}
