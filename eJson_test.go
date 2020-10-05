package eJson

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/mailru/easyjson/jlexer"
	fflib "github.com/pquerna/ffjson/fflib/v1"
)

func TestRandomData(t *testing.T) {
	var lstr string
	defer func() {
		if v := recover(); v != nil {
			println("'" + hex.EncodeToString([]byte(lstr)) + "'")
			println("'" + lstr + "'")
			panic(v)
		}
	}()
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 200)
	for i := 0; i < 2000000; i++ {
		n, err := rand.Read(b[:rand.Int()%len(b)])
		if err != nil {
			t.Fatal(err)
		}
		lstr = string(b[:n])
		Get(lstr, "zzzz")
	}
}

func TestRandomValidStrings(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 200)
	for i := 0; i < 100000; i++ {
		n, err := rand.Read(b[:rand.Int()%len(b)])
		if err != nil {
			t.Fatal(err)
		}
		sm, err := json.Marshal(string(b[:n]))
		if err != nil {
			t.Fatal(err)
		}
		var su string
		if err := json.Unmarshal([]byte(sm), &su); err != nil {
			t.Fatal(err)
		}
		token := Get(`{"str":`+string(sm)+`}`, "str")
		if token.Type != String || token.Str != su {
			println("["+token.Raw+"]", "["+token.Str+"]", "["+su+"]", "["+string(sm)+"]")
			t.Fatal("string mismatch")
		}
	}
}

var basicJSON = `{"age":100, "name":{"here":"B\\\"R"},
	"noop":{"what is a wren?":"a bird"},
	"happy":true,"immortal":false,
	"escaped\\\"":true,
	"arr":["1",2,"3",{"hello":"world"},"4",5],
	"vals":[1,2,3,{"sadf":sdf"asdf"}],"name":{"first":"tom","last":null}}`

func TestBasic(t *testing.T) {
	var token Result

	token = Get(basicJSON, "name.here")
	if token.String() != "B\\\"R" {
		t.Fatal("expecting 'B\\\"R'", "got", token.String())
	}
	token = Get(basicJSON, "arr.#")
	if token.String() != "6" {
		t.Fatal("expecting '6'", "got", token.String())
	}
	token = Get(basicJSON, "arr.3.hello")
	if token.String() != "world" {
		t.Fatal("expecting 'world'", "got", token.String())
	}
	_ = token.Value().(string)
	token = Get(basicJSON, "name.first")
	if token.String() != "tom" {
		t.Fatal("expecting 'tom'", "got", token.String())
	}
	_ = token.Value().(string)
	token = Get(basicJSON, "name.last")
	if token.String() != "null" {
		t.Fatal("expecting 'null'", "got", token.String())
	}
	if token.Value() != nil {
		t.Fatal("should be nil")
	}
	token = Get(basicJSON, "age")
	if token.String() != "100" {
		t.Fatal("expecting '100'", "got", token.String())
	}
	_ = token.Value().(float64)
	token = Get(basicJSON, "happy")
	if token.String() != "true" {
		t.Fatal("expecting 'true'", "got", token.String())
	}
	_ = token.Value().(bool)
	token = Get(basicJSON, "immortal")
	if token.String() != "false" {
		t.Fatal("expecting 'false'", "got", token.String())
	}
	_ = token.Value().(bool)
	token = Get(basicJSON, "noop")
	if token.String() != `{"what is a wren?":"a bird"}` {
		t.Fatal("expecting '"+`{"what is a wren?":"a bird"}`+"'", "got", token.String())
	}
	_ = token.Value().(string)

	if Get(basicJSON, "").Value() != nil {
		t.Fatal("should be nil")
	}

	if !Get(basicJSON, "escaped\\\"").Value().(bool) {
		t.Fatal("could not escape")
	}

	Get(basicJSON, "vals.hello")
}

