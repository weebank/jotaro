package shared

const Service = "producer"

const (
	EventReceivePokémon = "receive-pokémon"
)

type Pokémon struct {
	Name string `json:"name"`
}
