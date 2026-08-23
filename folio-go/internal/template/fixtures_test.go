package template

// This file holds the trap and passthrough fixtures the round-trip
// property tests (roundtrip_test.go) run over. Each is asserted
// canonical by TestP3FixturesAreCanonical before being trusted for
// anything else.

// unknownKeysFixture carries an unknown key at every nesting level this
// story's finisher review (Finding 2, D-1.4.9 OWNER) confirmed must
// accept opaque passthrough: top level, inside a band, inside an
// element, inside a column, inside style, inside page, inside
// page.margin, inside style.padding, inside style.border, and inside
// one assets[entry] object — ten of the document's eleven object
// levels. (bands itself is the deliberate eleventh, closed exception —
// see Bands' doc comment in model.go and TestBandsRejectsAnUnknownKey
// below.) Also carries a non-ASCII top-level key ("\u00e4NonAsciiOrderMarker")
// alongside an ASCII one ("zzAsciiOrderMarker") so AC18's byte-order key
// sort is asserted rather than holding "by coincidence over an
// all-ASCII corpus" (Finding 11): byte-order sorts "zz..." before
// "\u00e4..." (0x7A < 0xC3), which a locale-aware collation would not.
var unknownKeysFixture = []byte(`{
  "assets": {
    "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb": {
      "data": [
        "Zg=="
      ],
      "futureAssetEntryKey": true,
      "mediaType": "image/png"
    }
  },
  "bands": {
    "content": {
      "elements": [
        {
          "as": "row",
          "bind": "items[]",
          "columns": [
            {
              "bind": "{{row.n}}",
              "futureColumnKey": true,
              "id": "e2",
              "label": "N",
              "width": 40
            }
          ],
          "futureElementKey": [
            1,
            2
          ],
          "headerHeight": 14,
          "id": "e1",
          "style": {
            "border": {
              "edges": [
                "bottom"
              ],
              "futureBorderKey": true
            },
            "fontSize": 9,
            "futureStyleKey": "z",
            "padding": {
              "futurePaddingKey": true,
              "left": 3
            }
          },
          "type": "table",
          "x": 0,
          "y": 0
        }
      ],
      "futureBandKey": 1
    },
    "pageFooter": {
      "elements": [],
      "height": 20
    },
    "pageHeader": {
      "elements": [],
      "height": 20
    }
  },
  "fonts": {},
  "futureTopLevelKey": {
    "nested": [
      1,
      2,
      "th ท 中"
    ]
  },
  "locale": "en",
  "nextId": 3,
  "page": {
    "futurePageKey": true,
    "margin": {
      "bottom": 36,
      "futureMarginKey": true,
      "left": 36,
      "right": 36,
      "top": 36
    },
    "orientation": "portrait",
    "size": "A4"
  },
  "utcOffset": "+00:00",
  "version": "1.0",
  "zzAsciiOrderMarker": true,
  "äNonAsciiOrderMarker": true
}
`)

// htmlEscapeTrapFixture carries '<', '>' and '&' in a text value — AC19:
// HTML escaping must be OFF, so these appear literally, never as
// entity-escaped forms.
var htmlEscapeTrapFixture = []byte(`{
  "assets": {},
  "bands": {
    "content": {
      "elements": []
    },
    "pageFooter": {
      "elements": [],
      "height": 20
    },
    "pageHeader": {
      "elements": [
        {
          "height": 14,
          "id": "e1",
          "type": "text",
          "value": "a<b>c&d",
          "width": 100,
          "x": 0,
          "y": 0
        }
      ],
      "height": 20
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
  "utcOffset": "+00:00",
  "version": "1.0"
}
`)

// utf8TrapFixture carries Thai and CJK text — AC20: UTF-8 is emitted
// literally, no \uXXXX escape above 0x1F.
var utf8TrapFixture = []byte(`{
  "assets": {},
  "bands": {
    "content": {
      "elements": []
    },
    "pageFooter": {
      "elements": [],
      "height": 20
    },
    "pageHeader": {
      "elements": [
        {
          "height": 14,
          "id": "e1",
          "type": "text",
          "value": "ใบแจ้งหนี้ 中文 日本語",
          "width": 200,
          "x": 0,
          "y": 0
        }
      ],
      "height": 20
    }
  },
  "fonts": {},
  "locale": "th",
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
  "version": "1.0"
}
`)

