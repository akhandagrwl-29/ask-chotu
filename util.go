package main

func getTrimmedText(text string) string {
	maxLen := 3
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen]
}

var (
	l1 = "Hi Friend, I am tired. I am going to do some YOGA to get some energy back! Can you please ping me later?"
	l2 = "Chou is currently unavailable. Please try again later."
)
