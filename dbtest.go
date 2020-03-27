package main

import (
	"fmt"

	"github.com/miska12345/DDPoll/db"
)

func main() {
	_, err := db.Dial("mongodb+srv://admin:wassup@cluster0-n0w7a.mongodb.net/test?retryWrites=true&w=majority", 2, 5)
	fmt.Println(err)
}
