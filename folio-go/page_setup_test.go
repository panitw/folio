package folio

import (
	"bytes"
	"os"
	"testing"
)

func TestPageSetupCommandChangesOnlyValidCanonicalPageState(t *testing.T) {
	input, err := os.ReadFile("testdata/template/golden/worked-example.json")
	if err != nil {
		t.Fatal(err)
	}
	tpl, err := ParseTemplate(input)
	if err != nil {
		t.Fatal(err)
	}
	before, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	projection, err := ApplyPageSetupCommand(tpl, []byte(`{"kind":"pageSetup","version":1,"preset":"custom","orientation":"landscape","width":300.125,"height":400.5,"margin":{"top":10,"right":11.5,"bottom":12,"left":13}}`))
	if err != nil {
		t.Fatal(err)
	}
	if projection.Width != 400500 || projection.Height != 300125 || projection.Bands[0].Name != "pageHeader" || projection.Bands[1].Name != "content" || projection.Bands[2].Name != "pageFooter" {
		t.Fatalf("unexpected projection: %#v", projection)
	}
	after, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(before, after) {
		t.Fatal("valid page setup did not change canonical bytes")
	}
	stable := append([]byte(nil), after...)
	for _, command := range [][]byte{[]byte(`{"kind":"pageSetup","version":1,"preset":"custom","orientation":"portrait","width":0,"height":200,"margin":{"top":1,"right":1,"bottom":1,"left":1}}`), []byte(`{"kind":"pageSetup","version":1,"preset":"custom","orientation":"portrait","width":1.0001,"height":2,"margin":{"top":1,"right":1,"bottom":1,"left":1}}`), []byte(`{"kind":"pageSetup","version":1,"preset":"A4","orientation":"diagonal","width":0,"height":0,"margin":{"top":1,"right":1,"bottom":1,"left":1}}`)} {
		if _, err := ApplyPageSetupCommand(tpl, command); err == nil {
			t.Fatalf("invalid command succeeded: %s", command)
		}
		got, err := SerializeTemplate(tpl)
		if err != nil || !bytes.Equal(got, stable) {
			t.Fatal("invalid command changed canonical bytes")
		}
	}
}

func TestCustomLandscapeApplyIsByteStableAndTransportSafe(t *testing.T) {
	input, err := os.ReadFile("testdata/template/golden/worked-example.json")
	if err != nil {
		t.Fatal(err)
	}
	tpl, err := ParseTemplate(input)
	if err != nil {
		t.Fatal(err)
	}
	command := []byte(`{"kind":"pageSetup","version":1,"preset":"custom","orientation":"landscape","width":300.125,"height":400.5,"margin":{"top":10,"right":11.5,"bottom":12,"left":13}}`)
	projection, err := ApplyPageSetupCommand(tpl, command)
	if err != nil || projection.CommandWidth != 300125 || projection.CommandHeight != 400500 || projection.Width != 400500 || projection.Height != 300125 {
		t.Fatalf("custom landscape projection = %#v, %v", projection, err)
	}
	before, _ := SerializeTemplate(tpl)
	if _, err := ApplyPageSetupCommand(tpl, command); err != nil {
		t.Fatal(err)
	}
	after, _ := SerializeTemplate(tpl)
	if !bytes.Equal(before, after) {
		t.Fatal("unchanged custom landscape command changed canonical bytes")
	}
	unsafe := []byte(`{"kind":"pageSetup","version":1,"preset":"custom","orientation":"portrait","width":9007199254740.992,"height":200,"margin":{"top":1,"right":1,"bottom":1,"left":1}}`)
	if _, err := ApplyPageSetupCommand(tpl, unsafe); err == nil {
		t.Fatal("first JavaScript-unsafe millipoint was accepted")
	}
}
