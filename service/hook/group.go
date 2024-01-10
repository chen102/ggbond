package hook

type Room struct {
	RommID   int32
	RoomName string
}

func (g *Room) ID() int32 {
	return g.RommID
}

func (g *Room) Name() string {
	return g.RoomName
}
