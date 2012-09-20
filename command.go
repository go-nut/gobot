package main

import (
  "container/list"
  "strings"
)

type callback func(*IRCEvent)

func (irc *IRC) AddCallback(key string, value callback) {
  l, ok := irc.Callbacks[key]
  if !ok || l == nil {
    l = list.New()
    irc.Callbacks[key] = l
  }
  l.PushBack(value)
}

func (irc *IRC) addcallbacks() {
  if irc.Callbacks == nil {
    irc.Callbacks = make(map[string]*list.List)
  }

  // Ping
  irc.AddCallback("PING", func(e *IRCEvent) {
    irc.SendRaw("Pong :%s" + strings.Join(e.Args, " "))
  })

  irc.AddCallback("PRIVMSG", func(e *IRCEvent) {
    irc.Privmsg(e.Args[0], strings.Join(e.Args[1:], " "))
  })
}

func (irc *IRC) runcallback(e *IRCEvent) {
  if l, ok := irc.Callbacks[e.Command]; ok {
    for v := l.Front(); v != nil; v = v.Next() {
      v.Value.(callback)(e)
    }
  }
}
