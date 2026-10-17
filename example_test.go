package lovable_test

import (
	"context"
	"net/http"
	"os"

	"github.com/psyb0t/go-lovable"
)

// This is the README usage example. It requires a real API key.
//
//nolint:testableexamples // live API example, compile check only
func Example() {
	auth := lovable.WithRequestEditorFn(
		func(_ context.Context, req *http.Request) error {
			req.Header.Set("Lovable-API-Key", os.Getenv("LOVABLE_API_KEY"))

			return nil
		},
	)

	client, err := lovable.NewClientWithResponses(
		"https://api.lovable.dev",
		auth,
	)
	if err != nil {
		panic(err)
	}

	res, err := client.ListWorkspacesWithResponse(
		context.Background(), &lovable.ListWorkspacesParams{},
	)
	if err != nil {
		panic(err)
	}

	if res.JSON200 == nil {
		panic(res.Status())
	}

	_ = res.JSON200.Data
}
