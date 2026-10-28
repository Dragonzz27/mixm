package service

import (
	"context"
	"strings"
	"testing"

	"mixm/backend/internal/config"
	"mixm/backend/internal/infra/cluster"
	"mixm/backend/internal/model"
	"mixm/backend/internal/repository"
	"mixm/backend/internal/testutil"
)

func TestQueuedInstanceSSHUsernameMatchesPersistedID(t *testing.T) {
	db := testutil.Database(t)
	ctx := context.Background()
	user, err := repository.Write(ctx, db, "users", "", model.Record{"username": "admin", "hashed_password": "test"})
	if err != nil {
		t.Fatal(err)
	}
	uid := model.String(user, "id")
	s := &Instances{Core: NewCore(config.Config{SSHHost: "localhost", SSHPort: 30022}, db, nil, &cluster.Client{})}
	for range 2 {
		out, err := s.enqueue(ctx, model.Actor{ID: uid}, model.Record{"name": "test", "ssh_username": "client-chosen"})
		if err != nil {
			t.Fatal(err)
		}
		id := model.String(out, "instance_id")
		instance, err := repository.Get(ctx, db, "instances", id)
		if err != nil {
			t.Fatal(err)
		}
		want := "admin" + strings.ReplaceAll(id, "-", "")[:8]
		if got := model.String(instance, "ssh_username"); got != want {
			t.Fatalf("SSH username must use persisted instance ID %s: got %q, want %q", id, got, want)
		}
		task, err := repository.Get(ctx, db, "instance_tasks", model.String(out, "task_id"))
		if err != nil {
			t.Fatal(err)
		}
		if model.String(task, "instance_id") != id || model.String(model.Object(task["payload"]), "ssh_username") != want {
			t.Fatal("queued task does not match the persisted instance identity")
		}
		if _, err = repository.Write(ctx, db, "instances", id, model.Record{"status": "running"}); err != nil {
			t.Fatal(err)
		}
		target, err := repository.RunningSSHInstance(ctx, db, want)
		if err != nil || model.String(target, "id") != id {
			t.Fatalf("SSH gateway lookup does not resolve to instance %s: %v", id, err)
		}
		response := s.instanceResponse(instance)
		if model.String(response, "ssh_command") != "ssh "+want+"@localhost -p 30022" || model.String(response, "endpoint") != "/instance/"+id {
			t.Fatal("instance response uses inconsistent SSH or HTTP identity")
		}
	}
}
