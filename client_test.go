package lovable_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/psyb0t/go-lovable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAPIKey      = "EXAMPLE-DO-NOT-USE"
	testProjectID   = "project-example"
	testWorkspaceID = "workspace-example"
	testProject     = `{
		"id":"project-example", "workspace_id":"workspace-example",
		"owner_id":null, "name":"Example", "description":null,
		"visibility":null, "latest_screenshot_url":null,
		"is_published":false, "publish_audience":null,
		"created_at":null, "last_edited_at":null,
		"latest_commit_sha":null, "preview_url":null,
		"publish_audience_targets":[]
	}`
)

func newTestClient(
	t *testing.T,
	handler http.HandlerFunc,
	opts ...lovable.ClientOption,
) *lovable.ClientWithResponses {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	auth := lovable.WithRequestEditorFn(
		func(_ context.Context, req *http.Request) error {
			req.Header.Set("Lovable-API-Key", testAPIKey)

			return nil
		},
	)

	opts = append([]lovable.ClientOption{auth}, opts...)
	client, err := lovable.NewClientWithResponses(srv.URL, opts...)
	require.NoError(t, err)

	return client
}

func writeJSON(t *testing.T, w http.ResponseWriter, body string) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write([]byte(body))
	assert.NoError(t, err)
}

func TestListProjectsPagination(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v1/projects", r.URL.Path)
		assert.Equal(t, testAPIKey, r.Header.Get("Lovable-API-Key"))
		assert.Equal(t, []string{testWorkspaceID},
			r.URL.Query()["workspace_id"])
		assert.Equal(t, "1", r.URL.Query().Get("limit"))

		if r.URL.Query().Get("cursor") == "" {
			writeJSON(t, w, `{
				"data":[{"id":"project-first",
				"workspace_id":"workspace-example", "name":null,
				"owner_id":null, "description":null, "visibility":null,
				"latest_screenshot_url":null, "is_published":false,
				"publish_audience":null, "created_at":null,
				"last_edited_at":null}],
				"pagination":{"has_more":true,"next_cursor":"next+/="}
			}`)

			return
		}

		assert.Equal(t, "next+/=", r.URL.Query().Get("cursor"))
		writeJSON(t, w, `{
			"data":[{"id":"project-last",
			"workspace_id":"workspace-example", "name":"Last",
			"owner_id":null, "description":null,
			"visibility":"workspace_edit",
			"latest_screenshot_url":null, "is_published":true,
			"publish_audience":"public", "created_at":null,
			"last_edited_at":null}],
			"pagination":{"has_more":false,"next_cursor":null}
		}`)
	})

	limit := int64(1)
	params := &lovable.ListProjectsParams{
		WorkspaceId: []string{testWorkspaceID},
		Limit:       &limit,
	}

	first, err := client.ListProjectsWithResponse(t.Context(), params)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, first.StatusCode())
	require.NotNil(t, first.JSON200)
	require.Len(t, first.JSON200.Data, 1)
	assert.Equal(t, "project-first", first.JSON200.Data[0].Id)
	assert.True(t, first.JSON200.Data[0].Visibility.IsNull())
	assert.True(t, first.JSON200.Data[0].Name.IsNull())
	assert.True(t, first.JSON200.Pagination.HasMore)
	cursor, err := first.JSON200.Pagination.NextCursor.Get()
	require.NoError(t, err)

	params.Cursor = &cursor
	last, err := client.ListProjectsWithResponse(t.Context(), params)
	require.NoError(t, err)
	require.NotNil(t, last.JSON200)
	require.Len(t, last.JSON200.Data, 1)
	assert.Equal(t, "project-last", last.JSON200.Data[0].Id)
	visibility, err := last.JSON200.Data[0].Visibility.Get()
	require.NoError(t, err)
	assert.Equal(t, "workspace_edit", string(visibility))
	assert.False(t, last.JSON200.Pagination.HasMore)
	assert.True(t, last.JSON200.Pagination.NextCursor.IsNull())
}

func TestGetProjectBearerAuth(t *testing.T) {
	t.Parallel()

	bearer := lovable.WithRequestEditorFn(
		func(_ context.Context, req *http.Request) error {
			req.Header.Del("Lovable-API-Key")
			req.Header.Set("Authorization", "Bearer "+testAPIKey)

			return nil
		},
	)
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v1/projects/"+testProjectID, r.URL.Path)
		assert.Empty(t, r.Header.Get("Lovable-API-Key"))
		assert.Equal(t, "Bearer "+testAPIKey,
			r.Header.Get("Authorization"))
		writeJSON(t, w, testProject)
	}, bearer)

	res, err := client.GetProjectWithResponse(
		t.Context(), testProjectID, nil,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode())
	require.NotNil(t, res.JSON200)
	assert.Equal(t, testProjectID, res.JSON200.Id)
	name, err := res.JSON200.Name.Get()
	require.NoError(t, err)
	assert.Equal(t, "Example", name)
	assert.True(t, res.JSON200.Description.IsNull())
	assert.True(t, res.JSON200.Visibility.IsNull())
	assert.False(t, res.JSON200.IsPublished)
}

