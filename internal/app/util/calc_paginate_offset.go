package util

func CalcPaginateOffset(pageNum, perPage int) int {
	if pageNum <= 1 {
		return 0
	}

	return (pageNum - 1) * perPage
}
