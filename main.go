package main

import (
	"flag"
	"fmt"

	"github.com/tyrust/tirc/irc"
	"time"
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

	addr := fmt.Sprintf("%s:%d", *host, *port)
	err := c.Connect(&addr, password, in)

	if err != nil {
		fmt.Printf("ERROR: %s", err)
	}

	c.Send(*irc.NewJoinMessage(*c.Prefix, []string{"#general"}, nil))
	c.Send(*irc.NewPrivateMessage(*c.Prefix, "#general", "hello #general"))
	c.Send(*irc.NewPrivateMessage(*c.Prefix, "tyrus", "hello tyrus"))
	c.Quit("later nerds")
	go reader(in)
	//time.Sleep(time.Second * 5)
	//c.Disconnect()

}

func reader(in <-chan irc.Message) {
	for m := range in {
		fmt.Printf("R: %s", m.String())
	}
}
