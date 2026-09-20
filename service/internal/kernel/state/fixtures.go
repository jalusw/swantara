package state

func newOrderMachine() Machine {
	return NewMachine(
		Transition{From: "draft", To: "confirmed"},
		Transition{From: "draft", To: "cancelled"},
		Transition{From: "confirmed", To: "done"},
		Transition{From: "confirmed", To: "cancelled"},
	)
}
