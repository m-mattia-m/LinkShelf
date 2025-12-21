package main

import (
	"backend/internal/infrastructure/api/model"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_API_Section_Create(t *testing.T) {

	shelfId, err := getShelfInclusiveItsOwnerUser()
	require.NoError(t, err)

	request := model.SectionBase{
		Title:   "section-title-creation",
		ShelfId: shelfId,
	}

	resp := doRequest(
		t,
		http.MethodPost,
		"/v1/sections",
		strings.NewReader(ObjectToJSON(request)),
	)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Errorf("Failed to close response body: %v", err)
		}
	}(resp.Body)

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var sectionResp model.Section
	err = json.Unmarshal(body, &sectionResp)
	require.NoError(t, err)

	require.Equal(t, request.Title, sectionResp.Title)
	require.Equal(t, request.ShelfId, shelfId)
}

func Test_API_Section_Update(t *testing.T) {
	shelfId, err := getShelfInclusiveItsOwnerUser()
	require.NoError(t, err)

	section, err := TestService.SectionService.Create(&model.Section{
		SectionBase: model.SectionBase{
			Title:   "test-section-update",
			ShelfId: shelfId,
		},
	})
	require.NoError(t, err)

	updateRequest := model.SectionBase{
		Title:   "section-title-updated",
		ShelfId: shelfId,
	}

	resp := doRequest(
		t,
		http.MethodPut,
		fmt.Sprintf("/v1/sections/%s", section.Id),
		strings.NewReader(ObjectToJSON(updateRequest)),
	)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Errorf("Failed to close response body: %v", err)
		}
	}(resp.Body)

	//require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var sectionResp model.Section
	err = json.Unmarshal(body, &sectionResp)
	require.NoError(t, err)

	require.Equal(t, updateRequest.Title, sectionResp.Title)
	require.Equal(t, updateRequest.ShelfId, shelfId)
}

func Test_API_Section_Delete(t *testing.T) {
	shelfId, err := getShelfInclusiveItsOwnerUser()
	require.NoError(t, err)

	section, err := TestService.SectionService.Create(&model.Section{
		SectionBase: model.SectionBase{
			Title:   "test-section-delete",
			ShelfId: shelfId,
		},
	})
	require.NoError(t, err)

	resp := doRequest(
		t,
		http.MethodDelete,
		fmt.Sprintf("/v1/sections/%s", section.Id),
		nil,
	)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Errorf("Failed to close response body: %v", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "", string(body))

	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func Test_API_Section_Get(t *testing.T) {
	shelfId, err := getShelfInclusiveItsOwnerUser()
	require.NoError(t, err)

	section, err := TestService.SectionService.Create(&model.Section{
		SectionBase: model.SectionBase{
			Title:   "test-section-get",
			ShelfId: shelfId,
		},
	})
	require.NoError(t, err)

	resp := doRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/v1/sections?shelfId=%s", shelfId),
		nil,
	)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Errorf("Failed to close response body: %v", err)
		}
	}(resp.Body)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var sectionResp []model.Section
	err = json.Unmarshal(body, &sectionResp)
	require.NoError(t, err)

	require.Equal(t, section.Id, sectionResp[0].Id)
	require.Equal(t, section.Title, sectionResp[0].Title)
	require.Equal(t, shelfId, sectionResp[0].ShelfId)
}
