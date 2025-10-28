package rabbit

import (
	"errors"

	"github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/zlog"
)

type ChannelPool struct {
	conn     *rabbitmq.Connection
	channels chan *rabbitmq.Channel
}

func NewChannelPool(conn *rabbitmq.Connection, poolSize int) (*ChannelPool, error) {
	pool := &ChannelPool{
		conn:     conn,
		channels: make(chan *rabbitmq.Channel, poolSize),
	}

	for i := 0; i < poolSize; i++ {
		channel, err := conn.Channel()
		if err != nil {
			return pool, errors.New("could not create channel")
		}
		pool.channels <- channel
	}
	return pool, nil
}

func (p *ChannelPool) Get() *rabbitmq.Channel {
	return <-p.channels
}

func (p *ChannelPool) Put(ch *rabbitmq.Channel) {
	if ch.IsClosed() {
		zlog.Logger.Info().Msg("ChannelPool is closed, creating a new one")
		newCh, err := p.conn.Channel()
		if err != nil {
			zlog.Logger.Err(err).Msg("could not create a new channel")
			p.channels <- ch
		}
		ch = newCh
	}
	p.channels <- ch
}
