package archivar

import (
	"fmt"
	"strconv"
	"time"
)

const (
	version string = "0.0.1"
)

// Main Wrapper for a set of tests
func Start(set_file string) {

	var s Setfile
	var total float32

	s.ReadFile(set_file)

	LogMessage(fmt.Sprintf("Archivar %v\n", version), "green")

	fmt.Printf("Set: %s\n", s.Description)

	// Loop over different challenges
	for _, entry := range s.Cards {
		card := fetch(entry)
		fmt.Println(card)
		t, _ := strconv.ParseFloat(card.Prices.Eur.(string), 32)
		total = total + float32(t)
		time.Sleep(100 * time.Millisecond)
	}

	// build new valueset
	v := &Value{}
	v.Date = time.Now().Format("2006-01-02 15:04:05")
	v.Value = total

	// add new valueset to set
	s.Value = append(s.Value, *v)

	LogMessage(fmt.Sprintf("Total value in this set %.2f", total), "green")

	s.Write(set_file)

}
