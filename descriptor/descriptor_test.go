package descriptor_test

import (
	"testing"

	"github.com/ThingiverseIO/thingiverseio/descriptor"
)

var testdesc = `
function funcname(param1 string, param2 []int) (outp1 string, outp2 []bool)
function fun1(param1 string, param2 []int) ()
function fun2() (outp1 string, outp2 []bool)
function emptyfun()()
property p2: bla string
stream fluss: wasser float
tag simple_tag
tag key_tag: tag_value
#define multiple tags in one line
tags multisimple, multikey:val
`

func TestParseDescriptor(t *testing.T) {
	desc, err := descriptor.Parse(testdesc)
	if err != nil {
		t.Fatal(err)
	}
	if len(desc.Properties) != 1 {
		t.Error("Wrong number of properties, want 1, got", len(desc.Properties))
	}
	if len(desc.Properties[0].Parameter) != 1 {
		t.Error("Wrong number of properties parameter, want 1, got", len(desc.Properties[0].Parameter))
	}
	if len(desc.Streams) != 1 {
		t.Error("Wrong number of properties, want 1, got", len(desc.Streams))
	}
	if len(desc.Streams[0].Parameter) != 1 {
		t.Error("Wrong number of properties parameter, want 1, got", len(desc.Streams[0].Parameter))
	}
	if len(desc.Functions) != 4 {
		t.Error("Wrong number of functions, want 3, got", len(desc.Functions))
	}
	if len(desc.Functions[0].Input) != 2 {
		t.Error("Wrong number of input parameters, want 2, got", len(desc.Functions[0].Input))
	}
	if len(desc.Functions[0].Output) != 2 {
		t.Error("Wrong number of output parameters, want 2, got", len(desc.Functions[0].Output))
	}

	if len(desc.Tags) != 4 {
		t.Error("Wrong Number of tags, want 4, got", len(desc.Tags))
	}
}
