package main

import (
	"reflect"
	"testing"

	"github.com/jbrukh/bayesian"
	_ "github.com/mattn/go-sqlite3"
)

func Test_getLabels(t *testing.T) {

	var database DatabaseSQLite = DatabaseSQLite{path: "../db/main.db", db: nil, rows: nil}
	tests := []struct {
		name    string
		want    []bayesian.Class
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := database.getLabels()
			if (err != nil) != tt.wantErr {
				t.Errorf("getLabels() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}
