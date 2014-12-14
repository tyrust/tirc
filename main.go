package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/tyrust/tirc/irc"
)

var (
	host     = flag.String("host", "localhost", "Server hostname.")
	port     = flag.Int("port", 6667, "Server port.")
	user     = flag.String("user", "botu", "Username.")
	nick     = flag.String("nick", "botn", "Nick.")
	name     = flag.String("name", "cool guy", "Real name.")
	password = flag.String("password", "", "Password.")
)

func main() {
	flag.Parse()

	localhost := "poop.com"
	c := irc.NewClient(user, nick, name, &localhost)

	in := make(chan irc.Message, 25)
	go reader(in)

	addr := fmt.Sprintf("%s:%d", *host, *port)
	err := c.Connect(&addr, password, in)

	if err != nil {
		fmt.Printf("ERROR: %s", err)
	}

	c.Send(*irc.NewJoinMessage(*c.Prefix, []string{"#general"}, nil))
	for {
		time.Sleep(time.Second * 60)
	}
}

func reader(in <-chan irc.Message) {
	for m := range in {
		fmt.Printf("R: %s", m.String())
	}
}
