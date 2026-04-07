package meilisearch

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/fishkaoff/ts-backend/internal/config"
	"github.com/fishkaoff/ts-backend/internal/domain/types"
	"github.com/meilisearch/meilisearch-go"
)

type Product struct {
	Code         int               `json:"code"`
	PartNumber   string            `json:"part_number"`
	Name         string            `json:"name"`
	Manufacturer string            `json:"manufacturer"`
	Unit         types.ProductUnit `json:"unit"`
	Price        int64             `json:"price"`
	Balance      int64             `json:"balance"`
	IsNew        bool              `json:"is_new"`
	ImageURL     string            `json:"image_url"`
	Active       bool              `json:"active"`
}

func (p *Product) FromGlobalProduct(gProduct types.Product) {
	p.Code = gProduct.Code
	p.PartNumber = gProduct.PartNumber
	p.Name = gProduct.Name
	p.Manufacturer = gProduct.Manufacturer
	p.Unit = gProduct.Unit
	p.Price = gProduct.Price
	p.Balance = gProduct.Balance
	p.IsNew = gProduct.IsNew
	p.ImageURL = gProduct.ImageURL
	p.Active = gProduct.Active
}

type Engine struct {
	cfg    config.MeilisearchConfig
	client meilisearch.ServiceManager
}

func New(cfg config.MeilisearchConfig) *Engine {
	client := meilisearch.New(
		cfg.URL,
		meilisearch.WithAPIKey(cfg.ApiKey),
	)

	return &Engine{
		cfg:    cfg,
		client: client,
	}
}

func (e *Engine) UpsertProducts(products []types.Product) error {
	const op = "meilisearch.UpsertProducts"

	if e.cfg.Init {
		err := e.createIndex()
		if err != nil {
			return fmt.Errorf("%s:%w", op, err)
		}

		attrs := []string{
			"name",
			"part_number",
		}

		err = e.updateSearchable(attrs)
		if err != nil {
			return fmt.Errorf("%s:%w", op, err)
		}
	}

	primaryKey := "code"
	opts := meilisearch.DocumentOptions{
		PrimaryKey: &primaryKey,
	}

	parsedProducts := e.mapProducts(products)

	task, err := e.client.Index("products").AddDocuments(parsedProducts, &opts)
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}

	_, err = e.client.WaitForTask(task.TaskUID, 30*time.Minute)
	if err != nil {
		return fmt.Errorf("%s: wait task: %w", op, err)
	}

	return nil
}

func (e *Engine) SearchProducts(query string, page int, limit int) ([]Product, error) {
	const op = "meilisearch.SearchProducts"

	offset := (page - 1) * limit

	res, err := e.client.Index("products").Search(query, &meilisearch.SearchRequest{
		Limit:  int64(limit),
		Offset: int64(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	var products []Product
	data, _ := json.Marshal(res.Hits)
	if err := json.Unmarshal(data, &products); err != nil {
		return nil, err
	}

	return products, nil
}

func (e *Engine) createIndex() error {
	const op = "meilisearch.createIndex"
	task, err := e.client.CreateIndex(&meilisearch.IndexConfig{
		Uid: "products",
	})
	if err != nil {
		return fmt.Errorf("%s:failed to create index: %w", op, err)
	}

	_, err = e.client.WaitForTask(task.TaskUID, 30*time.Minute)
	if err != nil {
		return fmt.Errorf("%s: wait task: %w", op, err)
	}
	return nil
}

func (e *Engine) updateSearchable(attrs []string) error {
	const op = "meilisearch.updateSearchable"
	task, err := e.client.Index("products").UpdateSearchableAttributes(&attrs)
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}

	_, err = e.client.WaitForTask(task.TaskUID, 30*time.Minute)
	if err != nil {
		return fmt.Errorf("%s: wait task: %w", op, err)
	}
	return nil
}

func (e *Engine) mapProducts(products []types.Product) []Product {
	parsedProducts := make([]Product, 0, len(products))

	for _, gProduct := range products {
		var parsedProduct Product
		parsedProduct.FromGlobalProduct(gProduct)
		parsedProducts = append(parsedProducts, parsedProduct)
	}

	return parsedProducts
}
