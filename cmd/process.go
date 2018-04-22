// Copyright © 2018 Sebastian Green-Husted <geoffcake@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/ricecake/funky/datastore"
	"github.com/ricecake/funky/engine"
	"github.com/ricecake/rascal"
	"github.com/spf13/cobra"
	"github.com/streadway/amqp"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// processCmd represents the process command
var processCmd = &cobra.Command{
	Use:   "process",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("process called")
		amqpHandler := new(rascal.Rascal)

		err := amqpHandler.Connect()
		if err != nil {
			panic(err)
		}
		defer amqpHandler.Cleanup()

		var mapLock sync.Mutex
		reqTable := make(map[string]chan amqp.Delivery)
		amqpHandler.SetHandler(amqpHandler.Default, func(msg amqp.Delivery, ch *amqp.Channel) {
			rawUUID := msg.CorrelationId
			mapLock.Lock()
			channel, exists := reqTable[rawUUID]
			if exists {
				delete(reqTable, rawUUID)
				mapLock.Unlock()
				channel <- msg
				msg.Ack(false)
			} else {
				mapLock.Unlock()
				nackErr := msg.Nack(false, false)
				if nackErr != nil {
					log.Printf("failed to Nack() (no requeue): %s", nackErr)
				}
			}
		})

		amqpHandler.SetHandler(amqpHandler.Custom, func(msg amqp.Delivery, ch *amqp.Channel) {
			log.Printf("%s\n", msg.Body)

			var data payload
			decodeErr := json.Unmarshal(msg.Body, &data)
			if decodeErr != nil {
				log.Println(decodeErr.Error())
				return
			}

			handler, lookupErr := datastore.LookupEventRoute(data.Owner, data.Scope, data.Route)
			if lookupErr != nil {
				log.Println(lookupErr.Error())
				return
			}
			handler.Load()
			log.Printf("Route: %+v\n", handler)

			inputChan := make(chan engine.Message)
			outputChan := make(chan engine.Message)

			defer close(inputChan)

			replyChan := make(chan amqp.Delivery)
			eng, _ := engine.Create(handler.Scope, inputChan, outputChan)
			defer eng.Cleanup()

			script, readErr := handler.ReadScript()
			if readErr != nil {
				log.Println(readErr.Error())
				return
			}

			go eng.Run()
			inputChan <- engine.Message{
				Type: engine.LoadScript,
				Data: script,
			}
			inputChan <- engine.Message{
				Type: engine.Request,
				Data: data,
			}

			select {
			case result := <-outputChan:
				if msg.ReplyTo != "" && msg.CorrelationId != "" {
					pubErr := ch.Publish(
						"",          // exchange
						msg.ReplyTo, // routing key
						false,       // mandatory
						false,       // immediate
						amqp.Publishing{
							ContentType:   "text/plain",
							CorrelationId: msg.CorrelationId,
							Body:          []byte(result.Data.(string)),
						})
					if pubErr != nil {
						log.Println("reply error: %s", pubErr.Error())
						msg.Nack(false, false)
					}
				}
				msg.Ack(false)
			case request := <-inputChan:
				log.Println("request!", request)
			case reply := <-replyChan:
				log.Println("reply!", reply)
			case <-time.After(1 * time.Second):
				msg.Nack(false, false)
				log.Println("Timeout!")
			}
		})

		amqpHandler.Consume()

		osChannel := make(chan os.Signal, 2)
		signal.Notify(osChannel, os.Interrupt, syscall.SIGTERM)
		sig := <-osChannel
		log.Printf("Received [%+v]; shutting down...", sig)
		os.Exit(0)

	},
}

func init() {
	rootCmd.AddCommand(processCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// processCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// processCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
