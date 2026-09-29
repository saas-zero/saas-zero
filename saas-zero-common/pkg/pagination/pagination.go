package pagination

const (
	DefaultPage = 1
	DefaultSize = 20
	MaxSize     = 100
)

// Normalize 规范化分页参数，返回安全的 page、size 和 offset 值。
// page < 1 默认为 1，size < 1 或 > MaxSize 默认为 DefaultSize。
func Normalize(page, size int) (normalizedPage, normalizedSize, offset int) {
	if page < 1 {
		page = DefaultPage
	}
	if size < 1 || size > MaxSize {
		size = DefaultSize
	}
	return page, size, (page - 1) * size
}
