# Exported API inventory

Observed from the local working tree at commit `3890ba8` on 2026-09-11. This is a coverage baseline, not the finished user-facing reference. Signatures retain Go doc spelling; explain internal type references and actual constant values in the final guide. Refresh at implementation time.

## folio

```go
const (
	MaxStandInDataPaths = 512
	MaxStandInDataBytes = 256 << 10
)
const DiagCodeBindingPathAbsent = string(diag.CodeBindingPathAbsent)
const DiagCodeContentUnlayoutable = string(diag.CodeContentUnlayoutable)
const DiagCodeDocumentDateInvalid = string(diag.CodeDocumentDateInvalid)
const DiagCodeEmptyAverage = string(diag.CodeEmptyAverage)
const DiagCodeExpressionInvalid = string(diag.CodeExpressionInvalid)
const DiagCodeInternalUnhandledCaveat = string(diag.CodeInternalUnhandledCaveat)
const DiagCodeStyleColorInvalid = string(diag.CodeStyleColorInvalid)
const DiagCodeStyleLineSpacingInvalid = string(diag.CodeStyleLineSpacingInvalid)
const DiagCodeTableFooterOrphanSuppressed = string(diag.CodeTableFooterOrphanSuppressed)
const DiagCodeTableFooterSourceForbidden = string(diag.CodeTableFooterSourceForbidden)
const DiagCodeTableFooterSourceUnresolved = string(diag.CodeTableFooterSourceUnresolved)
const DiagCodeTableHeaderRepeatSuppressed = string(diag.CodeTableHeaderRepeatSuppressed)
const DiagCodeTableRowClippedHeight = string(diag.CodeTableRowClippedHeight)
const DiagCodeTemplateFieldInvalid = string(diag.CodeTemplateFieldInvalid)
const DiagCodeTemplateMalformed = string(diag.CodeTemplateMalformed)
const DiagCodeTextClippedWidth = string(diag.CodeTextClippedWidth)
const DiagCodeTextMissingGlyph = string(diag.CodeTextMissingGlyph)
const DiagCodeTextStyleFaceUndeclared = string(diag.CodeTextStyleFaceUndeclared)
const GridIncrement int64 = 6000
const LocaleTableVersion = expr.LocaleTableVersion
const MaxCanvasMillipoints int64 = 9007199254740991
const MaxParameterReferenceNameLength = 128
const StandInInstant = "2024-01-15T12:00:00Z"
const Version = "0.0.0-dev"
func ApplyComponentCommand(t *Template, command []byte) (CanvasProjection, error)
func ApplyPageSetupCommand(t *Template, command []byte) (CanvasProjection, error)
func AssetBytes(t *Template, key string) ([]byte, string, error)
func Canvas(t *Template) (CanvasProjection, error)
func CanvasWithTextPaint(t *Template, fs FontSet) (CanvasProjection, error)
func LoadTemplate(path string) (*Template, error)
func ParameterReferences(tpl *Template) ([]string, error)
func ParseTemplate(b []byte) (*Template, error)
func PreviewIdentity(template []byte, data Data, params Params, fontSet FontSet) string
func Render(t *Template, d Data, p Params, f FontSet) (Result, error)
func RenderTo(w io.Writer, t *Template, d Data, p Params, f FontSet) ([]Diagnostic, error)
func SerializeTemplate(t *Template) ([]byte, error)
func SnapToGrid(proposed geom.Length) (geom.Length, bool)
func StandInData(tpl *Template) ([]byte, error)
func TableColumns(t *Template, tableID string) (TableColumnsProjection, error)
func Validate(b []byte, d Data, p Params, f FontSet) ([]Diagnostic, error)
type CanvasBand struct {
	Name   string `json:"name"`
	X      int64  `json:"x"`
	Y      int64  `json:"y"`
	Width  int64  `json:"width"`
	Height int64  `json:"height"`
}
type CanvasComponent struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Band      string `json:"band"`
	X         int64  `json:"x"`
	Y         int64  `json:"y"`
	Width     int64  `json:"width"`
	Height    int64  `json:"height"`
	Resizable bool   `json:"resizable"`
	Value *string `json:"value,omitempty"`
	Binding    *string `json:"binding,omitempty"`
	VisibleIf  *string `json:"visibleIf,omitempty"`
	FontFamily *string `json:"fontFamily,omitempty"`
	FontSize   *int64  `json:"fontSize,omitempty"`
	LineSpacing   *int64            `json:"lineSpacing,omitempty"`
	Bold          *bool             `json:"bold,omitempty"`
	Italic        *bool             `json:"italic,omitempty"`
	Align         *string           `json:"align,omitempty"`
	Valign        *string           `json:"valign,omitempty"`
	Background    *string           `json:"background,omitempty"`
	Color         *string           `json:"color,omitempty"`
	BorderWidth   *int64            `json:"borderWidth,omitempty"`
	BorderColor   *string           `json:"borderColor,omitempty"`
	BorderEdges   []string          `json:"borderEdges,omitempty"`
	TableBind     *string           `json:"tableBind,omitempty"`
	PaddingTop    *int64            `json:"paddingTop,omitempty"`
	PaddingRight  *int64            `json:"paddingRight,omitempty"`
	PaddingBottom *int64            `json:"paddingBottom,omitempty"`
	PaddingLeft   *int64            `json:"paddingLeft,omitempty"`
	TextPaint     *CanvasTextPaint  `json:"textPaint,omitempty"`
	Image         *CanvasImagePaint `json:"image,omitempty"`
	ImageUnavailable *string `json:"imageUnavailable,omitempty"`
	Columns []CanvasTableColumn `json:"columns,omitempty"`
}
type CanvasFontChain struct {
	Name string `json:"name"`
	Entries []CanvasFontChainEntry `json:"entries"`
}
type CanvasFontChainEntry struct {
	Face     string `json:"face"`
	AssetKey string `json:"assetKey"`
	Family   string `json:"family"`
	Style    string `json:"style"`
	Bold       string `json:"bold"`
	Italic     string `json:"italic"`
	BoldItalic string `json:"boldItalic"`
}
type CanvasImagePaint struct {
	MediaType  string `json:"mediaType"`
	AssetKey   string `json:"assetKey"`
	Width      int64  `json:"width"`
	Height     int64  `json:"height"`
	DrawX      int64  `json:"drawX"`
	DrawY      int64  `json:"drawY"`
	DrawWidth  int64  `json:"drawWidth"`
	DrawHeight int64  `json:"drawHeight"`
}
type CanvasProjection struct {
	Width  int64 `json:"width"`
	Height int64 `json:"height"`
	Locale        string            `json:"locale"`
	UTCOffset     string            `json:"utcOffset"`
	Orientation   string            `json:"orientation"`
	Preset        string            `json:"preset"`
	MarginTop     int64             `json:"marginTop"`
	MarginRight   int64             `json:"marginRight"`
	MarginBottom  int64             `json:"marginBottom"`
	MarginLeft    int64             `json:"marginLeft"`
	GridIncrement int64             `json:"gridIncrement"`
	CommandWidth  int64             `json:"commandWidth"`
	CommandHeight int64             `json:"commandHeight"`
	Bands         []CanvasBand      `json:"bands"`
	Components    []CanvasComponent `json:"components"`
	FontFamilies []string `json:"fontFamilies"`
	FontChains []CanvasFontChain `json:"fontChains"`
	DefaultFontSize int64 `json:"defaultFontSize"`
	DefaultLineSpacing int64 `json:"defaultLineSpacing"`
	ContentWindowHeight int64 `json:"contentWindowHeight"`
	ContentWindowCount int64 `json:"contentWindowCount"`
	ContentWindowOrigins []int64 `json:"contentWindowOrigins"`
	ContentWindowCountIsExact bool `json:"contentWindowCountIsExact"`
}
type CanvasTableColumn struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Width int64  `json:"width"`
	HeaderAlign string `json:"headerAlign"`
	CellAlign   string `json:"cellAlign"`
	Bind        string `json:"bind"`
}
type CanvasTextFragment struct {
	Text string `json:"text"`
	X    int64  `json:"x"`
	AssetKey string `json:"assetKey,omitempty"`
	Face string `json:"face,omitempty"`
}
type CanvasTextLine struct {
	Top       int64                `json:"top"`
	Baseline  int64                `json:"baseline"`
	Advance   int64                `json:"advance"`
	Width     int64                `json:"width"`
	Fragments []CanvasTextFragment `json:"fragments"`
}
type CanvasTextPaint struct {
	Overflow bool `json:"overflow"`
	Truncated bool             `json:"truncated"`
	Lines     []CanvasTextLine `json:"lines"`
}
type ComponentCommandError struct {
	ElementID string
	DataPath  string
	Message   string
	// Has unexported fields.
}
type Data []byte
type Diagnostic struct {
	Severity Severity
	Code string
	ElementID string
	DataPath string
	Message string
}
type FontSet map[string][]byte
type Params []byte
type RenderError struct {
	Diagnostic Diagnostic
	Err error
}
func (e *RenderError) Error() string
func (e *RenderError) Unwrap() error
type Result struct {
	Bytes       []byte
	Diagnostics []Diagnostic
}
type Severity int
const (
	SeverityWarning Severity
	SeverityError
)
func (s Severity) String() string
type TableColumnProjection struct {
	ID               string `json:"id"`
	Header           string `json:"header"`
	Width            int64  `json:"width"`
	Align            string `json:"align"`
	Binding          string `json:"binding"`
	RowField         string `json:"rowField"`
	RowFieldEditable bool   `json:"rowFieldEditable"`
	Footer           string `json:"footer"`
	FooterOf         string `json:"footerOf"`
	FooterFormat     string `json:"footerFormat"`
}
type TableColumnsProjection struct {
	TableID    string `json:"tableId"`
	Collection string `json:"collection"`
	Alias      string `json:"alias"`
	HeaderHeight     int64  `json:"headerHeight"`
	AltRowBackground string `json:"altRowBackground"`
	HeaderFontFamily          string `json:"headerFontFamily"`
	HeaderFontFamilyResolved  string `json:"headerFontFamilyResolved"`
	HeaderFontSize            int64  `json:"headerFontSize"`
	HeaderFontSizeResolved    int64  `json:"headerFontSizeResolved"`
	HeaderLineSpacing         int64  `json:"headerLineSpacing"`
	HeaderLineSpacingResolved int64  `json:"headerLineSpacingResolved"`
	HeaderBackground          string `json:"headerBackground"`
	HeaderBackgroundResolved  string `json:"headerBackgroundResolved"`
	HeaderColor               string `json:"headerColor"`
	HeaderColorResolved       string `json:"headerColorResolved"`
	HeaderValign              string `json:"headerValign"`
	HeaderValignResolved      string `json:"headerValignResolved"`
	HeaderAlign               string `json:"headerAlign"`
	HeaderAlignResolved       string `json:"headerAlignResolved"`
	HeaderBold                bool   `json:"headerBold"`
	HeaderBoldResolved        bool   `json:"headerBoldResolved"`
	HeaderItalic              bool   `json:"headerItalic"`
	HeaderItalicResolved      bool   `json:"headerItalicResolved"`
	HeaderBorderWidth         string `json:"headerBorder.width"`
	HeaderBorderWidthResolved string `json:"headerBorder.widthResolved"`
	HeaderBorderColor         string `json:"headerBorder.color"`
	HeaderBorderColorResolved string `json:"headerBorder.colorResolved"`
	HeaderBorderEdges         string `json:"headerBorder.edges"`
	HeaderBorderEdgesResolved string `json:"headerBorder.edgesResolved"`
	Columns []TableColumnProjection `json:"columns"`
}
type Template struct {
	// Has unexported fields.
}
```

