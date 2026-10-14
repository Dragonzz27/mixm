package model

func BankType(t string) string {
	switch t {
	case "fill_blank":
		return "blank"
	case "short_answer":
		return "essay"
	}
	return t
}
