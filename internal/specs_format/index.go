package specs_format

// TODO handle big ints

// --- Inputs
type Balances = map[string]map[string]int64
type AccountsMeta = map[string]map[string]string
type Vars = map[string]string

// --- Outputs
type Posting struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Amount      int64  `json:"amount"`
	Asset       string `json:"asset"`
}
type TxMeta = map[string]string

// --- Specs:
type Specs struct {
	It               string       `json:"it"`
	Balances         Balances     `json:"balances,omitempty"`
	Vars             Vars         `json:"vars,omitempty"`
	Meta             AccountsMeta `json:"accountsMeta,omitempty"`
	TestCases        []Specs      `json:"testCases,omitempty"`
	ExpectedPostings []Posting    `json:"expectedPostings,omitempty"`
}
