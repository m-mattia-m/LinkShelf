package domain

import (
	"backend/internal/config"
	"backend/internal/infrastructure/api/model"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_Unit_NormalizeDomain(t *testing.T) {
	cases := map[string]string{
		"profile.example.com":             "profile.example.com",
		"  Profile.Example.COM  ":         "profile.example.com",
		"profile.example.com.":            "profile.example.com",
		"profile.example.com/":            "profile.example.com",
		"profile.example.com./":           "profile.example.com",
		"profile.example.com:443":         "profile.example.com",
		"profile.example.com:80":          "profile.example.com",
		"profile.example.com.:443/":       "profile.example.com",
		"profile.example.com:9443":        "profile.example.com:9443",
		"PROFILE.example.com:9443/":       "profile.example.com:9443",
		"profile.example.com:0443":        "profile.example.com:0443",
		"profile.example.com:":            "profile.example.com:",
		"https://profile.example.com":     "https://profile.example.com",
		"https://profile.example.com/a/b": "https://profile.example.com/a/b",
		"":                                "",
		"   ":                             "",
	}

	for input, want := range cases {
		require.Equal(t, want, NormalizeDomain(input), "input %q", input)
	}
}

func Test_Unit_NormalizeDomain_IsIdempotent(t *testing.T) {
	for _, input := range []string{"A.b.COM.:443/", "x.y.io:9443", "junk", "profile.example.com:0443"} {
		once := NormalizeDomain(input)
		require.Equal(t, once, NormalizeDomain(once), "input %q", input)
	}
}

func Test_Unit_ValidateDomain_Accepts(t *testing.T) {
	valid := []string{
		"example.com",
		"profile.example.com",
		"a.b.c.d.example.co.uk",
		"profile.example.com:9443",
		"profile.example.com:1",
		"profile.example.com:65535",
		"my-profile.example.com",
		"123.example.com",
		"x1.example.io",
		"profile.example.xn--p1ai",
		"xn--bcher-kva.example.com",
		strings.Repeat("a", 63) + ".example.com",
	}

	for _, domain := range valid {
		require.NoError(t, ValidateDomain(domain), domain)
	}
}

func Test_Unit_ValidateDomain_Rejects(t *testing.T) {
	invalid := map[string]string{
		"localhost":                              "single label",
		"example":                                "single label",
		"1.2.3.4":                                "IPv4 address",
		"192.168.0.1:8080":                       "IPv4 address with port",
		"*.example.com":                          "wildcard",
		"profile.*.com":                          "wildcard",
		"https://profile.example.com":            "scheme",
		"profile.example.com/path":               "path",
		"profile.example.com:0":                  "port zero",
		"profile.example.com:65536":              "port too big",
		"profile.example.com:0443":               "port with a leading zero",
		"profile.example.com:abc":                "port not a number",
		"profile.example.com:":                   "empty port",
		"profile.example.com:1:2":                "two ports",
		"profile_x.example.com":                  "underscore",
		"-profile.example.com":                   "leading hyphen",
		"profile-.example.com":                   "trailing hyphen",
		"profile..example.com":                   "empty label",
		".example.com":                           "leading dot",
		"profile.example.c":                      "one-letter TLD",
		"profile.example.c0m":                    "TLD with a digit",
		"profile.example.123":                    "numeric TLD",
		"profile.exämple.com":                    "unicode label",
		"profile example.com":                    "whitespace",
		strings.Repeat("a", 64) + ".example.com": "label over 63 characters",
		strings.Repeat("a.", 130) + "com":        "name over 253 characters",
		"":                                       "empty",
	}

	for domain, why := range invalid {
		err := ValidateDomain(domain)
		require.ErrorIs(t, err, ErrInvalidInput, "%q should be rejected: %s", domain, why)
	}
}

func Test_Unit_CheckShelfDomain_ReservesTheInstanceHosts(t *testing.T) {
	config.Reset()
	t.Cleanup(config.Reset)
	config.Set("app.frontendUrl", "https://linkshelf.example.com")
	config.Set("server.host", "api.example.com")
	config.Set("server.port", 8085)

	for _, reserved := range []string{
		"linkshelf.example.com",
		"LinkShelf.Example.com.",
		"linkshelf.example.com:443",
		"api.example.com",
		"api.example.com:8085",
	} {
		_, err := checkShelfDomain(reserved)
		require.ErrorIs(t, err, ErrInvalidInput, reserved)
		require.ErrorContains(t, err, "belongs to this instance", reserved)
	}

	for _, free := range []string{"profile.example.com", "linkshelf.example.com:9443", "example.com"} {
		got, err := checkShelfDomain(free)
		require.NoError(t, err, free)
		require.Equal(t, NormalizeDomain(free), got)
	}
}

func Test_Unit_CheckShelfDomain_ReservesTheFrontendPortToo(t *testing.T) {
	config.Reset()
	t.Cleanup(config.Reset)
	config.Set("app.frontendUrl", "https://linkshelf.example.com:9443")

	_, err := checkShelfDomain("linkshelf.example.com:9443")
	require.ErrorIs(t, err, ErrInvalidInput)
	_, err = checkShelfDomain("linkshelf.example.com")
	require.ErrorIs(t, err, ErrInvalidInput)
}

func Test_Unit_CheckShelfDomain_NothingReservedWhenNothingIsConfigured(t *testing.T) {
	config.Reset()
	t.Cleanup(config.Reset)

	got, err := checkShelfDomain("Profile.example.com")
	require.NoError(t, err)
	require.Equal(t, "profile.example.com", got)
}

func Test_Unit_CheckShelfDomain_EmptyStaysEmpty(t *testing.T) {
	got, err := checkShelfDomain("  ")
	require.NoError(t, err)
	require.Empty(t, got)
}

func Test_Unit_Shelf_Creation_DomainRules(t *testing.T) {
	t.Run("a free domain is stored normalized", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		userBasedPaths(t, false)

		svc.UserRepository.EXPECT().Get("user-1").Return(&model.User{Id: "user-1"}, nil)
		svc.ShelfRepository.EXPECT().DomainInUse("profile.example.com", "").Return(false, nil)
		svc.ShelfRepository.EXPECT().Create(gomock.Any()).DoAndReturn(func(shelf *model.Shelf) (string, error) {
			require.Equal(t, "profile.example.com", shelf.Domain)
			require.Empty(t, shelf.Path)
			return "shelf-1", nil
		})

		id, err := svc.Service.ShelfService.Create("user-1", shelfWithDomain("  Profile.Example.com.:443/ "))

		require.NoError(t, err)
		require.Equal(t, "shelf-1", id)
	})

	t.Run("a taken domain is a conflict", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		userBasedPaths(t, false)

		svc.UserRepository.EXPECT().Get("user-1").Return(&model.User{Id: "user-1"}, nil)
		svc.ShelfRepository.EXPECT().DomainInUse("profile.example.com", "").Return(true, nil)

		id, err := svc.Service.ShelfService.Create("user-1", shelfWithDomain("profile.example.com"))

		require.ErrorIs(t, err, ErrConflict)
		require.Empty(t, id)
	})

	t.Run("an invalid domain is rejected before the database is asked", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		userBasedPaths(t, false)

		svc.UserRepository.EXPECT().Get("user-1").Return(&model.User{Id: "user-1"}, nil)

		id, err := svc.Service.ShelfService.Create("user-1", shelfWithDomain("localhost"))

		require.ErrorIs(t, err, ErrInvalidInput)
		require.Empty(t, id)
	})

	t.Run("one of this instance's own hosts is rejected", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		userBasedPaths(t, false)
		config.Set("app.frontendUrl", "https://linkshelf.example.com")

		svc.UserRepository.EXPECT().Get("user-1").Return(&model.User{Id: "user-1"}, nil)

		_, err := svc.Service.ShelfService.Create("user-1", shelfWithDomain("linkshelf.example.com"))

		require.ErrorIs(t, err, ErrInvalidInput)
		require.ErrorContains(t, err, "belongs to this instance")
	})

	t.Run("a shelf can't have a path and a domain", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		userBasedPaths(t, false)

		svc.UserRepository.EXPECT().Get("user-1").Return(&model.User{Id: "user-1"}, nil)

		shelf := shelfWithPath("my-path")
		shelf.Domain = "profile.example.com"
		id, err := svc.Service.ShelfService.Create("user-1", shelf)

		require.ErrorIs(t, err, ErrInvalidInput)
		require.ErrorContains(t, err, "not both")
		require.Empty(t, id)
	})

	t.Run("a failing lookup is returned", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		userBasedPaths(t, false)

		svc.UserRepository.EXPECT().Get("user-1").Return(&model.User{Id: "user-1"}, nil)
		svc.ShelfRepository.EXPECT().DomainInUse("profile.example.com", "").Return(false, errors.New("db unavailable"))

		_, err := svc.Service.ShelfService.Create("user-1", shelfWithDomain("profile.example.com"))

		require.ErrorContains(t, err, "db unavailable")
	})
}

