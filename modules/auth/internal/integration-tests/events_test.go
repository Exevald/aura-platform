package integration_tests

import (
	"testing"
	"time"

	rabbit "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"
)

func subscribeDomainEvents(t *testing.T) <-chan rabbit.Delivery {
	connection, err := rabbit.Dial(testEnv.conf.AMQP.URL)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, connection.Close()) })
	channel, err := connection.Channel()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, channel.Close()) })
	queue, err := channel.QueueDeclare("", false, true, true, false, nil)
	require.NoError(t, err)
	require.NoError(t, channel.QueueBind(queue.Name, "#", "domain-event", false, nil))
	deliveries, err := channel.Consume(queue.Name, "", true, true, false, false, nil)
	require.NoError(t, err)
	return deliveries
}

func requireEventTypes(t *testing.T, deliveries <-chan rabbit.Delivery, expected ...string) {
	want := make(map[string]struct{}, len(expected))
	for _, eventType := range expected {
		want[eventType] = struct{}{}
	}
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	for len(want) > 0 {
		select {
		case delivery := <-deliveries:
			delete(want, delivery.Type)
		case <-timer.C:
			require.Empty(t, want, "expected domain events were not published")
		}
	}
}
