package main

import (
  "fmt"
  "bufio"
  "net"
  "os"
  "log"
)


type IRCEvent struct {
  Code string
  Message string
  Raw string
  Nick string //<nick>
  Host string //<nick>!<usr>@<host>
  Source string //<host>
  User string //<usr>

  Arguments []string
}


type IRC struct {
  conn net.Conn
  stdlog, errlog *log.Logger

  read, write chan string // Raw in/out
  readsync, writesync chan bool
  err chan error

  nick string
  server string
  quit bool
}

func writer(irc *IRC)  {
  for output := range irc.write {
    // If conn is nil, log
    if irc.conn == nil {
      irc.stdlog.Println("irc.conn is nil when writing")
      break
    }
    // If nothing is to be sent, don't send it
    if output == "" {
      break
    }
    // Attempt to write
    if _, err := irc.conn.Write([]byte(output)); err != nil {
      irc.errlog.Printf("Error in writer(): &s\n", err)
      irc.err <- err
    }
  }
}

// Find a good way for main thread to communicate when routine should close
// Possibly have a channel specificaly for this task.  
func reader(irc *IRC) {
  br := bufio.NewReader(irc.conn)
  for {
    msg, err := br.ReadString('\n') 
    if err != nil {
      irc.errlog.Printf("Error while reading: %s", err)
      irc.err <- err
    }
    if msg != "" {
      irc.read <- msg
    }
    // Check if it is time to exit
    select {
    case <-irc.readsync:
      return
    }
  }
}

func (irc *IRC) Connect(nick, server string) error {

  irc.server = server
  irc.nick = nick

  // Start stdlog
  if irc.stdlog == nil {
    irc.stdlog = log.New(os.Stdout, server + ": ", 0)
  }

  // Start errlog
  if irc.errlog == nil {
    irc.errlog = log.New(os.Stderr, server + ": ", 0)
  }


  irc.stdlog.Printf("Attempting to connect to: %s\n", irc.server)
  var err error
  if irc.conn, err = net.Dial("tcp", irc.server); err != nil {
    irc.errlog.Printf("Failed to connect to %s: %s\n", irc.server, err)
    return err
  }
  irc.stdlog.Printf("Connected to %s\n", irc.server)

  irc.read = make(chan string, 64)
  irc.write = make(chan string, 64)
  irc.err = make(chan error)
  irc.readsync = make(chan bool)
  irc.writesync = make(chan bool)

  go writer(irc)
  go reader(irc)
  irc.SendRaw(fmt.Sprintf("NICK %s", irc.nick))
  irc.SendRaw(fmt.Sprintf("USER %s 0.0.0.0 0.0.0.0 :%s", irc.nick, irc.nick))
  return nil
}

func (irc *IRC) ReConnect() error {
  irc.stdlog.Println("Reconnecting")
  // Close read/write channels
  close(irc.read)
  close(irc.write)
  // Let last read/write finish
  irc.readsync <- true
  <- irc.writesync
  // Tell server we are leaving
  irc.Quit()
  return irc.Connect(irc.nick, irc.server)
}

// Send text directly to server.
func (irc *IRC) SendRaw(output string) {
  irc.write <- fmt.Sprintf("%s\r\n", output)
}

func (irc *IRC) Quit() {
  irc.SendRaw("QUIT")
}

func (irc *IRC) Join(channel string) {
  irc.SendRaw(fmt.Sprintf("JOIN %s", channel))
}

func (irc *IRC) Part(channel string) {
  irc.SendRaw(fmt.Sprintf("PART %s", channel))
}

func (irc *IRC) Privmsg(target, message string) {
  irc.SendRaw(fmt.Sprintf("PRIVMSG %s :%s", target, message))
}

func (irc *IRC) Notice(target, message string) {
  irc.SendRaw(fmt.Sprintf("NOTICE %s :%s", target, message))
}
