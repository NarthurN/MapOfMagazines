package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var citiesOfAllWorld []string = []string{"Moscow", "Saint", "Goland"}

type Magazine struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	City string `json:"city"`
}

type MapOfMagazines struct {
	client *http.Client
}

func newMapOfMagazines() *MapOfMagazines {
	return &MapOfMagazines{client: &http.Client{Timeout: 1 * time.Second}}
}

func (m *MapOfMagazines) findMagazinesInCity(city string) ([]Magazine, error) {
	baseURL := "http://localhost:8080/getMagazinesByCity/"
	requestedURL := baseURL + city
	fmt.Println("requestedURL", requestedURL)

	req, err := http.NewRequest(http.MethodGet, requestedURL, nil)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	response, err := m.client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var magazine []Magazine
	if err = json.NewDecoder(response.Body).Decode(&magazine); err != nil {
		fmt.Println(err)
		return nil, err
	}

	return magazine, nil
}

func main() {
	ch := connectRabbitMQ()
	defer ch.Close()

	mag := newMapOfMagazines()
	consumeOrders(ch)

	// создаём канал генерации для городов
	citiesChanel := make(chan string)
	// В этот канал записываем города из citiesOfAllWorld в течении 5 секунд
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	go generateCities(ctx, citiesChanel)

	for city := range citiesChanel {
		storiesOfMoscow, err := mag.findMagazinesInCity(city)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(storiesOfMoscow)
	}
}

func generateCities(ctx context.Context, citiesChanel chan<- string) {
	source := rand.NewSource(time.Now().Unix())
	r := rand.New(source)
	for {
		select {
		case <-ctx.Done():
			close(citiesChanel)
			return
		case <- time.Tick(1 * time.Second):
			index := r.Int() % len(citiesOfAllWorld)
			citiesChanel <- citiesOfAllWorld[index]
		}
	}
}

func connectRabbitMQ() *amqp.Channel {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}

	return ch
}

func consumeOrders(ch *amqp.Channel) {
	q, err := ch.QueueDeclare(
		"magazinesNotification", // name
		false,                   // durable
		false,                   // delete when unused
		false,                   // exclusive
		false,                   // no-wait
		nil,                     // arguments
	)
	if err != nil {
		panic(err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args,
	)
	if err != nil {
		panic(err)
	}

	go func() {
		for d := range msgs {
			fmt.Println("СООБЩЕНИЕ ПОЛУЧЕНО:", string(d.Body))
			fmt.Println()
		}
	}()
}
