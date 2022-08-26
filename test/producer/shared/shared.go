package shared

const Service = "producer"

const (
	EventReceivePokémon = "receive_pokémon"
)

type Pokémon struct {
	Name string `json:"name"`
}
