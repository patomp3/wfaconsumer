package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/streadway/amqp"
)

// ReceiveQueue struct...
type ReceiveQueue struct {
	URL       string
	QueueName string
}

// UpdateRequest for ...
type UpdateRequest struct {
	OrderTransID    string `json:"order_trans_id"`
	OrderID         string `json:"order_id"`
	Status          string `json:"status"`
	ErrorCode       string `json:"error_code"`
	ErrorDesc       string `json:"error_desc"`
	ResponseMessage string `json:"response_message"`
}

// UpdateResponse for ..
type UpdateResponse struct {
	OrderTransID     string `json:"order_trans_id"`
	ErrorCode        string `json:"error_code"`
	ErrorDescription string `json:"error_description"`
}

func failOnError(err error, msg string) {
	if err != nil {
		fmt.Printf("%s: %s", msg, err)
	}
}

// Close for
func (r ReceiveQueue) Close() {
	//q.conn.Close()
	//q.ch.Close()
}

// Connect for
func (r ReceiveQueue) Connect() *amqp.Channel {
	conn, err := amqp.Dial(r.URL)
	//defer conn.Close()
	if err != nil {
		failOnError(err, "Failed to connect to RabbitMQ")
		return nil
	}

	ch, err := conn.Channel()
	//defer ch.Close()
	if err != nil {
		failOnError(err, "Failed to open a channel")
		return nil
	}

	return ch
}

// Receive for receive message from queue
func (r ReceiveQueue) Receive(ch *amqp.Channel) {

	/*conn, err := amqp.Dial(q.URL)
	defer conn.Close()
	if err != nil {
		failOnError(err, "Failed to connect to RabbitMQ")
		return false
	}

	ch, err := conn.Channel()
	defer ch.Close()
	if err != nil {
		failOnError(err, "Failed to open a channel")
		return false
	}*/

	q, err := ch.QueueDeclarePassive(
		r.QueueName, // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		failOnError(err, "Failed to declare a queue")
	}

	msgs, err := ch.Consume(
		q.Name, // routing key
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		failOnError(err, "Failed to publish a message")

	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			orderTransID := string(d.Body)
			orderID := d.MessageId

			log.Printf("## Received Order Trans Id : %s", orderTransID)
			log.Printf("## >> Id : %s", orderID)

			//TODO : Process for consumer to

			//TODO : Notify result to wfa core
			var res UpdateRequest
			res.OrderTransID = orderTransID
			res.OrderID = orderID
			res.Status = "Z"
			res.ErrorCode = "0"
			res.ErrorDesc = ""
			notifyResult(res)
		}
	}()

	log.Printf("## [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func notifyResult(req UpdateRequest) UpdateResponse {
	var resultRes UpdateResponse

	reqPost, _ := json.Marshal(req)

	response, err := http.Post(cfg.updateOrderURL, "application/json", bytes.NewBuffer(reqPost))
	if err != nil {
		log.Printf("The HTTP request failed with error %s", err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		//fmt.Println(string(data))
		//myReturn = json.Unmarshal(string(data))
		err = json.Unmarshal(data, &resultRes)
		if err != nil {
			//panic(err)
			log.Printf("The HTTP response failed with error %s", err)
		} else {
			log.Printf("## Result >> %v", resultRes)
		}
	}

	return resultRes
}
