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
	"github.com/gin-gonic/gin"
	"github.com/ricecake/funky/datastore"
	"github.com/ricecake/rascal"
	"github.com/satori/go.uuid"
	"github.com/spf13/cobra"
	"github.com/streadway/amqp"
	"log"
	"sync"
	"time"
)

// listenCmd represents the listen command
var listenCmd = &cobra.Command{
	Use:   "listen",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("listen called")

		amqpHandler := new(rascal.Rascal)

		amqpHandler.Connect()
		defer amqpHandler.Cleanup()

		var mapLock sync.Mutex
		reqTable := make(map[uuid.UUID]chan amqp.Delivery)
		amqpHandler.SetHandler(amqpHandler.Default, func(msg amqp.Delivery, ch *amqp.Channel) {
			rawUUID, uuidErr := uuid.FromString(msg.CorrelationId)
			if uuidErr == nil {
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
			} else {
				log.Printf("Malformed UUID: %s", msg.CorrelationId)
				nackErr := msg.Nack(false, false)
				if nackErr != nil {
					log.Printf("failed to Nack() (no requeue): %s", nackErr)
				}
			}
		})

		amqpHandler.Consume()

		r := gin.Default()
		r.NoRoute(func(c *gin.Context) {
			corrID, uuidErr := uuid.NewV4()
			if uuidErr != nil {
				c.String(500, uuidErr.Error())
				return
			}
			reqChan := make(chan amqp.Delivery)

			rawBody, bodyErr := c.GetRawData()
			if bodyErr != nil {
				c.String(500, bodyErr.Error())
				return
			}

			outbound := payload{
				Method:  c.Request.Method,
				Url:     c.Request.URL.String(),
				Host:    c.Request.Host,
				Body:    rawBody,
				Headers: c.Request.Header,
			}
			routes, owner, scope, routeErr := datastore.LookupDomainRoute(outbound.Host, outbound.Url)
			if routeErr != nil {
				c.String(500, routeErr.Error())
				return
			}

			if len(routes) == 0 {
				c.String(404, "No Such Route")
				return
			}

			outbound.Owner = owner
			outbound.Scope = scope

			ch, chErr := amqpHandler.Channel()
			if chErr != nil {
				c.String(500, chErr.Error())
				return
			}
			defer ch.Close()

			mapLock.Lock()
			reqTable[corrID] = reqChan
			mapLock.Unlock()

			for _, route := range routes {
				outbound.Route = route
				encodedOutbound, encodeErr := json.Marshal(outbound)
				if encodeErr != nil {
					c.String(500, encodeErr.Error())
					return
				}
				log.Printf("SENDING %s\n", encodedOutbound)
				pubErr := ch.Publish(
					"request",         // exchange
					"execute.initial", // routing key
					false,             // mandatory
					false,             // immediate
					amqp.Publishing{
						ContentType:   "text/plain",
						CorrelationId: corrID.String(),
						ReplyTo:       amqpHandler.Default,
						Body:          encodedOutbound,
						Expiration:    "1000",
					})
				if pubErr != nil {
					c.String(500, pubErr.Error())
					return
				}
			}

			select {
			case reply := <-reqChan:
				c.JSON(200, gin.H{
					"message": string(reply.Body),
				})
			case <-time.After(1 * time.Second):
				mapLock.Lock()
				delete(reqTable, corrID)
				mapLock.Unlock()
				c.String(500, "TIMEOUT")
			}

		})
		r.Run() // listen and serve on 0.0.0.0:8080

	},
}

func init() {
	rootCmd.AddCommand(listenCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listenCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listenCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

type payload struct {
	Route   string
	Owner   string
	Scope   string
	Method  string
	Url     string
	Host    string
	Headers map[string][]string
	Body    []byte
}