## fonts

```go
func Shipped() folio.FontSet
```

## wasm

```go
var ErrNoRedo = errors.New("folio wasm: no redo history")
var ErrNoUndo = errors.New("folio wasm: no undo history")
type Engine struct {
	// Has unexported fields.
}
func NewEngine() *Engine
func (e *Engine) Apply(command []byte) (Snapshot, error)
func (e *Engine) AssetBytes(key string) ([]byte, Snapshot, error)
func (e *Engine) Initialize(input []byte) (Snapshot, error)
func (e *Engine) Load(input []byte) (Snapshot, error)
func (e *Engine) ParameterReferences() ([]string, uint64, error)
func (e *Engine) PreviewIdentity(data, params []byte) (string, uint64, error)
func (e *Engine) Redo() (Snapshot, error)
func (e *Engine) Render(template, data, params []byte) ([]byte, RenderResult, error)
func (e *Engine) Serialize() ([]byte, Snapshot, error)
func (e *Engine) Snapshot() Snapshot
func (e *Engine) StandInData() ([]byte, error)
func (e *Engine) TableColumns(tableID string) (TableColumnsResult, error)
func (e *Engine) Undo() (Snapshot, error)
func (e *Engine) Validate() (Snapshot, error)
type RenderResult struct {
	PDFSHA256   string             `json:"pdfSha256"`
	Identity    string             `json:"identity"`
	Revision    uint64             `json:"revision"`
	Diagnostics []folio.Diagnostic `json:"diagnostics"`
	ElapsedMs   int64              `json:"elapsedMs"`
	Version     string             `json:"version"`
}
type Snapshot struct {
	DocumentState string                  `json:"documentState"`
	Revision      uint64                  `json:"revision"`
	ByteLength    int                     `json:"byteLength"`
	CanUndo       bool                    `json:"canUndo"`
	CanRedo       bool                    `json:"canRedo"`
	Canvas        *folio.CanvasProjection `json:"canvas,omitempty"`
}
type TableColumnsResult struct {
	Revision uint64                       `json:"revision"`
	Table    folio.TableColumnsProjection `json:"table"`
}
```