func TestUpdateProject(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, "/v1/projects/"+testProjectID, r.URL.Path)
		assert.Equal(t, testAPIKey, r.Header.Get("Lovable-API-Key"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var body lovable.UpdateProjectRequest

		assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "Example", *body.Name)
		assert.Nil(t, body.Visibility)
		writeJSON(t, w, testProject)
	})

	name := "Example"
	res, err := client.UpdateProjectWithResponse(
		t.Context(), testProjectID, nil,
		lovable.UpdateProjectJSONRequestBody{Name: &name},
	)
	require.NoError(t, err)
	require.NotNil(t, res.JSON200)
	assert.Equal(t, testProjectID, res.JSON200.Id)
	gotName, err := res.JSON200.Name.Get()
	require.NoError(t, err)
	assert.Equal(t, name, gotName)
}

func TestGetProjectCanceled(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("canceled request reached the server")
	})
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := client.GetProjectWithResponse(ctx, testProjectID, nil)
	require.ErrorIs(t, err, context.Canceled)
}

func TestGetProjectMalformedJSON(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, `{"id":`)
	})

	_, err := client.GetProjectWithResponse(t.Context(), testProjectID, nil)

	var syntaxError *json.SyntaxError

	require.ErrorAs(t, err, &syntaxError)
}

func TestNullableRoundTrip(t *testing.T) {
	t.Parallel()

	var project lovable.Project

	require.NoError(t, json.Unmarshal([]byte(testProject), &project))
	assert.True(t, project.Visibility.IsNull())
	assert.True(t, project.Description.IsNull())
	data, err := json.Marshal(project)
	require.NoError(t, err)
	assert.JSONEq(t, testProject, string(data))
	assert.False(t, lovable.ProjectVisibility("<nil>").Valid())
}

func TestGetProjectErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		status int
		pick   func(*lovable.GetProjectResponse) *lovable.ErrorResponse
	}{
		{
			"invalid request", http.StatusBadRequest,
			func(r *lovable.GetProjectResponse) *lovable.ErrorResponse {
				return r.JSON400
			},
		},
		{
			"missing key", http.StatusUnauthorized,
			func(r *lovable.GetProjectResponse) *lovable.ErrorResponse {
				return r.JSON401
			},
		},
		{
			"insufficient scope", http.StatusForbidden,
			func(r *lovable.GetProjectResponse) *lovable.ErrorResponse {
				return r.JSON403
			},
		},
		{
			"wrong workspace", http.StatusNotFound,
			func(r *lovable.GetProjectResponse) *lovable.ErrorResponse {
				return r.JSON404
			},
		},
		{
			"rate limit", http.StatusTooManyRequests,
			func(r *lovable.GetProjectResponse) *lovable.ErrorResponse {
				return r.JSON429
			},
		},
		{
			"timeout", http.StatusServiceUnavailable,
			func(r *lovable.GetProjectResponse) *lovable.ErrorResponse {
				return r.JSON503
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := newTestClient(t,
				func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, http.MethodGet, r.Method)
					assert.Equal(t, "/v1/projects/"+testProjectID,
						r.URL.Path)
					w.Header().Set("Content-Type", "application/json")
					w.Header().Set("Retry-After", "3")
					w.WriteHeader(tc.status)

					err := json.NewEncoder(w).Encode(map[string]any{
						"type":       "example_error",
						"title":      tc.name,
						"status":     tc.status,
						"request_id": "request-example",
						"detail":     nil,
						"errors":     nil,
						"props":      nil,
					})
					assert.NoError(t, err)
				})

			res, err := client.GetProjectWithResponse(
				t.Context(), testProjectID, nil,
			)
			require.NoError(t, err)
			assert.Equal(t, tc.status, res.StatusCode())
			assert.Nil(t, res.JSON200)
			body := tc.pick(res)
			require.NotNil(t, body)
			assert.Equal(t, tc.status, body.Status)
			assert.Equal(t, "example_error", body.Type)
			assert.Equal(t, "request-example", body.RequestId)
			assert.True(t, body.Detail.IsNull())
			assert.Equal(t, "3",
				res.HTTPResponse.Header.Get("Retry-After"))
		})
	}
}
