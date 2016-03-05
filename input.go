package thingiverseio

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/joernweissenborn/eventual2go"
	"github.com/joernweissenborn/eventual2go/typed_events"
	"github.com/joernweissenborn/thingiverseio/config"
	"github.com/joernweissenborn/thingiverseio/service/connection"
	"github.com/joernweissenborn/thingiverseio/service/manager"
	"github.com/joernweissenborn/thingiverseio/service/messages"
)

type Input struct {
	cfg     *config.Config
	m       *manager.Manager
	r       *eventual2go.Reactor
	results *messages.ResultStream
	listen  map[string]interface{}
	logger  *log.Logger
}

func NewInput(desc string) (i *Input, err error) {
	i, err = NewInputFromConfig(config.Configure(ioutil.Discard, false, descFromYaml(desc).AsTagSet()))
	return
}

func NewInputFromConfig(cfg *config.Config) (i *Input, err error) {
	m, err := manager.New(cfg)
	i = &Input{
		m:       m,
		cfg:     cfg,
		r:       eventual2go.NewReactor(),
		listen:  map[string]interface{}{},
		logger:  log.New(cfg.Logger(), fmt.Sprintf("%s input ", cfg.UUID()), log.Lshortfile),
		results: &messages.ResultStream{m.MessagesOfType(messages.RESULT).Transform(connection.ToMessage)},
	}

	i.r.React(arriveEvent{}, i.sendListenFunctions)
	i.r.AddStream(arriveEvent{}, m.PeerArrive().Stream)

	i.r.React(startListenEvent{}, i.startListen)
	i.r.React(stopListenEvent{}, i.stopListen)

	return
}

func (i *Input) UUID() config.UUID {
	return i.cfg.UUID()
}

func (i *Input) Remove() (errs []error) {
	errs = i.m.Shutdown()
	i.r.Shutdown(nil)
	return
}

func (i *Input) Run() {
	i.m.Run()
}

func (i *Input) Connected() *typed_events.BoolFuture {
	return i.m.Connected().FirstWhere(func(b bool) bool { return b })
}

func (i *Input) Disconnected() *typed_events.BoolFuture {
	return i.m.Connected().FirstWhereNot(func(b bool) bool { return b })
}

func (i *Input) sendListenFunctions(d eventual2go.Data) {
	uuid := d.(config.UUID)
	i.logger.Println("found output", uuid)
	for f := range i.listen {
		i.m.SendTo(uuid, &messages.Listen{f})
	}
	return
}

func (i *Input) Call(function string, parameter interface{}) (f *messages.ResultFuture) {
	i.logger.Println("Call", function)
	req := i.newRequest(function, parameter, messages.ONE2ONE)
	f = i.results.FirstWhere(isRes(req.UUID))
	akn := i.m.SendGuaranteed(req)
	i.logger.Println("Call", akn)
	f.Future.Then(acknowledgeResult(akn))
	return
}

func acknowledgeResult(akn *eventual2go.Completer) eventual2go.CompletionHandler {
	return func(eventual2go.Data) eventual2go.Data {
		fmt.Println("Received Result", akn)
		akn.Complete(nil)
		return nil
	}
}

func (i *Input) CallAll(function string, parameter interface{}, s *messages.ResultStreamController) {
	i.logger.Println("CallAll", function)
	req := i.newRequest(function, parameter, messages.ONE2MANY)
	s.Join(i.results.Where(isRes(req.UUID)))
	i.m.SendToAll(req)
	return
}

func (i *Input) Trigger(function string, parameter interface{}) {
	i.m.Send(i.newRequest(function, parameter, messages.MANY2ONE))
}

func (i *Input) TriggerAll(function string, parameter interface{}) {
	i.m.SendToAll(i.newRequest(function, parameter, messages.MANY2MANY))
}

func (i *Input) Listen(function string) {
	i.r.Fire(startListenEvent{}, function)
}
func (i *Input) startListen(d eventual2go.Data) {
	function := d.(string)
	i.logger.Println("started listenting to functipn", function)
	i.listen[function] = nil
	i.m.SendToAll(&messages.Listen{function})
}

func (i *Input) StopListen(function string) {
	i.r.Fire(stopListenEvent{}, function)
}
func (i *Input) stopListen(d eventual2go.Data) {
	function := d.(string)
	if _, ok := i.listen[function]; ok {
		delete(i.listen, function)
		i.m.SendToAll(&messages.StopListen{function})
	}

}

func (i *Input) Results() *messages.ResultStream {
	return i.results
}

func (i *Input) ListenResults() *messages.ResultStream {
	return i.results.Where(func(d *messages.Result) bool {
		return d.Request.CallType == messages.MANY2MANY || d.Request.CallType == messages.MANY2ONE
	})
}

func (i *Input) newRequest(function string, parameter interface{}, ctype messages.CallType) (req *messages.Request) {

	req = messages.NewRequest(i.cfg.UUID(), function, ctype, parameter)

	return
}

func (i *Input) NewRequestBin(function string, parameter []byte, ctype messages.CallType) (req *messages.Request) {
	req = messages.NewEncodedRequest(i.cfg.UUID(), function, ctype, parameter)
	return
}

func isRes(uuid config.UUID) messages.ResultFilter {
	return func(d *messages.Result) bool {
		return d.Request.UUID == uuid
	}
}
