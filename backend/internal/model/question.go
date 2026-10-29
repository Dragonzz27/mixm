package model

// BankType 将题目类型转换为统一的题库类型。
func BankType(t string) string {
	switch t {
	case "fill_blank":
		return "blank"
	case "short_answer":
		return "essay"
	}

	return t
}
