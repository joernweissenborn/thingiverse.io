package peer

import (
	"fmt"
	"log"
	"time"

	"github.com/joernweissenborn/eventual2go"
	"github.com/joernweissenborn/thingiverse.io/config"
	"github.com/joernweissenborn/thingiverse.io/service/connection"
	"github.com/joernweissenborn/thingiverse.io/service/messages"
)

//go:generate event_generator -t *Peer -n Peer

// Peer is node with an rpc connection.
type Peer struct {
	uuid string
	cfg  *config.Config

	incoming *connection.Incoming

	initialized *eventual2go.Completer
	connected   *PeerCompleter

	msgIn  *messages.MessageStream
	msgOut *connection.Outgoing

	removed *eventual2go.Completer

	logger *log.Logger
}

// New creates a new Peer.
func New(uuid, address string, port int, incoming *connection.Incoming, cfg *config.Config) (p *Peer, err error) {
	p = &Peer{
		uuid:      uuid,
		cfg:       cfg,
		incoming:  incoming,
		msgIn:     incoming.MessagesFromSender(uuid),
		connected: NewPeerCompleter(),
		removed:   eventual2go.NewCompleter(),
		logger:    log.New(cfg.Logger(), fmt.Sprintf("PEER %s@%s ", uuid[:5], cfg.UUID()[:5]), 0),
	}

	p.msgOut, err = connection.NewOutgoing(cfg.UUID(), address, port)
	if err != nil {
		return
	}

	p.msgIn.Where(messages.Is(messages.HELLOOK)).Listen(p.onHelloOk)

	if cfg.Exporting() {
		p.msgIn.Where(messages.Is(messages.DOHAVE)).Listen(p.onDoHave)
		p.msgIn.Where(messages.Is(messages.CONNECT)).Listen(p.onConnected)
	}

	p.msgIn.CloseOnFuture(p.removed.Future())
	p.removed.Future().Then(p.closeOutgoing)
	return
}

// NewFromHello creates a new Peer from a HELLO message.
func NewFromHello(uuid string, m *messages.Hello, incoming *connection.Incoming, cfg *config.Config) (p *Peer, err error) {
	p, err = New(uuid, m.Address, m.Port, incoming, cfg)

	p.logger.Println("Received HELLO")
	if err != nil {
		return
	}

	p.Send(&messages.HelloOk{})
	p.setupInitCompleter()

	return
}

func (p *Peer) setupInitCompleter() {
	p.initialized = eventual2go.NewTimeoutCompleter(5 * time.Second)
	p.initialized.Future().Err(p.onTimeout)
}

// InitConnection initializes the connection.
func (p *Peer) InitConnection() {

	if p.initialized != nil {
		return
	}

	p.setupInitCompleter()

	hello := &messages.Hello{p.incoming.Addr(), p.incoming.Port()}

	p.Send(hello)

}

// Messages returns the peers message stream.
func (p *Peer) Messages() *messages.MessageStream {
	return p.msgIn
}

// Connected returns a future which gets completed if the peer succesfully connects.
func (p *Peer) Connected() *PeerFuture {
	return p.connected.Future()
}

// Remove closes the peer.
func (p *Peer) Remove() {
	p.removed.Complete(nil)
}

func (p *Peer) onTimeout(error) (eventual2go.Data, error) {
	p.removed.Complete(nil)
	return nil, nil
}

func (p *Peer) onHelloOk(messages.Message) {
	p.logger.Println("Received HELLO_OK")
	if !p.initialized.Completed() {
		p.initialized.Complete(nil)
		p.Send(&messages.HelloOk{})
	}
}

func (p *Peer) onDoHave(m messages.Message) {

	p.logger.Println("Received DOHAVE")
	if p.cfg.Exporting() {
		dohave := m.(*messages.DoHave)
		v, have := p.cfg.Tags()[dohave.TagKey]
		if have {
			have = v == dohave.TagValue
		}
		p.logger.Println("Checking tag", dohave, have)
		p.Send(&messages.Have{have, dohave.TagKey, dohave.TagValue})
		if !have {
			p.Remove()
		}
	}
}

func (p *Peer) onConnected(m messages.Message) {
	p.logger.Println("Received CONNECTED")
	p.connected.Complete(p)
}

// Removed returns a future which is completed when the connection gets closed.
func (p *Peer) Removed() *eventual2go.Future {
	return p.removed.Future()
}

// Send send a message to the peer.
func (p *Peer) Send(m messages.Message) {
	p.msgOut.Send(messages.Flatten(m))
}

func (p *Peer) closeOutgoing(eventual2go.Data) eventual2go.Data {
	p.msgOut.Close()
	return nil
}

// Check sends all tags to the peer and gets an answer if the peer supports it.
func (p *Peer) Check() {
	p.initialized.Future().Then(p.check)
}

func (p *Peer) check(eventual2go.Data) eventual2go.Data {
	msgs := p.Messages().Where(messages.Is(messages.HAVE))
	c := msgs.AsChan()
	p.logger.Println("checking tags")
	for k, v := range p.cfg.Tags() {

		p.logger.Println("checking", k, v)
		p.Send(&messages.DoHave{k, v})

		select {
		case <-time.After(5 * time.Second):
			p.logger.Println("timeout")
			p.Remove()
			return nil
		case m, ok := <-c:
			if !ok {
				p.Remove()
				p.logger.Println("incoming closed")
				return nil
			}
			have := m.(*messages.Have)

			if !have.Have || have.TagKey != k || have.TagValue != v {
				p.Remove()
				p.logger.Printf("Tag %s:%s not supported", k, v)
				return nil
			}
		}
	}
	p.Send(&messages.Connect{})
	p.logger.Println("connected successfully")
	p.connected.Complete(p)
	return nil
}
