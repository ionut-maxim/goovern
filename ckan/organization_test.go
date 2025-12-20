package ckan_test

import (
	"context"
	"testing"

	"github.com/ionut-maxim/goovern/ckan"
)

func Test_ONRC(t *testing.T) {
	ctx := context.Background()
	client, err := ckan.New()
	if err != nil {
		t.Fatal(err)
	}

	// Get ONRC details
	org, err := client.Organization(ctx, "16c83dbe-5a2b-466b-abda-7722354b665c")
	if err != nil {
		t.Fatalf("request failed: %s", err)
	}

	if org.Name != "onrc" {
		t.Errorf("want: %s got: %s", "onrc", org.Name)
	}

	org, err = client.Organization(ctx, "something")
	if err == nil {
		t.Errorf("should fail")
	}
}
