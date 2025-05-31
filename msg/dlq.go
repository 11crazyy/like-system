package msg

import "github.com/apache/pulsar-client-go/pulsar"

func CreateConsumerWithDLQ() (consumer pulsar.Consumer, err error) {
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: "pulsar://localhost:6650",
	})
	if err != nil {
		return nil, err
	}

	consumer, err = client.Subscribe(pulsar.ConsumerOptions{
		Topic:            "thumb-topic",
		SubscriptionName: "thumb-subscription",
		Type:             pulsar.Shared,
		DLQ: &pulsar.DLQPolicy{
			DeadLetterTopic:  "thumb-dlq-topic",
			RetryLetterTopic: "thumb-retry-topic",
			MaxDeliveries:    3,
		},
	})
	return consumer, err
}
