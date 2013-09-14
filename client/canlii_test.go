package client

import (
	"fmt"
	"testing"
)

func TestGetDatabaseLise(t *testing.T) {
	dbList, err := DatabaseList()
	if err != nil {
		t.Fatal(err)
	}

	for _, db := range dbList {
		fmt.Printf("Found `%v`\n", db.Name)
	}
}

func TestCaseList(t *testing.T) {
	db := Database{ID: "abwcac"}
	cases, err := db.CaseList(0, 10)
	if err != nil {
		t.Fatal(err)
	}

	if len(cases) == 0 {
		t.Fatal("No cases returned")
	}

	for _, aCase := range cases {
		fmt.Printf("Found `%v`\n", aCase.Title)
	}
}
