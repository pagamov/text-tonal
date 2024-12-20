package data

type Data struct {
	Label string    `json:"label"`
	Words []string  `json:"text"`
	Vec   []float32 `json:"vec"`
}