func Test_Unit_Shelf_Update_DomainRules(t *testing.T) {
	stored := func(path, domain string) *model.Shelf {
		return &model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-1", Path: path, Title: "shelf-title-test"}, Domain: domain, UserId: "user-1"}
	}
	expectLookups := func(svc *MockService, existing *model.Shelf) {
		svc.UserRepository.EXPECT().Get("user-1").Return(&model.User{Id: "user-1"}, nil)
		svc.ShelfRepository.EXPECT().Get("shelf-1").Return(existing, nil)
	}
	expectSaved := func(svc *MockService, saved *model.Shelf) {
		svc.ShelfRepository.EXPECT().Update(gomock.Any()).Return(nil)
		svc.ShelfRepository.EXPECT().Get("shelf-1").Return(saved, nil)
	}

	t.Run("switching from a path to a domain checks only the domain", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		userBasedPaths(t, false)

		expectLookups(svc, stored("old-path", ""))
		svc.ShelfRepository.EXPECT().DomainInUse("profile.example.com", "shelf-1").Return(false, nil)
		expectSaved(svc, stored("", "profile.example.com"))

		shelf, err := svc.Service.ShelfService.Update("shelf-1", "user-1", false, shelfWithDomain("Profile.example.com"))

		require.NoError(t, err)
		require.Equal(t, "profile.example.com", shelf.Domain)
	})

	t.Run("switching from a domain to a path checks only the path", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		userBasedPaths(t, false)

		expectLookups(svc, stored("", "profile.example.com"))
		svc.ShelfRepository.EXPECT().PathInUse("new-path", "shelf-1").Return(false, nil)
		expectSaved(svc, stored("new-path", ""))

		_, err := svc.Service.ShelfService.Update("shelf-1", "user-1", false, shelfWithPath("new-path"))

		require.NoError(t, err)
	})

	t.Run("a changed domain that another shelf has is a conflict", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		userBasedPaths(t, false)

		expectLookups(svc, stored("", "old.example.com"))
		svc.ShelfRepository.EXPECT().DomainInUse("new.example.com", "shelf-1").Return(true, nil)

		shelf, err := svc.Service.ShelfService.Update("shelf-1", "user-1", false, shelfWithDomain("new.example.com"))

		require.ErrorIs(t, err, ErrConflict)
		require.Nil(t, shelf)
	})

	t.Run("adding a domain to a shelf that has a path is rejected", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		userBasedPaths(t, false)

		expectLookups(svc, stored("my-path", ""))

		request := shelfWithPath("my-path")
		request.Domain = "profile.example.com"
		shelf, err := svc.Service.ShelfService.Update("shelf-1", "user-1", false, request)

		require.ErrorIs(t, err, ErrInvalidInput)
		require.ErrorContains(t, err, "not both")
		require.Nil(t, shelf)
	})

	t.Run("clearing both is rejected", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		userBasedPaths(t, false)

		expectLookups(svc, stored("my-path", ""))

		_, err := svc.Service.ShelfService.Update("shelf-1", "user-1", false, shelfWithPath(""))

		require.ErrorIs(t, err, ErrInvalidInput)
		require.ErrorContains(t, err, "path or a domain")
	})

	t.Run("a shelf that has both from before can still have its title edited", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		userBasedPaths(t, false)

		expectLookups(svc, stored("my-path", "profile.example.com"))
		// No PathInUse or DomainInUse expectation: nothing about the location changed.
		expectSaved(svc, stored("my-path", "profile.example.com"))

		request := shelfWithPath("my-path")
		request.Domain = "profile.example.com"
		_, err := svc.Service.ShelfService.Update("shelf-1", "user-1", false, request)

		require.NoError(t, err)
	})

	t.Run("a shelf with a stored domain in another case is normalized when edited", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		userBasedPaths(t, false)

		expectLookups(svc, stored("", "Profile.Example.com"))
		svc.ShelfRepository.EXPECT().DomainInUse("profile.example.com", "shelf-1").Return(false, nil)
		expectSaved(svc, stored("", "profile.example.com"))

		_, err := svc.Service.ShelfService.Update("shelf-1", "user-1", false, shelfWithDomain("Profile.Example.com"))

		require.NoError(t, err)
	})

	t.Run("a stored domain that was never valid doesn't block unrelated edits", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		userBasedPaths(t, false)

		expectLookups(svc, stored("", "not a domain"))
		expectSaved(svc, stored("", "not a domain"))

		_, err := svc.Service.ShelfService.Update("shelf-1", "user-1", false, shelfWithDomain("not a domain"))

		require.NoError(t, err)
	})
}

