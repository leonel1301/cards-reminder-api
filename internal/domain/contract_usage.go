package domain

// ContractAnalyzeLimit is the beta lifetime cap per user for AI contract analysis.
const ContractAnalyzeLimit = 5

// ContractUsage reports how many AI contract analyses a user has consumed.
type ContractUsage struct {
	Used      int `json:"used"`
	Limit     int `json:"limit"`
	Remaining int `json:"remaining"`
}

func NewContractUsage(used int) ContractUsage {
	if used < 0 {
		used = 0
	}
	remaining := ContractAnalyzeLimit - used
	if remaining < 0 {
		remaining = 0
	}
	return ContractUsage{
		Used:      used,
		Limit:     ContractAnalyzeLimit,
		Remaining: remaining,
	}
}
