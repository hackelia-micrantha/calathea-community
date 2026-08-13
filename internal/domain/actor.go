package domain

// Authority identifies the domain authority asserted by an actor.
type Authority string

const (
	AuthorityMaintainer Authority = "maintainer"
)

// Actor identifies who is responsible for a canonical decision.
type Actor struct {
	id        ActorID
	authority Authority
}

// NewMaintainerActor creates the only canonical decision authority supported by v0.
func NewMaintainerActor(id ActorID) (Actor, error) {
	if err := requireIdentifier("actor id", string(id)); err != nil {
		return Actor{}, err
	}
	return Actor{id: id, authority: AuthorityMaintainer}, nil
}

func (a Actor) ID() ActorID          { return a.id }
func (a Actor) Authority() Authority { return a.authority }
func (a Actor) IsMaintainer() bool   { return a.authority == AuthorityMaintainer && a.id != "" }
