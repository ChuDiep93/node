package promise

import (
	"github.com/mysteriumnetwork/node/communication"
)

// PromiseRequest structure represents message from service consumer to send a promise
type PromiseRequest struct {
	PromiseMessage PromiseMessage `json:"promiseMessage"`
}

type PromiseMessage struct {
	Amount     uint64 `json:"amount"`
	SequenceID uint64 `json:"sequenceID"`
	Signature  string `json:"signature"`
}

const endpointPromise = "session-promise"
const messageEndpointPromise = communication.MessageEndpoint(endpointPromise)

type PromiseSender struct {
	sender communication.Sender
}

func NewPromiseSender(sender communication.Sender) *PromiseSender {
	return &PromiseSender{
		sender: sender,
	}
}

func (ps *PromiseSender) Send(pm PromiseMessage) error {
	err := ps.sender.Send(&promiseMessageProducer{PromiseMessage: pm})
	return err
}

type PromiseListener struct {
	promiseMessageConsumer *promiseMessageConsumer
}

func NewPromiseListener(promiseChan chan PromiseMessage) *PromiseListener {
	return &PromiseListener{
		promiseMessageConsumer: &promiseMessageConsumer{
			queue: promiseChan,
		},
	}
}

func (pl *PromiseListener) GetConsumer() *promiseMessageConsumer {
	return pl.promiseMessageConsumer
}

// Consume handles requests from endpoint and replies with response
func (pmc *promiseMessageConsumer) Consume(requestPtr interface{}) (err error) {
	request := requestPtr.(*PromiseRequest)
	pmc.queue <- request.PromiseMessage
	return nil
}

// Dialog boilerplate below, please ignore

type promiseMessageConsumer struct {
	queue chan PromiseMessage
}

// GetMessageEndpoint returns endpoint where to receive messages
func (pmc *promiseMessageConsumer) GetMessageEndpoint() communication.MessageEndpoint {
	return messageEndpointPromise
}

// NewRequest creates struct where request from endpoint will be serialized
func (pmc *promiseMessageConsumer) NewMessage() (requestPtr interface{}) {
	return &PromiseRequest{}
}

// promiseMessageProducer
type promiseMessageProducer struct {
	PromiseMessage PromiseMessage
}

// GetMessageEndpoint returns endpoint where to receive messages
func (pmp *promiseMessageProducer) GetMessageEndpoint() communication.MessageEndpoint {
	return messageEndpointPromise
}

func (pmp *promiseMessageProducer) Produce() (requestPtr interface{}) {
	return &PromiseRequest{
		PromiseMessage: pmp.PromiseMessage,
	}
}

func (pmp *promiseMessageProducer) NewResponse() (responsePtr interface{}) {
	return nil
}
