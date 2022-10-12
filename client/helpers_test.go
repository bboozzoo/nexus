package client_test

import (
	"reflect"
	"testing"

	"github.com/gammazero/nexus/v3/client"
	"github.com/gammazero/nexus/v3/wamp"
)

func TestUnpackSimple(t *testing.T) {
	l := wamp.List{"foo", 1,
		[]int{1, 2, 3},
		[]string{"foo", "bar"},
		map[string]string{"a": "b", "c": "d"},
	}
	var s string
	var i int
	var ls []string
	var li []int
	var ms map[string]string
	err := client.Unpack(l, &s, &i, &li, &ls, &ms)
	if err != nil {
		t.Fatal(err)
	}
	if s != "foo" {
		t.Fatalf("got %q", s)
	}
	if i != 1 {
		t.Fatalf("got %v", i)
	}
	if !reflect.DeepEqual(li, []int{1, 2, 3}) {
		t.Fatalf("got %v", li)
	}
	if !reflect.DeepEqual(ls, []string{"foo", "bar"}) {
		t.Fatalf("got %v", ls)
	}
	expectedMs := map[string]string{"a": "b", "c": "d"}
	if !reflect.DeepEqual(ms, expectedMs) {
		t.Fatalf("got: %v", ms)
	}
}

func TestUnpackListToStruct(t *testing.T) {
	l := wamp.List{
		map[string]interface{}{
			"string-field": "b",
			"float-field":  0.2,
			"int-field":    1,
			"ImplicitName": 2,
		},
	}

	type ts struct {
		StringField  string  `wamp:"string-field"`
		FloatField   float64 `wamp:"float-field"`
		IntField     int     `wamp:"int-field"`
		ImplicitName int
	}
	var s ts
	err := client.Unpack(l, &s)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(s, ts{
		StringField:  "b",
		FloatField:   0.2,
		IntField:     1,
		ImplicitName: 2,
	}) {
		t.Fatalf("got: %v", s)
	}
}

func TestUnpackDictToStruct(t *testing.T) {
	d := wamp.Dict{
		"string-field":      "b",
		"float-field":       0.2,
		"int-field":         1,
		"ImplicitName":      2,
		"no-matching-field": 123,
		"ignored":           111,
		"optional-field":    -1,
	}

	type ts struct {
		StringField   string  `wamp:"string-field"`
		FloatField    float64 `wamp:"float-field"`
		IntField      int     `wamp:"int-field"`
		ImplicitName  int
		UnusedField   interface{} `wamp:",optional"`
		IgnoredField  int         `wamp:",optional"`
		OptionalField int64       `wamp:"optional-field,optional"`
	}
	var s ts
	err := client.UnpackDict(d, &s)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(s, ts{
		StringField:   "b",
		FloatField:    0.2,
		IntField:      1,
		ImplicitName:  2,
		OptionalField: -1,
	}) {
		t.Fatalf("got: %v", s)
	}
}
