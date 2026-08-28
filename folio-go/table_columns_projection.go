package folio

import (
	"fmt"

	"github.com/panitw/folio/folio-go/internal/template"
)

const maxTableColumns = 128

// TableColumnsProjection is a bounded paint/editor projection. It deliberately
// omits table bindings, footer schema, canonical bytes, and every field that
// is not editable by Story 6.4.
type TableColumnsProjection struct {
	TableID string                  `json:"tableId"`
	Columns []TableColumnProjection `json:"columns"`
}

type TableColumnProjection struct {
	ID     string `json:"id"`
	Header string `json:"header"`
	Width  int64  `json:"width"`
	Align  string `json:"align"`
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
	projection := TableColumnsProjection{TableID: tableID, Columns: make([]TableColumnProjection, 0, len(element.Table.Value.Columns))}
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
		projection.Columns = append(projection.Columns, TableColumnProjection{ID: string(column.ID), Header: column.Label, Width: int64(column.Width), Align: align})
	}
	return projection, nil
}
