package entity

//BuildStep build step
type BuildStep struct {
	Target string `json:"target"`
	From   string `json:"from"`
}

//Layout layout
type Layout struct {
	BuildSteps []BuildStep `json:"build_step"`
}
