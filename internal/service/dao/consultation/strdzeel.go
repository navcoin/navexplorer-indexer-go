package consultation

type strdzeel struct {
	Question string   `json:"q"`
	Min      int      `json:"m"`
	Max      int      `json:"n"`
	Version  int      `json:"v"`
	Answers  []string `json:"a"`
}
