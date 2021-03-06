package message

import (
	"bytes"

	"github.com/ThingiverseIO/thingiverseio/descriptor"
	"github.com/ugorji/go/codec"
)

type Have struct {
	Have bool
	Tag  descriptor.Tag
}

func (*Have) New() Message {
	return new(Have)
}

func (*Have) GetType() Type { return HAVE }

func (h *Have) Unflatten(d [][]byte) {
	dec := codec.NewDecoder(bytes.NewBuffer(d[0]), &mh)
	dec.Decode(&h)
}

func (h *Have) Flatten() [][]byte {
	var payload bytes.Buffer
	enc := codec.NewEncoder(&payload, &mh)
	enc.Encode(h)
	return [][]byte{payload.Bytes()}
}

func init() {
	registerMessage(new(Have))
}
