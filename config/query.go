package config

import(
	"errors"
	"encoding/json"
	"fmt"
	"bufio"
	"os"
	"strings"
	"regexp"
	"strconv"
	"github.com/aki2o/esa-cui/util"
)

type query struct {}

func init() {
	registProcessor(func() util.Processable { return &query{} }, "query", "Configure query.", "")
}

func (self *query) Do(args []string) error {
	var action_name string = ""

	if len(args) > 0 { action_name = args[0] }

	switch action_name {
	case "list":	return printQuery()
	case "add":		return addQuery()
	case "remove":	return removeQuery()
	default:		return errors.New("Unknown action!")
	}
}

func printQuery() error {
	for _, query := range Team.Queries {
		bytes, err := json.MarshalIndent(query, "", "\t")
		if err != nil { return err }

		fmt.Println(string(bytes))
	}
	return nil
}

func addQuery() error {
	name := scanName(true)
	
	entries_string := scanValue()
	if entries_string == "" { return errors.New("Require value!") }
	
	fuzzy_key_re, _ := regexp.Compile("[a-z]: +")
	trimmer := func (s string) string { return strings.TrimSpace(s) }
	entries_string = fuzzy_key_re.ReplaceAllStringFunc(entries_string, trimmer)
	
	entry_re, _					:= regexp.Compile("^(-?[a-z]+):(.+)$")
	beginning_of_quote_re, _	:= regexp.Compile("^['\"]")
	end_of_quote_re, _			:= regexp.Compile("['\"]$")
	
	query := &Query{ Name: name }

	for _, entry_str := range strings.Fields(entries_string) {
		tokens := entry_re.FindStringSubmatch(entry_str)
		if len(tokens) != 3 { continue }

		entry_key := tokens[1]
		entry_value := beginning_of_quote_re.ReplaceAllString(tokens[2], "")
		entry_value = end_of_quote_re.ReplaceAllString(entry_value, "")
		
		entry := &QueryEntry{ Key: entry_key, Value: entry_value }

		query.Entries = append(query.Entries, *entry)
	}

	Team.Queries = append(Team.Queries, *query)
	Save()
	return nil
}

func removeQuery() error {
	name := scanName(false)
	if name == "" { return errors.New("Require name!") }

	index := indexQuery(name)
	next_index := index + 1
	if index < 0 { return errors.New("Not found query of '"+name+"'!") }
	
	Team.Queries = append(Team.Queries[:index], Team.Queries[next_index:]...)
	Save()
	return nil
}

func scanName(required bool) string {
	fmt.Print("Name: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	name := strings.TrimSpace(scanner.Text())

	if name != "" { return name }
	if ! required { return name }

	index := 1
	for {
		name = "query"+strconv.Itoa(index)
		
		if indexQuery(name) < 0 { return name }
		
		index = index + 1
	}
}

func scanValue() string {
	fmt.Print("Value: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	
	return scanner.Text()
}

func indexQuery(name string) int {
	for index, query := range Team.Queries {
		if query.Name != name { continue }

		return index
	}
	return -1
}
