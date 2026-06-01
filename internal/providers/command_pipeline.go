package providers

// CommandToken is one token in a pipeline stage from the model.
type CommandToken struct {
	Value string `json:"value"`
	Expr  bool   `json:"expr"`
}

// CommandStage is one pipe segment (connector is always "|", owned by CLX).
type CommandStage struct {
	Tokens []CommandToken `json:"tokens"`
}
