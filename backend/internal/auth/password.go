package auth

import "golang.org/x/crypto/bcrypt"

// Hash 使用 bcrypt 生成密码哈希。
func Hash(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), err
}

// Verify 校验明文密码与已保存的哈希是否匹配。
func Verify(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
