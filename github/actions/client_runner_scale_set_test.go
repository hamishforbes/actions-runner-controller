package actions_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/actions/actions-runner-controller/github/actions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRunnerScaleSet(t *testing.T) {
	ctx := context.Background()
	auth := &actions.ActionsAuth{
		Token: "token",
	}

	scaleSetName := "ScaleSet"
	runnerScaleSet := actions.RunnerScaleSet{Id: 1, Name: scaleSetName}

	t.Run("Get existing scale set", func(t *testing.T) {
		want := &runnerScaleSet
		runnerScaleSetsResp := []byte(`{"count":1,"value":[{"id":1,"name":"ScaleSet"}]}`)
		server := newActionsServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Write(runnerScaleSetsResp)
		}))

		client, err := actions.NewClient(ctx, server.configURLForOrg("my-org"), auth)
		require.NoError(t, err)

		got, err := client.GetRunnerScaleSet(ctx, scaleSetName)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("GetRunnerScaleSet calls correct url", func(t *testing.T) {
		runnerScaleSetsResp := []byte(`{"count":1,"value":[{"id":1,"name":"ScaleSet"}]}`)
		url := url.URL{}
		server := newActionsServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(runnerScaleSetsResp)
			url = *r.URL
		}))

		client, err := actions.NewClient(ctx, server.configURLForOrg("my-org"), auth)
		require.NoError(t, err)

		_, err = client.GetRunnerScaleSet(ctx, scaleSetName)
		require.NoError(t, err)

		u := url.String()
		expectedUrl := fmt.Sprintf("/_apis/runtime/runnerscalesets?name=%s&api-version=6.0-preview", scaleSetName)
		assert.Equal(t, expectedUrl, u)
	})

	t.Run("Status code not found", func(t *testing.T) {
		server := newActionsServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))

		client, err := actions.NewClient(ctx, server.configURLForOrg("my-org"), auth)
		require.NoError(t, err)

		_, err = client.GetRunnerScaleSet(ctx, scaleSetName)
		assert.NotNil(t, err)
	})

	t.Run("Error when Content-Type is text/plain", func(t *testing.T) {
		server := newActionsServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "text/plain")
		}))

		client, err := actions.NewClient(ctx, server.configURLForOrg("my-org"), auth)
		require.NoError(t, err)

		_, err = client.GetRunnerScaleSet(ctx, scaleSetName)
		assert.NotNil(t, err)
	})

	t.Run("Default retries on server error", func(t *testing.T) {
		actualRetry := 0
		server := newActionsServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
			actualRetry++
		}))

		retryMax := 1
		retryWaitMax := 1 * time.Microsecond

		client, err := actions.NewClient(
			ctx,
			server.configURLForOrg("my-org"),
			auth,
			actions.WithRetryMax(retryMax),
			actions.WithRetryWaitMax(retryWaitMax),
		)
		require.NoError(t, err)

		_, err = client.GetRunnerScaleSet(ctx, scaleSetName)
		assert.NotNil(t, err)
		expectedRetry := retryMax + 1
		assert.Equalf(t, actualRetry, expectedRetry, "A retry was expected after the first request but got: %v", actualRetry)
	})

	t.Run("RunnerScaleSet count is zero", func(t *testing.T) {
		want := (*actions.RunnerScaleSet)(nil)
		runnerScaleSetsResp := []byte(`{"count":0,"value":[{"id":1,"name":"ScaleSet"}]}`)
		server := newActionsServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Write(runnerScaleSetsResp)
		}))

		client, err := actions.NewClient(ctx, server.configURLForOrg("my-org"), auth)
		require.NoError(t, err)

		got, err := client.GetRunnerScaleSet(ctx, scaleSetName)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("Multiple runner scale sets found", func(t *testing.T) {
		wantErr := fmt.Errorf("multiple runner scale sets found with name %s", scaleSetName)
		runnerScaleSetsResp := []byte(`{"count":2,"value":[{"id":1,"name":"ScaleSet"}]}`)
		server := newActionsServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Write(runnerScaleSetsResp)
		}))

		client, err := actions.NewClient(ctx, server.configURLForOrg("my-org"), auth)
		require.NoError(t, err)

		_, err = client.GetRunnerScaleSet(ctx, scaleSetName)
		require.NotNil(t, err)
		assert.Equal(t, wantErr.Error(), err.Error())
	})
}

