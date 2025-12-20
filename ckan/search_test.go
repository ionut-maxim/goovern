package ckan_test

import (
	"context"
	"testing"

	"github.com/ionut-maxim/goovern/ckan"
)

func Test_Search(t *testing.T) {
	ctx := context.Background()
	client, err := ckan.New()
	if err != nil {
		t.Fatal(err)
	}

	search, err := client.Search(ctx, "onrc", 4)
	if err != nil {
		t.Fatal(err)
	}

	if search.Count == 0 {
		t.Error("Search returned zero results")
	}
}
