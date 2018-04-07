package engine

import (
	"errors"
	"github.com/olebedev/go-duktape"
	"github.com/satori/go.uuid"
	"log"
)

type Engine struct {
	Interp     *duktape.Context
	CanExecute bool
}

func Create() (*Engine, error) {
	this := new(Engine)
	ctx := duktape.New()
	this.Interp = ctx

	ctx.PushGlobalStash()
	ctx.PushObject()
	ctx.PutPropString(-2, "remoteCalls")

	ctx.PushGlobalGoFunction("log", func(c *duktape.Context) int {
		// log(message)
		log.Println(c.SafeToString(-1))
		return 0
	})
// Need getValue, setValue, and setResponse functions
	ctx.PushGlobalGoFunction("setHandler", func(c *duktape.Context) int {
		// setHandler(handler(type, data))
		c.RequireCallable(0)
		c.PushGlobalStash()
		c.Dup(0)
		c.PutPropString(-2, "handler")
		this.CanExecute = true
		log.Println("set handler")
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

func (this *Engine) LoadScript(script string) error {
	ctx := this.Interp
	ctx.PevalString(script)
	return nil
}

func (this *Engine) Execute(name string, context map[string]string) (string, error) {
	if !this.CanExecute {
		return "", errors.New("Engine not executable")
	}
	log.Println("executing")
	ctx := this.Interp
	ctx.PushGlobalStash()
	ctx.GetPropString(-1, "handler")
	ctx.PushString(name)

	ctx.PushObject()
	for key, value := range context {
		ctx.PushString(value)
		ctx.PutPropString(-2, key)
	}
	ctx.Pcall(2)
	ctx.JsonEncode(-1)
	result := ctx.GetString(-1)
	ctx.Pop()
	return result, nil
}

func (this *Engine) Cleanup() error {
	defer this.Interp.DestroyHeap()
	log.Println("Cleanup")
	return nil
}
