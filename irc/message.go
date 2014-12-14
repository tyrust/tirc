package irc

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

type ParamDef map[int]string

var commandToParamDef = map[string]ParamDef{
	// Connection Registration
	"PASS": ParamDef{0: "password"},
	"NICK": ParamDef{0: "nick"},
	"USER": ParamDef{0: "user", 1: "mode", 2: "_", 3: "realname"},
	"QUIT": ParamDef{0: "message"},
	// Channel Operations
	"JOIN": ParamDef{0: "channels", 1: "keys"},
	// Sending Messages
	"PRIVMSG": ParamDef{0: "msgtarget", 1: "text"},
	// Miscellaneous Messages
	"PING": ParamDef{0: "server1", 1: "server2"},
	"PONG": ParamDef{0: "server", 1: "server2"},
}

//var commandToParamDef = make(map[string]ParamDef)

var replyParamDef = ParamDef{0: "target", 1: "reply"}

type Prefix struct {
	Nick string
	Host string
	User string
}

type Message struct {
	*Prefix
	Command string
	params  []string
	Params  map[string]*string
}

func NewMessage(p Prefix, command string, params map[string]string) *Message {
	paramDef := getParamDef(command)
	if paramDef == nil {
		log.Fatal("Unrecognized command: %v", command)
	}
	params["_"] = "*" // For unused fields (e.g. rfc2812 3.1.3).
	m := Message{&p, command,
		make([]string, len(paramDef)),
		make(map[string]*string, len(paramDef))}
	for idx, field := range paramDef {
		if pValue, ok := params[field]; ok && pValue != "" {
			m.params[idx] = pValue
			m.Params[field] = &m.params[idx]
		}
	}
	return &m
}

func ParseMessage(s string) *Message {
	// FIXME: handle empty
	if len(s) == 0 {
		return new(Message)
	}
	s = strings.TrimSpace(s)
	var p *Prefix
	// Prefix
	if s[:1] == ":" {
		p = ParsePrefix(s[1:strings.Index(s, " ")])
		s = s[strings.Index(s, " ")+1:]
	} else {
		p = new(Prefix)
	}
	// Command
	cmdIdx := strings.Index(s, " ")
	command := s[:cmdIdx]
	s = s[cmdIdx+1:]
	// Params
	colonSplit := strings.SplitN(s, ":", 2)
	if colonSplit[0] == "" {
		colonSplit = colonSplit[1:]
	}
	// The "middle" params
	params := strings.Split(strings.TrimSpace(colonSplit[0]), " ")
	if len(colonSplit) > 1 {
		// The "trailing" param
		params = append(params, colonSplit[1])
	}
	m := Message{p, command, params, make(map[string]*string)}

	paramDef := getParamDef(m.Command)
	if isReply(m.Command) {
		// Coalesce reply.
		m.params[1] = strings.Join(m.params[1:], " ")
		m.params = m.params[:2]
	}
	// Fill Params
	nParams := len(m.params)
	for i, param := range paramDef {
		if i < nParams {
			m.Params[param] = &m.params[i]
		}
	}
	return &m
}

// String returns the String representation of a Message in the format
// specified by RFC 2812.  See
// https://tools.ietf.org/html/rfc2812#section-3
func (m *Message) String() string {
	prefix := m.Prefix.String()
	if prefix != "" {
		prefix = ":" + prefix
	}
	last := len(m.params) - 1
	for i, s := range m.params {
		if s == "" {
			last = i - 1
			break
		}
	}
	params := ""
	if last >= 0 {
		if last > 0 {
			params = " "
			params += strings.Join(m.params[:last], " ")
		}
		params += " :" + m.params[last]
	}
	return fmt.Sprintf("%s %s%s\r\n", prefix, m.Command, params)
}

func ParsePrefix(s string) *Prefix {
	// servername / ( nickname [ [ "!" user ] "@" host ] )
	var nick, host, user string
	atSplit := strings.Split(s, "@")
	bangSplit := strings.Split(atSplit[0], "!")
	nick = bangSplit[0]
	if len(atSplit) > 1 {
		host = atSplit[1]
		if len(bangSplit) > 1 {
			user = bangSplit[1]
		}
	}
	return &Prefix{nick, host, user}
}

func (p *Prefix) String() string {
	var s string
	if p.Nick != "" {
		s = p.Nick
		if p.Host != "" {
			if p.User != "" {
				s += "!" + p.User
			}
			s += "@" + p.Host
		}
	}
	return s
}

// Given a command string, return the appropriate ParamDef.
func getParamDef(command string) ParamDef {
	var pmap ParamDef
	// TODO idiomatic way?
	if pm, ok := commandToParamDef[command]; ok {
		pmap = pm
	} else if isReply(command) {
		pmap = replyParamDef
	}
	return pmap
}

func isReply(command string) bool {
	_, err := strconv.Atoi(command)
	return err == nil
}

// Message Makers
type paramMap map[string]string

// PASS => ParamDef{0: "password"}
// https://tools.ietf.org/html/rfc2812#section-3.1.1
func NewPassMessage(p Prefix, pw string) *Message {
	return NewMessage(p, "PASS",
		paramMap{"password": pw})
}

// NICK => ParamDef{0: "nick"}
// https://tools.ietf.org/html/rfc2812#section-3.1.2
func NewNickMessage(p Prefix, n string) *Message {
	return NewMessage(p, "NICK",
		paramMap{"nick": n})
}

// USER => ParamDef{0: "user", 1: "mode", 2: "_", 3: "realname"}
// https://tools.ietf.org/html/rfc2812#section-3.1.3
func NewUserMessage(p Prefix, u string, m string, rn string) *Message {
	return NewMessage(p, "USER",
		paramMap{"user": u, "mode": m, "realname": rn})
}

// QUIT => ParamDef{0: "message"}
// https://tools.ietf.org/html/rfc2812#section-3.1.7
func NewQuitMessage(p Prefix, m string) *Message {
	return NewMessage(p, "QUIT",
		paramMap{"message": m})
}

// JOIN => ParamDef{0: "channels", 1: "keys"}
// https://tools.ietf.org/html/rfc2812#section-3.2.1
func NewJoinMessage(p Prefix, chans []string, keys []string) *Message {
	if len(keys) > len(chans) {
		log.Fatal("Too many keys (%v) to join (%v) channels.", len(keys), len(chans))
	}
	pm := make(paramMap, 2)
	if len(chans) == 0 {
		pm["channels"] = "0"
	} else {
		pm["channels"] = strings.Join(chans, ",")
	}
	pm["keys"] = strings.Join(keys, ",")
	return NewMessage(p, "JOIN", pm)
}

// PRIVMSG => ParamDef{0: "msgtarget", 1: "text"}
// https://tools.ietf.org/html/rfc2812#section-3.3.1
func NewPrivateMessage(p Prefix, mt string, t string) *Message {
	return NewMessage(p, "PRIVMSG",
		paramMap{"msgtarget": mt, "text": t})
}

// PING => ParamDef{0: "server1", 1: "server2"}
// https://tools.ietf.org/html/rfc2812#section-3.7.2
func NewPingMessage(p Prefix, s1 string, s2 string) *Message {
	return NewMessage(p, "PING",
		paramMap{"server1": s1, "server2": s2})
}

// PONG => ParamDef{0: "server", 1: "server2"}
// https://tools.ietf.org/html/rfc2812#section-3.7.3
func NewPongMessage(p Prefix, s string, s2 string) *Message {
	return NewMessage(p, "PONG",
		paramMap{"server": s, "server2": s2})
}
