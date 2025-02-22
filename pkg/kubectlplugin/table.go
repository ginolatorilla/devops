package kubectlplugin

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/duration"
)

var ResourceKindColumn = metaV1.TableColumnDefinition{
	Name:        "Kind",
	Type:        "string",
	Description: "Kubernetes resource kind",
}

type TableBuilder struct {
	Table       metaV1.Table
	columnIndex map[string]int
}

func NewTableBuilder() *TableBuilder {
	defaultColumns := []metaV1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metaV1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "Age", Type: "string", Description: metaV1.ObjectMeta{}.SwaggerDoc()["creationTimestamp"]},
	}

	builder := TableBuilder{
		Table: metaV1.Table{
			TypeMeta: metaV1.TypeMeta{
				APIVersion: "meta.k8s.io/v1",
				Kind:       "Table",
			},
			ColumnDefinitions: defaultColumns,
			Rows:              []metaV1.TableRow{},
		},
		columnIndex: make(map[string]int, len(defaultColumns)),
	}

	for i, column := range defaultColumns {
		builder.columnIndex[column.Name] = i
	}

	return &builder
}

func (tb *TableBuilder) AdditionalColumns(columns ...metaV1.TableColumnDefinition) *TableBuilder {
	offset := len(tb.Table.ColumnDefinitions)
	for i, column := range columns {
		tb.Table.ColumnDefinitions = append(tb.Table.ColumnDefinitions, column)
		tb.columnIndex[column.Name] = i + offset
	}
	return tb
}

func (tb *TableBuilder) AddRow(object runtime.Object, cells map[string]any) {
	row := metaV1.TableRow{
		Cells:  make([]any, len(tb.Table.ColumnDefinitions)),
		Object: runtime.RawExtension{Object: object},
	}
	if acc, err := meta.Accessor(object); err == nil && acc != nil {
		row.Cells[tb.columnIndex["Name"]] = acc.GetName()
		row.Cells[tb.columnIndex["Age"]] = translateTimestampSince(acc.GetCreationTimestamp())
	}
	for column, cell := range cells {
		if index, ok := tb.columnIndex[column]; !ok {
			panic(fmt.Errorf("column %q not found", column))
		} else {
			row.Cells[index] = cell
		}
	}

	tb.Table.Rows = append(tb.Table.Rows, row)
}

func translateTimestampSince(timestamp metaV1.Time) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}

	return duration.HumanDuration(time.Since(timestamp.Time))
}
