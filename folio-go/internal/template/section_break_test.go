package template

import (
	"errors"
	"strings"
	"testing"

	"github.com/panitw/folio/folio-go/internal/diag"
)

// sectionBreakDoc is a canonical document whose three bands carry the given
// extra band keys (each a `"key": value` fragment, or "").
func sectionBreakDoc(content, footer, header, version string) string {
	withKey := func(fragment string) string {
		if fragment == "" {
			return ""
		}
		return ",\n      " + fragment
	}
	return `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {
          "height": 14,
          "id": "e1",
          "type": "text",
          "value": "Legend",
          "width": 100,
          "x": 0,
          "y": 500
        }
      ]` + withKey(content) + `
    },
    "pageFooter": {
      "elements": [],
      "height": 20` + withKey(footer) + `
    },
    "pageHeader": {
      "elements": [],
      "height": 20` + withKey(header) + `
    }
  },
  "fonts": {},
  "locale": "en",
  "nextId": 2,
  "page": {
    "margin": {
      "bottom": 36,
      "left": 36,
      "right": 36,
      "top": 36
    },
    "orientation": "portrait",
    "size": "A4"
  },
  "utcOffset": "+07:00",
  "version": "` + version + `"
}
`
}

func TestSectionBreakRoundTripsByteIdenticallyAt41(t *testing.T) {
	src := sectionBreakDoc(`"sectionBreak": 480.5`, "", "", "4.1")
	d, err := ParseDocument([]byte(src))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !d.Bands.Content.SectionBreak.Set || d.Bands.Content.SectionBreak.Value != 480500 {
		t.Fatalf("SectionBreak = %+v, want 480500mp", d.Bands.Content.SectionBreak)
	}
	if len(d.Bands.Content.Extra) != 0 {
		t.Fatalf("sectionBreak leaked into Extra: %+v", d.Bands.Content.Extra)
	}
	out, err := SerializeDocument(d)
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}
	if string(out) != src {
		t.Fatalf("round trip is not byte-identical:\n got: %s\nwant: %s", out, src)
	}
}

func TestSectionBreakRaisesTheVersionTo41(t *testing.T) {
	d, err := ParseDocument([]byte(sectionBreakDoc(`"sectionBreak": 480`, "", "", "1.0")))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	out, err := SerializeDocument(d)
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}
	if !strings.Contains(string(out), `"version": "4.1"`) {
		t.Fatalf("a document carrying sectionBreak must declare 4.1:\n%s", out)
	}

	plain, err := ParseDocument([]byte(sectionBreakDoc("", "", "", "1.0")))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	out, err = SerializeDocument(plain)
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}
	if !strings.Contains(string(out), `"version": "1.0"`) {
		t.Fatalf("a document without sectionBreak must keep its version:\n%s", out)
	}
}

func TestSectionBreakLoadRefusals(t *testing.T) {
	for _, c := range []struct {
		label, content, footer, header, field string
	}{
		{"on the page header", "", "", `"sectionBreak": 10`, "bands.pageHeader.sectionBreak"},
		{"on the page footer", "", `"sectionBreak": 10`, "", "bands.pageFooter.sectionBreak"},
		{"declared twice", `"sectionBreak": 400,
      "sectionBreak": 480`, "", "", "bands.content.sectionBreak"},
		{"null", `"sectionBreak": null`, "", "", "bands.content.sectionBreak"},
	} {
		t.Run(c.label, func(t *testing.T) {
			_, err := ParseDocument([]byte(sectionBreakDoc(c.content, c.footer, c.header, "4.1")))
			var le *LoadError
			if !errors.As(err, &le) {
				t.Fatalf("want a *LoadError, got %T %v", err, err)
			}
			if le.Code != diag.CodeSectionBreakInvalid {
				t.Errorf("code = %q, want %q", le.Code, diag.CodeSectionBreakInvalid)
			}
			if le.Field != c.field {
				t.Errorf("field = %q, want the band's %q", le.Field, c.field)
			}
		})
	}
}

func TestCountTopLevelKeyIgnoresNestedKeys(t *testing.T) {
	raw := []byte(`{"elements": [{"sectionBreak": 1}], "sectionBreak": 2, "x": {"sectionBreak": 3}}`)
	if got := countTopLevelKey(raw, "sectionBreak"); got != 1 {
		t.Fatalf("countTopLevelKey = %d, want 1", got)
	}
}
