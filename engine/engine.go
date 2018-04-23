package engine

import (
	"errors"
	//        "github.com/ricecake/funky/datastore"
	"github.com/olebedev/go-duktape"
	"github.com/satori/go.uuid"
	"log"
)

type MessageType int

const (
	Log MessageType = iota
	Request
	Event
	Reply
	NoReply
	Require
	DataStore
	DataLoad
	LoadScript
)

type Message struct {
	Id    string
	Type  MessageType
	Route string
	Data  interface{}
}

type Engine struct {
	Interp     *duktape.Context
	CanExecute bool
	context    string
	Input      chan Message
	Output     chan Message
}

func Create(context string, input chan Message, output chan Message) (*Engine, error) {
	this := new(Engine)
	ctx := duktape.New()
	this.Interp = ctx
	this.context = context

	this.Input = input
	this.Output = output

	ctx.PushGlobalStash()
	ctx.PushObject()
	ctx.PutPropString(-2, "remoteCalls")

	ctx.PushGlobalGoFunction("log", func(c *duktape.Context) int {
		// log(level, message)
		this.Output <- Message{
			Type: Log,
			Data: c.SafeToString(-1),
		}
		//		log.Println(c.SafeToString(-2) + ": " + c.SafeToString(-1))
		return 0
	})
	// Need getValue, setValue, loadScript and setResponse functions.  Add channel and context to struct.
	ctx.PushGlobalGoFunction("setHandler", func(c *duktape.Context) int {
		// setHandler(handler(type, data))
		c.RequireCallable(0)
		c.PushGlobalStash()
		c.Dup(0)
		c.PutPropString(-2, "handler")
		this.CanExecute = true
		return 0
	})

	ctx.PushGlobalGoFunction("getValue", func(c *duktape.Context) int {
		// getValue(key)
		return 0
	})
	ctx.PushGlobalGoFunction("setValue", func(c *duktape.Context) int {
		// setValue(key, value)
		return 0
	})
	ctx.PushGlobalGoFunction("loadScript", func(c *duktape.Context) int {
		// loadScript(name)
		return 0
	})
	ctx.PushGlobalGoFunction("setResponse", func(c *duktape.Context) int {
		// setResponse(response)
		if c.IsNullOrUndefined(0) {
			this.Output <- Message{Type: NoReply}
		} else {
			c.JsonEncode(0)
			response := c.GetString(0)
			this.Output <- Message{
				Type: Reply,
				Data: response,
			}
		}
		return 0
	})

	ctx.PushGlobalGoFunction("emitMessage", func(c *duktape.Context) int {
		// emitMessage(type, data)
		return 0
	})
	ctx.PushGlobalGoFunction("callRemote", func(c *duktape.Context) int {
		// callRemote(type, data, handler(type, data))
		UUID, err := uuid.NewV4()
		if err != nil {
			log.Printf("Something went wrong: %s", err)
			return 1
		}
		c.RequireString(0)
		c.RequireCallable(2)
		c.PushGlobalStash()
		c.GetPropString(-1, "remoteCalls")
		c.Dup(2)
		c.PutPropString(-2, UUID.String())
		c.JsonEncode(1)
		log.Printf("%s: %s\n", c.GetString(0), c.GetString(1))
		return 0
	})

	return this, nil
}

func (this *Engine) Run() error {
	defer close(this.Output)
	defer this.Cleanup()
	for msg := range this.Input {
		switch msg.Type {
		case LoadScript:
			this.LoadScript(msg.Data.(string))
		case Request:
			if !this.CanExecute {
				return errors.New("Engine not executable")
			}
			ctx := this.Interp
			ctx.PushGlobalStash()
			ctx.GetPropString(-1, "handler")
			ctx.PushString(msg.Route)

			ctx.PushObject()
			ctx.Pcall(2)
		}
	}
	return nil
}

func (this *Engine) LoadScript(script string) error {
	ctx := this.Interp
	ctx.PevalString(script)
	return nil
}

func (this *Engine) Cleanup() error {
	defer this.Interp.DestroyHeap()
	return nil
}
