package action

import (
	"testing"
	"strings"
)

func TestOverwrite(t *testing.T) {
	write_process := &write{}

	body := "This is a pen."
	write_text := "That is a paper."

	new_body, err := write_process.makeBody(body, []string{ write_text })
	if err != nil {
		t.Errorf("Failed to overwrite : %s", err.Error)
	}
	if new_body != write_text {
		t.Errorf("Failed to overwrite. expect[%s] actual[%s]", write_text, new_body)
	}
}

func TestAppend(t *testing.T) {
	write_process := &write{ Appending: true }

	body := "This is a pen."
	write_text := "That is a paper."

	new_body, err := write_process.makeBody(body, []string{ write_text })
	if err != nil {
		t.Errorf("Failed to append : %s", err.Error)
	}
	if new_body != body+write_text {
		t.Errorf("Failed to append. expect[%s] actual[%s]", body+write_text, new_body)
	}
}

func TestInsert(t *testing.T) {
	write_process := &write{ InsertConditions: []string{ "two" } }

	body := []string{
		"# one",
		"# two",
		"# three",
	}
	write_text := "Hey!!!"
	expect_body := []string{
		"# one",
		"# two",
		"Hey!!!",
		"# three",
	}

	new_body, err := write_process.makeBody(strings.Join(body, "\r\n"), []string{ write_text })
	if err != nil {
		t.Errorf("Failed to insert : %s", err.Error)
	}
	if new_body != strings.Join(expect_body, "\r\n") {
		t.Errorf("Failed to insert.\n[expect]\n%s\n[actual]\n%s", strings.Join(expect_body, "\r\n"), new_body)
	}
}