func TestGetRunnerScaleSetById(t *testing.T) {
	ctx := context.Background()
	auth := &actions.ActionsAuth{
		Token: "token",
	}

	scaleSetCreationDateTime := time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)
	runnerScaleSet := actions.RunnerScaleSet{Id: 1, Name: "ScaleSet", CreatedOn: scaleSetCreationDateTime, RunnerSetting: actions.RunnerSetting{}}

	t.Run("Get existing scale set by Id", func(t *testing.T) {
		want := &runnerScaleSet
		rsl, err := json.Marshal(want)
		require.NoError(t, err)
		sservere := newActionsServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Write(rsl)
		}))

		client, err := actions.NewClient(ctx, sservere.configURLForOrg("my-org"), auth)
		require.NoError(t, err)

		got, err := client.GetRunnerScaleSetById(ctx, runnerScaleSet.Id)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("GetRunnerScaleSetById calls correct url", func(t *testing.T) {
		rsl, err := json.Marshal(&runnerScaleSet)
		require.NoError(t, err)

		url := url.URL{}
		server := newActionsServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(rsl)
			url = *r.URL
		}))

		client, err := actions.NewClient(ctx, server.configURLForOrg("my-org"), auth)
		require.NoError(t, err)

		_, err = client.GetRunnerScaleSetById(ctx, runnerScaleSet.Id)
		require.NoError(t, err)

		u := url.String()
		expectedUrl := fmt.Sprintf("/_apis/runtime/runnerscalesets/%d?api-version=6.0-preview", runnerScaleSet.Id)
		assert.Equal(t, expectedUrl, u)
	})

	t.Run("Status code not found", func(t *testing.T) {
		server := newActionsServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))

		client, err := actions.NewClient(ctx, server.configURLForOrg("my-org"), auth)
		require.NoError(t, err)

		_, err = client.GetRunnerScaleSetById(ctx, runnerScaleSet.Id)
		assert.NotNil(t, err)
	})

	t.Run("Error when Content-Type is text/plain", func(t *testing.T) {
		server := newActionsServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "text/plain")
		}))

		client, err := actions.NewClient(ctx, server.configURLForOrg("my-org"), auth)
		require.NoError(t, err)

		_, err = client.GetRunnerScaleSetById(ctx, runnerScaleSet.Id)
		assert.NotNil(t, err)
	})

	t.Run("Default retries on server error", func(t *testing.T) {
		actualRetry := 0
		server := newActionsServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
			actualRetry++
		}))

		retryMax := 1
		retryWaitMax := 1 * time.Microsecond
		client, err := actions.NewClient(
			ctx,
			server.configURLForOrg("my-org"),
			auth,
			actions.WithRetryMax(retryMax),
			actions.WithRetryWaitMax(retryWaitMax),
		)
		require.NoError(t, err)

		_, err = client.GetRunnerScaleSetById(ctx, runnerScaleSet.Id)
		require.NotNil(t, err)
		expectedRetry := retryMax + 1
		assert.Equalf(t, actualRetry, expectedRetry, "A retry was expected after the first request but got: %v", actualRetry)
	})

	t.Run("No RunnerScaleSet found", func(t *testing.T) {
		want := (*actions.RunnerScaleSet)(nil)
		rsl, err := json.Marshal(want)
		require.NoError(t, err)
		server := newActionsServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Write(rsl)
		}))

		client, err := actions.NewClient(ctx, server.configURLForOrg("my-org"), auth)
		require.NoError(t, err)

		got, err := client.GetRunnerScaleSetById(ctx, runnerScaleSet.Id)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})
}

