//go:build integration
// +build integration

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

func Test_API_Link_Create(t *testing.T) {

	sectionId, err := getSectionAndShelfInclusiveItsOwnerUser()
	require.NoError(t, err)

	request := model.LinkBase{
		Title:     "link-title-creation",
		Link:      "https://link.example.com",
		Icon:      "",
		Color:     "",
		SectionId: sectionId,
	}

	resp := doRequest(
		t,
		http.MethodPost,
		"/v1/links",
		strings.NewReader(ObjectToJSON(request)),
	)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Errorf("Failed to close response body: %v", err)
		}
	}(resp.Body)

	//require.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var linkResp model.Link
	err = json.Unmarshal(body, &linkResp)
	require.NoError(t, err)

	require.Equal(t, request.Title, linkResp.Title)
	require.Equal(t, request.Link, linkResp.Link)
	require.Equal(t, request.Icon, linkResp.Icon)
	require.Equal(t, request.Color, linkResp.Color)
	require.Equal(t, request.SectionId, sectionId)

}

func Test_API_Link_Update(t *testing.T) {

	sectionId, err := getSectionAndShelfInclusiveItsOwnerUser()
	require.NoError(t, err)

	link, err := TestService.LinkService.Create(&model.Link{
		LinkBase: model.LinkBase{
			Title:     "link-title-to-update",
			Link:      "https://link-to-update.example.com",
			Icon:      "base-64-encoded-icon",
			Color:     "#FFFFFF",
			SectionId: sectionId,
		},
	})
	require.NoError(t, err)

	updateRequest := model.LinkBase{
		Title:     "link-title-updated",
		Link:      "https://link-updated.example.com",
		Icon:      "updated-icon",
		Color:     "#000000",
		SectionId: sectionId,
	}

	resp := doRequest(
		t,
		http.MethodPut,
		fmt.Sprintf("/v1/links/%s", link.Id),
		strings.NewReader(ObjectToJSON(updateRequest)),
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

	var linkResp model.Link
	err = json.Unmarshal(body, &linkResp)
	require.NoError(t, err)

	require.Equal(t, updateRequest.Title, linkResp.Title)
	require.Equal(t, updateRequest.Link, linkResp.Link)
	require.Equal(t, updateRequest.Icon, linkResp.Icon)
	require.Equal(t, updateRequest.Color, linkResp.Color)
	require.Equal(t, updateRequest.SectionId, sectionId)

}

func Test_API_Link_Delete(t *testing.T) {
	sectionId, err := getSectionAndShelfInclusiveItsOwnerUser()
	require.NoError(t, err)

	link, err := TestService.LinkService.Create(&model.Link{
		LinkBase: model.LinkBase{
			Title:     "link-title-to-update",
			Link:      "https://link-to-update.example.com",
			Icon:      "base-64-encoded-icon",
			Color:     "#FFFFFF",
			SectionId: sectionId,
		},
	})
	require.NoError(t, err)

	resp := doRequest(
		t,
		http.MethodDelete,
		fmt.Sprintf("/v1/links/%s", link.Id),
		nil,
	)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Errorf("Failed to close response body: %v", err)
		}
	}(resp.Body)

	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func Test_API_Link_List(t *testing.T) {
	sectionId, err := getSectionAndShelfInclusiveItsOwnerUser()
	require.NoError(t, err)

	_, err = TestService.LinkService.Create(&model.Link{
		LinkBase: model.LinkBase{
			Title:     "link-title-to-update",
			Link:      "https://link-to-update.example.com",
			Icon:      "base-64-encoded-icon",
			Color:     "#FFFFFF",
			SectionId: sectionId,
		},
	})
	require.NoError(t, err)

	resp := doRequest(
		t,
		http.MethodGet,
		"/v1/links?shelfId="+sectionId,
		nil,
	)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Errorf("Failed to close response body: %v", err)
		}
	}(resp.Body)

	//require.Equal(t, http.StatusNoContent, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var linksResp []model.Link
	err = json.Unmarshal(body, &linksResp)
	require.NoError(t, err)
}
