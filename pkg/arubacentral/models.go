package arubacentral

type User struct {
	Username string `json:"username"`
	Name     struct {
		First string `json:"firstname"`
		Last  string `json:"lastname"`
	} `json:"name"`
	Applications []struct {
		Name string `json:"name"`
		Info []struct {
			Role  string `json:"role"`
			Scope struct {
				Groups []string `json:"groups"`
			} `json:"scope"`
		} `json:"info"`
	} `json:"applications"`
}
