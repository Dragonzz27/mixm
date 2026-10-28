package app

import (
	"context"
	"net/url"
	"os"
	"testing"

	"mixm/backend/internal/auth"
	"mixm/backend/internal/model"
	"mixm/backend/internal/repository"
	"mixm/backend/internal/testutil"
)

func TestSeedAndAdminUseExistingSchema(t *testing.T) {
	db := testutil.Database(t)
	ctx := context.Background()
	schema, err := repository.One(ctx, db, "SELECT current_schema() AS name")
	if err != nil {
		t.Fatal(err)
	}
	database, err := url.Parse(os.Getenv("TEST_DATABASE_URL"))
	if err != nil {
		t.Fatal(err)
	}
	params := database.Query()
	params.Set("search_path", model.String(schema, "name"))
	database.RawQuery = params.Encode()
	t.Setenv("DATABASE_URL", database.String())
	t.Setenv("SECRET_KEY", "test-secret")
	t.Setenv("ADMIN_EMAIL", "bootstrap@example.com")
	t.Setenv("ADMIN_USERNAME", "bootstrap")
	t.Setenv("ADMIN_PASSWORD", "test-password")
	if err := Seed(); err != nil {
		t.Fatal(err)
	}
	if err := Seed(); err != nil {
		t.Fatal("seed not idempotent", err)
	}
	roles, err := repository.AllRoles(ctx, db)
	if err != nil || len(roles) != 2 {
		t.Fatal("roles not seeded", err)
	}
	if err := CreateAdmin(); err != nil {
		t.Fatal(err)
	}
	user, err := repository.LoginUser(ctx, db, "bootstrap@example.com")
	if err != nil || !model.Bool(user, "is_superuser") || !auth.Verify(model.String(user, "hashed_password"), "test-password") {
		t.Fatal("admin was not created", err)
	}
}
