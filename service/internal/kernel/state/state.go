package state

import (
	"errors"
	"sort"

	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

var ErrIllegalTransition = errors.New("illegal transition")

type Transition struct {
	From model.Status `json:"from"`
	To   model.Status `json:"to"`
}

type Machine struct {
	allowed map[model.Status]map[model.Status]bool
}

func NewMachine(transitions ...Transition) Machine {
	allowed := map[model.Status]map[model.Status]bool{}
	for _, t := range transitions {
		if allowed[t.From] == nil {
			allowed[t.From] = map[model.Status]bool{}
		}
		allowed[t.From][t.To] = true
	}
	return Machine{allowed: allowed}
}

func (m Machine) Can(from, to model.Status) bool {
	return m.allowed[from][to]
}

func (m Machine) TryTransition(from, to model.Status) error {
	if m.Can(from, to) {
		return nil
	}
	return ErrIllegalTransition
}

func (m Machine) AllowedTransitions(from model.Status) []model.Status {
	targets := m.allowed[from]
	next := make([]model.Status, 0, len(targets))
	for to := range targets {
		next = append(next, to)
	}
	sort.Slice(next, func(i, j int) bool { return next[i] < next[j] })
	return next
}

func (m Machine) IsTerminal(s model.Status) bool {
	return len(m.allowed[s]) == 0
}
