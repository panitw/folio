package folio

import (
	"fmt"

	"github.com/panitw/folio/folio-go/internal/expr"
	"github.com/panitw/folio/folio-go/internal/template"
)

const maxTableColumns = 128

// TableColumnsProjection is a bounded paint/editor projection. It deliberately
// carries only the bounded table configuration the focused editor can paint.
// It never contains canonical bytes, sample data, parsed sample shape, or
// aggregate values.
type TableColumnsProjection struct {
	TableID    string                  `json:"tableId"`
	Collection string                  `json:"collection"`
	Alias      string                  `json:"alias"`
	Columns    []TableColumnProjection `json:"columns"`
}

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

// TableColumns returns only the selected table's structural column paint
// state. The browser may display it but cannot use it as a template model.
func TableColumns(t *Template, tableID string) (TableColumnsProjection, error) {
	if t == nil {
		return TableColumnsProjection{}, errNilTemplate
	}
	if tableID == "" || len(tableID) > 128 {
		return TableColumnsProjection{}, fmt.Errorf("folio: table id is invalid")
	}
	_, _, _, element, err := findComponent(t, tableID)
	if err != nil {
		return TableColumnsProjection{}, fmt.Errorf("folio: table was not found")
	}
	if element.Type != template.ElementTable || !element.Table.Set || element.Table.Null {
		return TableColumnsProjection{}, fmt.Errorf("folio: component is not a table")
	}
	if len(element.Table.Value.Columns) > maxTableColumns {
		return TableColumnsProjection{}, fmt.Errorf("folio: table has too many columns for editor projection")
	}
	collection := element.Table.Value.Bind
	alias := "row"
	if element.Table.Value.As.Set && !element.Table.Value.As.Null {
		alias = element.Table.Value.As.Value
	}
	if collection == "" || !rootCollectionPath.MatchString(collection) || len(collection) > 256 || alias == "" || len(alias) > 64 || !boundedIdentifier.MatchString(alias) || reservedRowAlias(alias) {
		return TableColumnsProjection{}, fmt.Errorf("folio: table configuration cannot be projected")
	}
	projection := TableColumnsProjection{TableID: tableID, Collection: collection, Alias: alias, Columns: make([]TableColumnProjection, 0, len(element.Table.Value.Columns))}
	for _, column := range element.Table.Value.Columns {
		if len(column.Label) > 256 || column.Width <= 0 || len(column.ID) == 0 || len(column.ID) > 128 {
			return TableColumnsProjection{}, fmt.Errorf("folio: table column cannot be projected")
		}
		align := "left"
		if column.Align.Set && !column.Align.Null {
			align = column.Align.Value
		}
		if align != "left" && align != "center" && align != "right" {
			return TableColumnsProjection{}, fmt.Errorf("folio: table column cannot be projected")
		}
		footer, footerOf, footerFormat := "", "", ""
		if column.Footer.Set && !column.Footer.Null {
			footer = column.Footer.Value
		}
		if column.FooterOf.Set && !column.FooterOf.Null {
			footerOf = column.FooterOf.Value
		}
		if column.FooterFormat.Set && !column.FooterFormat.Null {
			footerFormat = column.FooterFormat.Value
		}
		if footer != "" && footer != "sum" && footer != "avg" && footer != "count" || (footer == "" && (footerOf != "" || footerFormat != "")) || (footer == "count" && footerOf != "") || len(column.Bind) > 256 || len(footerOf) > 256 || len(footerFormat) > 256 {
			return TableColumnsProjection{}, fmt.Errorf("folio: table column cannot be projected")
		}
		// A new column has no bind yet. It is deliberately editable so the
		// normal Table Editor can complete it through the Go command boundary;
		// once a non-empty expression exists, retain the stricter projection
		// rules for arbitrary/unsupported expressions.
		row := expr.RowBinding{Editable: column.Bind == ""}
		if column.Bind != "" {
			row, err = expr.ProjectRowBinding(column.Bind, alias)
			if err != nil {
				return TableColumnsProjection{}, fmt.Errorf("folio: table column cannot be projected")
			}
		}
		projection.Columns = append(projection.Columns, TableColumnProjection{ID: string(column.ID), Header: column.Label, Width: int64(column.Width), Align: align, Binding: column.Bind, RowField: row.Field, RowFieldEditable: row.Editable, Footer: footer, FooterOf: footerOf, FooterFormat: footerFormat})
	}
	return projection, nil
}
