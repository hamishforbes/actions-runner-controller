package fake

import (
	"context"

	"github.com/actions/actions-runner-controller/github/actions"
)

type MultiClientOption func(*fakeMultiClient)

func WithDefaultClient(client actions.ActionsService, err error) MultiClientOption {
	return func(f *fakeMultiClient) {
		f.defaultClient = client
		f.defaultErr = err
	}
}

type fakeMultiClient struct {
	defaultClient actions.ActionsService
	defaultErr    error
}

func NewMultiClient(opts ...MultiClientOption) actions.MultiClient {
	f := &fakeMultiClient{}

	for _, opt := range opts {
		opt(f)
	}

	if f.defaultClient == nil {
		f.defaultClient = NewFakeClient()
	}

	return f
}

func (f *fakeMultiClient) GetClientFor(ctx context.Context, githubConfigURL string, creds actions.ActionsAuth, namespace string) (actions.ActionsService, error) {
	return f.defaultClient, f.defaultErr
}

func (f *fakeMultiClient) GetClientFromSecret(ctx context.Context, githubConfigURL, namespace string, secretData actions.KubernetesSecretData) (actions.ActionsService, error) {
	return f.defaultClient, f.defaultErr
}