func TestCreateRunnerScaleSet(t *testing.T) {
	ctx := context.Background()
	auth := &actions.ActionsAuth{
		Token: "token",
	}

	scaleSetCreationDateTime := time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)
	runnerScaleSet := actions.RunnerScaleSet{Id: 1, Name: "ScaleSet", CreatedOn: scaleSetCreationDateTime, RunnerSetting: actions.RunnerSetting{}}

	t.Run("Create runner scale set", func(t *testing.T) {
		want := &runnerScaleSet
		rsl, err := json.Marshal(want)
		require.NoError(t, err)
		server := newActionsServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Write(rsl)
		}))

		client, err := actions.NewClient(ctx, server.configURLForOrg("my-org"), auth)
		require.NoError(t, err)

		got, err := client.CreateRunnerScaleSet(ctx, &runnerScaleSet)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("CreateRunnerScaleSet calls correct url", func(t *testing.T) {
		rsl, err := json.Marshal(&runnerScaleSet)
		require.NoError(t, err)
		url := url.URL{}
		server := newActionsServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(rsl)
			url = *r.URL
		}))

		client, err := actions.NewClient(ctx, server.configURLForOrg("my-org"), auth)
		require.NoError(t, err)

		_, err = client.CreateRunnerScaleSet(ctx, &runnerScaleSet)
		require.NoError(t, err)

		u := url.String()
		expectedUrl := "/_apis/runtime/runnerscalesets?api-version=6.0-preview"
		assert.Equal(t, expectedUrl, u)
	})

	t.Run("Error when Content-Type is text/plain", func(t *testing.T) {
		server := newActionsServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "text/plain")
		}))

		client, err := actions.NewClient(ctx, server.configURLForOrg("my-org"), auth)
		require.NoError(t, err)

		_, err = client.CreateRunnerScaleSet(ctx, &runnerScaleSet)
		require.NotNil(t, err)
		var expectedErr *actions.ActionsError
		assert.True(t, errors.As(err, &expectedErr))
	})

	t.Run("Default retries on server error", func(t *testing.T) {
		actualRetry := 0
		server := newActionsServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
			actualRetry++
		}))

		retryMax := 1
		retryWaitMax := 1 * time.Microsecond

		client, err := actions.NewClient(
			ctx,
			server.configURLForOrg("my-org"),
			auth,
			actions.WithRetryMax(retryMax),
			actions.WithRetryWaitMax(retryWaitMax),
		)
		require.NoError(t, err)

		_, err = client.CreateRunnerScaleSet(ctx, &runnerScaleSet)
		require.NotNil(t, err)
		expectedRetry := retryMax + 1
		assert.Equalf(t, actualRetry, expectedRetry, "A retry was expected after the first request but got: %v", actualRetry)
	})
}

func TestUpdateRunnerScaleSet(t *testing.T) {
	ctx := context.Background()
	auth := &actions.ActionsAuth{
		Token: "token",
	}

	scaleSetCreationDateTime := time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)
	runnerScaleSet := actions.RunnerScaleSet{Id: 1, Name: "ScaleSet", RunnerGroupId: 1, RunnerGroupName: "group", CreatedOn: scaleSetCreationDateTime, RunnerSetting: actions.RunnerSetting{}}

	t.Run("Update runner scale set", func(t *testing.T) {
		want := &runnerScaleSet
		rsl, err := json.Marshal(want)
		require.NoError(t, err)
		server := newActionsServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Write(rsl)
		}))

		client, err := actions.NewClient(ctx, server.configURLForOrg("my-org"), auth)
		require.NoError(t, err)

		got, err := client.UpdateRunnerScaleSet(ctx, 1, &actions.RunnerScaleSet{RunnerGroupId: 1})
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("UpdateRunnerScaleSet calls correct url", func(t *testing.T) {
		rsl, err := json.Marshal(&runnerScaleSet)
		require.NoError(t, err)
		url := url.URL{}
		method := ""
		server := newActionsServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(rsl)
			url = *r.URL
			method = r.Method
		}))

		client, err := actions.NewClient(ctx, server.configURLForOrg("my-org"), auth)
		require.NoError(t, err)

		_, err = client.UpdateRunnerScaleSet(ctx, 1, &runnerScaleSet)
		require.NoError(t, err)

		u := url.String()
		expectedUrl := "/_apis/runtime/runnerscalesets/1?api-version=6.0-preview"
		assert.Equal(t, expectedUrl, u)

		assert.Equal(t, "PATCH", method)
	})
}