func Test_Unit_Shelf_GetByDomain(t *testing.T) {
	t.Run("resolves a shelf and its theme", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		userBasedPaths(t, true) // a domain doesn't depend on the setting

		svc.ShelfRepository.EXPECT().GetByDomain("profile.example.com:9443").Return(&model.Shelf{
			PublicShelf: model.PublicShelf{Id: "shelf-1", Title: "Profile"},
		}, nil)

		shelf, err := svc.Service.ShelfService.GetByDomain(" Profile.Example.com:9443/ ")

		require.NoError(t, err)
		require.Equal(t, "shelf-1", shelf.Id)
	})

	t.Run("a value that can't be a domain is not found without a query", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()

		for _, host := range []string{"localhost", "1.2.3.4", "", "https://x.example.com"} {
			shelf, err := svc.Service.ShelfService.GetByDomain(host)
			require.NoError(t, err, host)
			require.Nil(t, shelf, host)
		}
	})

	t.Run("an unknown domain is not found", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()

		svc.ShelfRepository.EXPECT().GetByDomain("nobody.example.com").Return(nil, nil)

		shelf, err := svc.Service.ShelfService.GetByDomain("nobody.example.com")

		require.NoError(t, err)
		require.Nil(t, shelf)
	})

	t.Run("a failing lookup is returned", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()

		svc.ShelfRepository.EXPECT().GetByDomain("profile.example.com").Return(nil, errors.New("db unavailable"))

		_, err := svc.Service.ShelfService.GetByDomain("profile.example.com")

		require.ErrorContains(t, err, "db unavailable")
	})
}

func Test_Unit_Shelf_IsDomainRegistered(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ShelfRepository.EXPECT().DomainInUse("profile.example.com", "").Return(true, nil)
	registered, err := svc.Service.ShelfService.IsDomainRegistered("Profile.example.com")
	require.NoError(t, err)
	require.True(t, registered)

	// Not a domain: answered without asking the database.
	registered, err = svc.Service.ShelfService.IsDomainRegistered("localhost")
	require.NoError(t, err)
	require.False(t, registered)

	svc.ShelfRepository.EXPECT().DomainInUse("broken.example.com", "").Return(false, errors.New("db unavailable"))
	_, err = svc.Service.ShelfService.IsDomainRegistered("broken.example.com")
	require.ErrorContains(t, err, "db unavailable")
}
