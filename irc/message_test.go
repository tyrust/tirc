package irc

import (
	"testing"

	"github.com/tyrust/tirc/irc"
)

func TestShit(t *testing.T) {
	var m *irc.Message
	//m = irc.ParseMessage(":guy!u@host.com PRIVMSG :bababa asdf asdf")
	//m = irc.ParseMessage(":asdf.com NICK bob")
	//m = irc.ParseMessage("USER guest 0 * :Tyrus Yeah")
	m = irc.ParseMessage(":irc.example.net 265 botn :2 2 Current local users: 2, Max: 2")
	//m = irc.ParseMessage("PRIVMSG #asdf :Kappa slapp")
	//m = irc.ParseMessage("PONG poop.com")
	//m = irc.ParseMessage("PASS asdf")
	//m = irc.ParseMessage(":botn!~botu@localhost JOIN :#asdf")
	switch m.Command {
	default:
		t.Log(m)
	// case "NICK":
	// 	t.Logf("nick is %s\n", *m.Params["nick"])
	case "USER":
		t.Logf("u:%s m:%s n:\"%s\"\n", *m.Params["user"], *m.Params["mode"], *m.Params["realname"])
		// case "265":
		// 	t.Logf("n:%s t:%s r:%s\n", m.Command, *m.Params["target"], *m.Params["reply"])
	}
}