func TestUnescape(t *testing.T) {
	unescape(string([]byte{'\\', '\\', 0}))
	unescape(string([]byte{'\\', '/', '\\', 'b', '\\', 'f'}))
}
func assert(t testing.TB, cond bool) {
	if !cond {
		t.Fatal("assert failed")
	}
}
func TestLess(t *testing.T) {
	assert(t, !Result{Type: Null}.Less(Result{Type: Null}, true))
	assert(t, Result{Type: Null}.Less(Result{Type: False}, true))
	assert(t, Result{Type: Null}.Less(Result{Type: True}, true))
	assert(t, Result{Type: Null}.Less(Result{Type: JSON}, true))
	assert(t, Result{Type: Null}.Less(Result{Type: Number}, true))
	assert(t, Result{Type: Null}.Less(Result{Type: String}, true))
	assert(t, !Result{Type: False}.Less(Result{Type: Null}, true))
	assert(t, Result{Type: False}.Less(Result{Type: True}, true))
	assert(t, Result{Type: String, Str: "abc"}.Less(Result{Type: String, Str: "bcd"}, true))
	assert(t, Result{Type: String, Str: "ABC"}.Less(Result{Type: String, Str: "abc"}, true))
	assert(t, !Result{Type: String, Str: "ABC"}.Less(Result{Type: String, Str: "abc"}, false))
	assert(t, Result{Type: Number, Num: 123}.Less(Result{Type: Number, Num: 456}, true))
	assert(t, !Result{Type: Number, Num: 456}.Less(Result{Type: Number, Num: 123}, true))
	assert(t, !Result{Type: Number, Num: 456}.Less(Result{Type: Number, Num: 456}, true))
	assert(t, stringLessInsensitive("abcde", "BBCDE"))
	assert(t, stringLessInsensitive("abcde", "bBCDE"))
	assert(t, stringLessInsensitive("Abcde", "BBCDE"))
	assert(t, stringLessInsensitive("Abcde", "bBCDE"))
	assert(t, !stringLessInsensitive("bbcde", "aBCDE"))
	assert(t, !stringLessInsensitive("bbcde", "ABCDE"))
	assert(t, !stringLessInsensitive("Bbcde", "aBCDE"))
	assert(t, !stringLessInsensitive("Bbcde", "ABCDE"))
	assert(t, !stringLessInsensitive("abcde", "ABCDE"))
	assert(t, !stringLessInsensitive("Abcde", "ABCDE"))
	assert(t, !stringLessInsensitive("abcde", "ABCDE"))
	assert(t, !stringLessInsensitive("ABCDE", "ABCDE"))
	assert(t, !stringLessInsensitive("abcde", "abcde"))
	assert(t, !stringLessInsensitive("123abcde", "123Abcde"))
	assert(t, !stringLessInsensitive("123Abcde", "123Abcde"))
	assert(t, !stringLessInsensitive("123Abcde", "123abcde"))
	assert(t, !stringLessInsensitive("123abcde", "123abcde"))
	assert(t, !stringLessInsensitive("124abcde", "123abcde"))
	assert(t, !stringLessInsensitive("124Abcde", "123Abcde"))
	assert(t, !stringLessInsensitive("124Abcde", "123abcde"))
	assert(t, !stringLessInsensitive("124abcde", "123abcde"))
	assert(t, stringLessInsensitive("124abcde", "125abcde"))
	assert(t, stringLessInsensitive("124Abcde", "125Abcde"))
	assert(t, stringLessInsensitive("124Abcde", "125abcde"))
	assert(t, stringLessInsensitive("124abcde", "125abcde"))
}

var exampleJSON = ` 
{"widget": {
    "debug": "on",
    "window": {
        "title": "Sample Konfabulator Widget",
        "name": "main_window",
        "width": 500,
        "height": 500
    },
    "image": { 
        "src": "Images/Sun.png",
        "hOffset": 250,
        "vOffset": 250,
        "alignment": "center"
    },
    "text": {
        "data": "Click Here",
        "size": 36,
        "style": "bold",
        "vOffset": 100,
        "alignment": "center",
        "onMouseUp": "sun1.opacity = (sun1.opacity / 100) * 90;"
    }
}}    
`

type BenchStruct struct {
	Widget struct {
		Window struct {
			Name string `json:"name"`
		} `json:"window"`
		Image struct {
			HOffset int `json:"hOffset"`
		} `json:"image"`
		Text struct {
			OnMouseUp string `json:"onMouseUp"`
		} `json:"text"`
	} `json:"widget"`
}

var benchPaths = []string{
	"widget.window.name",
	"widget.image.hOffset",
	"widget.text.onMouseUp",
}

func BenchmarkGJSONGet(t *testing.B) {
	t.ReportAllocs()
	for i := 0; i < t.N; i++ {
		for j := 0; j < len(benchPaths); j++ {
			if Get(exampleJSON, benchPaths[j]).Type == Null {
				t.Fatal("did not find the value")
			}
		}
	}
	t.N *= len(benchPaths)
}

func BenchmarkJSONUnmarshalMap(t *testing.B) {
	t.ReportAllocs()
	for i := 0; i < t.N; i++ {
		for j := 0; j < len(benchPaths); j++ {
			parts := strings.Split(benchPaths[j], ".")
			var m map[string]interface{}
			if err := json.Unmarshal([]byte(exampleJSON), &m); err != nil {
				t.Fatal(err)
			}
			var v interface{}
			for len(parts) > 0 {
				part := parts[0]
				if len(parts) > 1 {
					m = m[part].(map[string]interface{})
					if m == nil {
						t.Fatal("did not find the value")
					}
				} else {
					v = m[part]
					if v == nil {
						t.Fatal("did not find the value")
					}
				}
				parts = parts[1:]
			}
		}
	}
	t.N *= len(benchPaths)
}

