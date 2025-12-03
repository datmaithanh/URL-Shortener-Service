package utils

const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func EncodeBase62(num int64) string {
	if num == 0 {
		return "0"
	}
	result := ""
	for num > 0 {
		remainder := num % 62
		result = string(base62Chars[remainder]) + result
		num = num / 62
	}
	return result
}
