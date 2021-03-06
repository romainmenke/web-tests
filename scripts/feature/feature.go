package feature

type Feature struct {
	Spec struct {
		Org     string `json:"org"`
		ID      string `json:"id"`
		Section string `json:"section"`
		Name    string `json:"name"`
		URL     string `json:"url"`
	} `json:"spec"`
	Notes []struct {
		Message string `json:"message"`
	} `json:"notes"`
	SearchTerms []string          `json:"search_terms"`
	Tests       map[string]string `json:"tests"`
	PolyfillIO  []string          `json:"polyfill.io"`
}

type FeatureWithDir struct {
	Feature
	Dir string `json:"dir"`
}