func BenchmarkJSONUnmarshalStruct(t *testing.B) {
	t.ReportAllocs()
	for i := 0; i < t.N; i++ {
		for j := 0; j < len(benchPaths); j++ {
			var s BenchStruct
			if err := json.Unmarshal([]byte(exampleJSON), &s); err != nil {
				t.Fatal(err)
			}
			switch benchPaths[j] {
			case "widget.window.name":
				if s.Widget.Window.Name == "" {
					t.Fatal("did not find the value")
				}
			case "widget.image.hOffset":
				if s.Widget.Image.HOffset == 0 {
					t.Fatal("did not find the value")
				}
			case "widget.text.onMouseUp":
				if s.Widget.Text.OnMouseUp == "" {
					t.Fatal("did not find the value")
				}
			}
		}
	}
	t.N *= len(benchPaths) // because we are running against 3 paths
}

func BenchmarkJSONDecoder(t *testing.B) {
	t.ReportAllocs()
	for i := 0; i < t.N; i++ {
		for j := 0; j < len(benchPaths); j++ {
			dec := json.NewDecoder(bytes.NewBuffer([]byte(exampleJSON)))
			var found bool
		outer:
			for {
				tok, err := dec.Token()
				if err != nil {
					if err == io.EOF {
						break
					}
					t.Fatal(err)
				}
				switch v := tok.(type) {
				case string:
					if found {
						break outer
					}
					switch benchPaths[j] {
					case "widget.window.name":
						if v == "name" {
							found = true
						}
					case "widget.image.hOffset":
						if v == "hOffset" {
							found = true
						}
					case "widget.text.onMouseUp":
						if v == "onMouseUp" {
							found = true
						}
					}
				}
			}
			if !found {
				t.Fatal("field not found")
			}
		}
	}
	t.N *= len(benchPaths)
}

func BenchmarkFFJSONLexer(t *testing.B) {
	t.ReportAllocs()
	for i := 0; i < t.N; i++ {
		for j := 0; j < len(benchPaths); j++ {
			l := fflib.NewFFLexer([]byte(exampleJSON))
			var found bool
		outer:
			for {
				t := l.Scan()
				if t == fflib.FFTok_eof {
					break
				}
				if t == fflib.FFTok_string {
					b, _ := l.CaptureField(t)
					v := string(b)
					if found {
						// break out once we find the value.
						break outer
					}
					switch benchPaths[j] {
					case "widget.window.name":
						if v == "\"name\"" {
							found = true
						}
					case "widget.image.hOffset":
						if v == "\"hOffset\"" {
							found = true
						}
					case "widget.text.onMouseUp":
						if v == "\"onMouseUp\"" {
							found = true
						}
					}
				}
			}
			if !found {
				t.Fatal("field not found")
			}
		}
	}
	t.N *= len(benchPaths)
}

func BenchmarkEasyJSONLexer(t *testing.B) {
	t.ReportAllocs()
	skipCC := func(l *jlexer.Lexer, n int) {
		for i := 0; i < n; i++ {
			l.Skip()
			l.WantColon()
			l.Skip()
			l.WantComma()
		}
	}
	skipGroup := func(l *jlexer.Lexer, n int) {
		l.WantColon()
		l.Delim('{')
		skipCC(l, n)
		l.Delim('}')
		l.WantComma()
	}
	for i := 0; i < t.N; i++ {
		for j := 0; j < len(benchPaths); j++ {
			l := &jlexer.Lexer{Data: []byte(exampleJSON)}
			l.Delim('{')
			if l.String() == "widget" {
				l.WantColon()
				l.Delim('{')
				switch benchPaths[j] {
				case "widget.window.name":
					skipCC(l, 1)
					if l.String() == "window" {
						l.WantColon()
						l.Delim('{')
						skipCC(l, 1)
						if l.String() == "name" {
							l.WantColon()
							if l.String() == "" {
								t.Fatal("did not find the value")
							}
						}
					}
				case "widget.image.hOffset":
					skipCC(l, 1)
					if l.String() == "window" {
						skipGroup(l, 4)
					}
					if l.String() == "image" {
						l.WantColon()
						l.Delim('{')
						skipCC(l, 1)
						if l.String() == "hOffset" {
							l.WantColon()
							if l.Int() == 0 {
								t.Fatal("did not find the value")
							}
						}
					}
				case "widget.text.onMouseUp":
					skipCC(l, 1)
					if l.String() == "window" {
						skipGroup(l, 4)
					}
					if l.String() == "image" {
						skipGroup(l, 4)
					}
					if l.String() == "text" {
						l.WantColon()
						l.Delim('{')
						skipCC(l, 5)
						if l.String() == "onMouseUp" {
							l.WantColon()
							if l.String() == "" {
								t.Fatal("did not find the value")
							}
						}
					}
				}
			}
		}
	}
	t.N *= len(benchPaths)
}
