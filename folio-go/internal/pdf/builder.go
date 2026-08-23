package pdf

import "github.com/panitw/folio/folio-go/internal/geom"

// builder is a small, deterministic multi-object PDF writer, generalising
// document.go's fixed four-object structure to an arbitrary object count
// (needed once font and page objects are dynamic in number). It follows
// exactly the same rules Story 1.1/1.2 established: every geometric
// number goes through appendLength, every count/offset/object/generation
// number through appendInt/appendIntPadded (AD-3), classic
// (non-stream) cross-reference table, content-derived /ID (AD-7), no
// compression, no /Info, no /CreationDate or /ModDate.
type builder struct {
	body    []byte
	offsets []int // offsets[objNum], index 0 unused (the free entry)
	nextID  int64
}

func newBuilder() *builder {
	b := &builder{offsets: []int{0}, nextID: 1}
	b.body = append(b.body, "%PDF-1.7\n%"...)
	b.body = append(b.body, binaryComment...)
	b.body = append(b.body, '\n')
	return b
}

// reserve allocates the next object number without writing anything yet
// — callers reserve every object id up front so forward references
// (e.g. a Catalog referencing Pages before Pages is written) can be
// spelled out immediately.
func (b *builder) reserve() int64 {
	b.offsets = append(b.offsets, 0)
	id := b.nextID
	b.nextID++
	return id
}

// begin starts object id: records its offset and writes "id 0 obj\n".
func (b *builder) begin(id int64) {
	b.offsets[id] = len(b.body)
	b.body = appendObjHeader(b.body, id)
}

// end closes the object most recently begun.
func (b *builder) end() {
	b.body = appendObjFooter(b.body)
}

func (b *builder) write(p []byte) {
	b.body = append(b.body, p...)
}

func (b *builder) writeRef(id int64) {
	b.body = appendRef(b.body, id)
}

func (b *builder) writeInt(v int64) {
	b.body = appendInt(b.body, v)
}

func (b *builder) writeLength(v geom.Length) {
	b.body = appendLength(b.body, v)
}

// finish appends the classic cross-reference table and trailer (with a
// content-derived /ID, AD-7) and returns the complete document.
func (b *builder) finish() []byte {
	xrefOffset := len(b.body)
	b.body = appendXrefGeneral(b.body, b.offsets)

	size := int64(len(b.offsets))
	b.body = append(b.body, "trailer\n<< /Size "...)
	b.body = appendInt(b.body, size)
	b.body = append(b.body, " /Root "...)
	b.body = appendRef(b.body, 1)
	b.body = append(b.body, ' ')

	idHex := computeID(b.body)
	b.body = append(b.body, "/ID ["...)
	b.body = append(b.body, '<')
	b.body = append(b.body, idHex...)
	b.body = append(b.body, '>')
	b.body = append(b.body, '<')
	b.body = append(b.body, idHex...)
	b.body = append(b.body, ">] >>\nstartxref\n"...)
	b.body = appendInt(b.body, int64(xrefOffset))
	b.body = append(b.body, "\n%%EOF\n"...)

	return b.body
}

// appendXrefGeneral is appendXref generalised to an arbitrary object
// count (document.go's appendXref stays fixed at exactly Story 1.1's
// four objects — untouched, per AC14a).
func appendXrefGeneral(dst []byte, offsets []int) []byte {
	n := int64(len(offsets))
	dst = append(dst, "xref\n0 "...)
	dst = appendInt(dst, n)
	dst = append(dst, '\n')

	dst = appendIntPadded(dst, 0, 10)
	dst = append(dst, ' ')
	dst = appendIntPadded(dst, 65535, 5)
	dst = append(dst, " f \n"...)

	for i := 1; i < len(offsets); i++ {
		off := int64(offsets[i])
		if off > 9999999999 {
			panic("internal/pdf: xref offset exceeds the 10-digit fixed field width")
		}
		dst = appendIntPadded(dst, off, 10)
		dst = append(dst, ' ')
		dst = appendIntPadded(dst, 0, 5)
		dst = append(dst, " n \n"...)
	}
	return dst
}
