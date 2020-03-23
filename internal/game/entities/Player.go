package entities

import (
	"github.com/tpetrychyn/rsps-comm-test/internal/game"
	"github.com/tpetrychyn/rsps-comm-test/internal/game/components"
	"github.com/tpetrychyn/rsps-comm-test/pkg/models"
	"github.com/tpetrychyn/rsps-comm-test/pkg/packet"
	"github.com/tpetrychyn/rsps-comm-test/pkg/packet/outgoing"
)

type Player struct {
	*models.Actor
	World             *game.World
	MovementComponent *components.MovementComponent
	OutgoingQueue     chan packet.DownstreamMessage
}

func NewPlayer(world *game.World) *Player {
	actor := models.NewActor()
	actor.UpdateMask.Appearance = true

	player := &Player{
		Actor:             actor,
		World:             world,
	}
	player.MovementComponent = components.NewMovementComponent(actor.Movement, world, player.UpdateMask.AddMovementFlag)

	return player
}

func (p *Player) Pretick() {
	last := p.Movement.LastKnownRegionBase
	if last == nil || shouldRebuildRegion(last, p.Movement.Position) {
		regionX := ((p.Movement.Position.X >> 3) - (outgoing.MaxViewport >> 4)) << 3
		regionY := ((p.Movement.Position.Y >> 3) - (outgoing.MaxViewport >> 4)) << 3
		p.Movement.LastKnownRegionBase = &models.Tile{X: regionX, Y: regionY}

		p.OutgoingQueue <- &outgoing.RebuildNormalPacket{Position: &models.Tile{X: p.Movement.Position.X >> 3, Y: p.Movement.Position.Y >> 3}}
	}
}

func (p *Player) Tick() {
	p.MovementComponent.Tick()

	p.OutgoingQueue <- outgoing.NewPlayerUpdatePacket(p.Actor)
}

func (p *Player) PostTick() {
	p.Actor.UpdateMask.Clear()
	p.Movement.WalkDirection = models.Direction.None
	p.Movement.RunDirection = models.Direction.None
	p.Movement.Teleported = false
}

const NormalViewDistance = 15

func shouldRebuildRegion(old *models.Tile, new *models.Tile) bool {
	dx := new.X - old.X
	dy := new.Y - old.Y
	return dx <= NormalViewDistance || dx >= outgoing.MaxViewport-NormalViewDistance-1 ||
		dy <= NormalViewDistance || dy >= outgoing.MaxViewport-NormalViewDistance-1
}
