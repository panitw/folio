package pdf

// appendImageContentStream appends one q/cm/Do/Q sequence per image
// placement on the page (AC18: the image is drawn with q/cm/Do/Q,
// operands routed by D-1.1.b — cm's operands are geometric, so they go
// through appendLength, never appendInt or a bare literal). imageIDs is
// only used to confirm every placement's ResourceName actually names an
// embedded XObject (a document-assembly invariant, mirroring
// buildTextContentStream's missingFaceError check for run.Face).
func appendImageContentStream(dst []byte, page TextPage, imageIDs map[string]int64) ([]byte, error) {
	for _, im := range page.Images {
		if _, ok := imageIDs[im.ResourceName]; !ok {
			return nil, &missingImageError{resourceName: im.ResourceName}
		}

		imgX := page.MarginLeft + im.X
		imgY := flipY(page.Height, page.MarginTop, im.Y, im.DrawHeight)

		dst = append(dst, "q\n"...)
		dst = appendLength(dst, im.DrawWidth)
		dst = append(dst, " 0 0 "...)
		dst = appendLength(dst, im.DrawHeight)
		dst = append(dst, ' ')
		dst = appendLength(dst, imgX)
		dst = append(dst, ' ')
		dst = appendLength(dst, imgY)
		dst = append(dst, " cm\n/"...)
		dst = append(dst, pdfNameEscape(im.ResourceName)...)
		dst = append(dst, " Do\nQ\n"...)
	}
	return dst, nil
}

type missingImageError struct{ resourceName string }

func (e *missingImageError) Error() string {
	return "internal/pdf: image placement names resource " + e.resourceName + ", not present in the document's image XObject set"
}
