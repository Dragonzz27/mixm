package auth

import "testing"

func TestPasswordHash(t *testing.T) {
	hash, err := Hash("课程-password")
	if err != nil {
		t.Fatal(err)
	}
	if !Verify(hash, "课程-password") || Verify(hash, "wrong") {
		t.Fatal("password verification failed")
	}
}
