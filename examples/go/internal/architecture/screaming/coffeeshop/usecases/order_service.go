package usecases

import (
	"log"

	"github.com/qjcg/arcadia/examples/go/internal/architecture/screaming/coffeeshop/entities"
	"github.com/qjcg/arcadia/examples/go/internal/architecture/screaming/coffeeshop/infra"
)

type OrderService struct {
	Store infra.OrderStore
}

func NewOrderService(store infra.OrderStore) *OrderService {
	return &OrderService{Store: store}
}

func (s *OrderService) AddItem(items ...entities.Item) {
	order := entities.Order{
		Items: items,
	}

	err := s.Store.Save(order)
	if err != nil {
		log.Fatalf("error saving order: %v", err)
	}
}

func (s *OrderService) Total() float64 {
	orders, err := s.Store.FindAll()
	if err != nil {
		log.Fatal(err)
	}

	var total float64
	for i := range orders {
		items := orders[i].Items

		for j := range items {
			total += items[j].Price
		}
	}

	return total
}
