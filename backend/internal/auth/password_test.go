package auth

import "testing"

// TestPasswordHash 验证密码哈希生成与校验行为。
func TestPasswordHash(t *testing.T) {
	hash, err := Hash("课程-password")
	if err != nil {
		t.Fatal(err)
	}

	if !Verify(hash, "课程-password") || Verify(hash, "wrong") {
		t.Fatal("password verification failed")
	}
}
