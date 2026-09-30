//go:build integration
// +build integration

package integrationtests

import (
	"backend/internal/infrastructure/api/model"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// End-to-end regressions for the findings of the security review, run
// against the real HTTP API and a real Postgres.

func Test_API_Security_Link_JavascriptPayloadIsRejected(t *testing.T) {
	sectionId, token := getSectionAndShelfInclusiveItsOwnerUser(t)

	for _, payload := range []string{
		"javascript:alert(document.domain)%2F%2F@example.com",
		"https://user:pass@example.com",
	} {
		resp := doAuthedRequest(t, http.MethodPost, "/v1/links",
			strings.NewReader(ObjectToJSON(model.LinkBase{Title: "xss", Link: payload, SectionId: sectionId})), token)
		_ = resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode, payload)
	}
}

func Test_API_Security_Link_StoresTheNormalizedUrl(t *testing.T) {
	sectionId, token := getSectionAndShelfInclusiveItsOwnerUser(t)

	resp := doAuthedRequest(t, http.MethodPost, "/v1/links",
		strings.NewReader(ObjectToJSON(model.LinkBase{Title: "no scheme", Link: "example.com/path", SectionId: sectionId})), token)
	defer func(Body io.ReadCloser) { _ = Body.Close() }(resp.Body)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var link model.Link
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(body, &link))
	require.Equal(t, "https://example.com/path", link.Link)

	stored, err := TestRepository.LinkRepository.Get(link.Id)
	require.NoError(t, err)
	require.Equal(t, "https://example.com/path", stored.Link)
}

// Without email verification (config.test.yaml) an email change is applied
// directly, but the account loses its verified status - so a first OIDC login
// with that address can no longer be linked into it.
func Test_API_Security_EmailChange_DropsTheVerifiedStatus(t *testing.T) {
	suffix := uuid.NewString()[:8]
	user := &model.UserCreate{
		UserBase: model.UserBase{
			Username:  uniqueUsername(),
			Email:     "security-email-" + suffix + "@test.com",
			FirstName: "Security",
			LastName:  "Email",
		},
		Password: "secret",
	}
	created, err := TestService.UserService.Create(user, false)
	require.NoError(t, err)
	require.NoError(t, TestRepository.UserRepository.MarkVerified(created.Id))

	token := loginAndGetToken(t, user.Email, user.Password)

	victimEmail := "Victim-" + suffix + "@Corp.com"
	resp := doAuthedRequest(t, http.MethodPut, fmt.Sprintf("/v1/users/%s", created.Id),
		strings.NewReader(ObjectToJSON(model.UserBase{
			Username:  user.Username,
			Email:     victimEmail,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		})), token)
	_ = resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	record, err := TestRepository.UserRepository.FindByEmail(victimEmail)
	require.NoError(t, err)
	require.NotNil(t, record)
	require.Equal(t, created.Id, record.Id)
	require.Equal(t, strings.ToLower(victimEmail), record.Email)
	require.False(t, record.EmailVerified)
}

func Test_API_Security_EmailChange_ToAnotherAccountsEmailIsAConflict(t *testing.T) {
	suffix := uuid.NewString()[:8]
	other := &model.UserCreate{
		UserBase: model.UserBase{Username: uniqueUsername(), Email: "security-other-" + suffix + "@test.com", FirstName: "O", LastName: "O"},
		Password: "secret",
	}
	_, err := TestService.UserService.Create(other, false)
	require.NoError(t, err)

	user := &model.UserCreate{
		UserBase: model.UserBase{Username: uniqueUsername(), Email: "security-self-" + suffix + "@test.com", FirstName: "S", LastName: "S"},
		Password: "secret",
	}
	created, err := TestService.UserService.Create(user, false)
	require.NoError(t, err)
	token := loginAndGetToken(t, user.Email, user.Password)

	resp := doAuthedRequest(t, http.MethodPut, fmt.Sprintf("/v1/users/%s", created.Id),
		strings.NewReader(ObjectToJSON(model.UserBase{
			Username:  user.Username,
			Email:     strings.ToUpper(other.Email),
			FirstName: user.FirstName,
			LastName:  user.LastName,
		})), token)
	_ = resp.Body.Close()

	require.Equal(t, http.StatusConflict, resp.StatusCode)
}

// Two spellings of one address must never be two accounts.
func Test_API_Security_Registration_EmailIsCaseInsensitive(t *testing.T) {
	email := "security-case-" + uuid.NewString()[:8] + "@test.com"

	first := doRequest(t, http.MethodPost, "/v1/users", strings.NewReader(ObjectToJSON(model.UserCreate{
		UserBase: model.UserBase{Username: uniqueUsername(), Email: email, FirstName: "A", LastName: "A"},
		Password: "secret",
	})))
	_ = first.Body.Close()
	require.Equal(t, http.StatusCreated, first.StatusCode)

	second := doRequest(t, http.MethodPost, "/v1/users", strings.NewReader(ObjectToJSON(model.UserCreate{
		UserBase: model.UserBase{Username: uniqueUsername(), Email: strings.ToUpper(email), FirstName: "B", LastName: "B"},
		Password: "secret",
	})))
	_ = second.Body.Close()
	require.Equal(t, http.StatusConflict, second.StatusCode)
}
