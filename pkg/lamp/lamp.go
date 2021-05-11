package lamp

type Lamp struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Group      string `json:"group"`
	Power      bool   `json:"power"`
	Brightness int    `json:"brightness"`
}
