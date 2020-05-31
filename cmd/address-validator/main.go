package main

import (
	"github.com/NavExplorer/navexplorer-indexer-go/internal/validator"
)

func main() {
	new(validator.AddressValidator).Execute()
}