// minimalEscapeTrapFixture carries '"', '\\', a control character below
// 0x20 (newline) and '/' — AC21: only '"', '\\' and sub-0x20 controls
// are escaped; '/' is never escaped.
var minimalEscapeTrapFixture = []byte("{\n  \"assets\": {},\n  \"bands\": {\n    \"content\": {\n      \"elements\": []\n    },\n    \"pageFooter\": {\n      \"elements\": [],\n      \"height\": 20\n    },\n    \"pageHeader\": {\n      \"elements\": [\n        {\n          \"height\": 14,\n          \"id\": \"e1\",\n          \"type\": \"text\",\n          \"value\": \"a \\\"quoted\\\" \\\\ line1\\nline2 a/b\",\n          \"width\": 200,\n          \"x\": 0,\n          \"y\": 0\n        }\n      ],\n      \"height\": 20\n    }\n  },\n  \"fonts\": {},\n  \"locale\": \"en\",\n  \"nextId\": 2,\n  \"page\": {\n    \"margin\": {\n      \"bottom\": 36,\n      \"left\": 36,\n      \"right\": 36,\n      \"top\": 36\n    },\n    \"orientation\": \"portrait\",\n    \"size\": \"A4\"\n  },\n  \"utcOffset\": \"+00:00\",\n  \"version\": \"1.0\"\n}\n")

// maximalFixture exercises every one of the 51 keys the serializer can
// emit (D-1.4.15's "maximal document"; this story's finisher review,
// Findings 1 and 8): every element kind, every style/border/padding
// field, the footer/footerOf/footerFormat trio, an image asset, and
// visibleIf. It is both a canonicalFixtures entry (closing Finding 8's
// round-trip coverage gap) and the runtime side of drift_test.go's
// behavioural set-equality assertion (D-1.4.15).
var maximalFixture = []byte(`{
  "assets": {
    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa": {
      "data": [
        "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk"
      ],
      "mediaType": "image/png"
    }
  },
  "bands": {
    "content": {
      "elements": [
        {
          "altRowBackground": "#EEEEEE",
          "as": "transaction",
          "bind": "transactions[]",
          "columns": [
            {
              "align": "left",
              "bind": "{{transaction.date}}",
              "id": "e3",
              "label": "Date",
              "width": 80
            },
            {
              "align": "right",
              "bind": "{{formatNumber(transaction.amount, \"#,##0.00\")}}",
              "footer": "sum",
              "footerFormat": "#,##0.00",
              "footerOf": "transactions.amount",
              "id": "e4",
              "label": "Amount",
              "width": 90
            }
          ],
          "headerHeight": 16,
          "id": "e1",
          "style": {
            "align": "left",
            "background": "#F1F4F7",
            "bold": true,
            "border": {
              "color": "#000000",
              "edges": [
                "bottom",
                "top"
              ],
              "width": 0.5
            },
            "fontFamily": "body",
            "fontSize": 8,
            "italic": true,
            "padding": {
              "bottom": 2,
              "left": 3,
              "right": 3,
              "top": 2
            },
            "valign": "top"
          },
          "type": "table",
          "visibleIf": "customer.hasTransactions",
          "x": 0,
          "y": 0
        },
        {
          "asset": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
          "height": 60,
          "id": "e5",
          "type": "image",
          "width": 100,
          "x": 0,
          "y": 200
        },
        {
          "height": 1,
          "id": "e6",
          "style": {
            "background": "#000000"
          },
          "type": "rect",
          "width": 500,
          "x": 0,
          "y": 300
        }
      ]
    },
    "pageFooter": {
      "elements": [
        {
          "height": 10,
          "id": "e7",
          "style": {
            "align": "center",
            "fontSize": 7
          },
          "type": "text",
          "value": "Page {{page}} of {{pages}}",
          "width": 523,
          "x": 0,
          "y": 8
        }
      ],
      "height": 30
    },
    "pageHeader": {
      "elements": [
        {
          "height": 16,
          "id": "e2",
          "style": {
            "bold": true,
            "fontSize": 12
          },
          "type": "text",
          "value": "Statement for {{customer.name}}",
          "width": 400,
          "x": 0,
          "y": 10
        }
      ],
      "height": 60
    }
  },
  "fonts": {
    "body": [
      "Noto Sans",
      "Noto Sans Thai",
      "Noto Sans SC"
    ]
  },
  "locale": "th",
  "nextId": 8,
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
  "version": "1.0"
}
`)

// nullFieldFixture carries an explicit JSON `null` on two Presence-typed
// fields (element.style, element.visibleIf) — D-1.4.16, this story's
// finisher review Finding 6 (Major): no canonical fixture previously
// carried an explicit null, so P1/P2/P3 never exercised that path as a
// fixed point. See TestPresenceThreePolarity's "serializes-differently"
// subtest (presence_test.go) for the direct absent/null/present
// serialization-distinctness assertion this fixture complements.
var nullFieldFixture = []byte(`{
  "assets": {},
  "bands": {
    "content": {
      "elements": []
    },
    "pageFooter": {
      "elements": [],
      "height": 20
    },
    "pageHeader": {
      "elements": [
        {
          "height": 14,
          "id": "e1",
          "style": null,
          "type": "text",
          "value": "v",
          "visibleIf": null,
          "width": 100,
          "x": 0,
          "y": 0
        }
      ],
      "height": 20
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
  "utcOffset": "+00:00",
  "version": "1.0"
}
`)
