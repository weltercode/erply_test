package api

import (
	"context"

	"github.com/erply/api-go-wrapper/pkg/api/customers"
)

type CustomerManagerInterface interface {
	GetCustomersBulk(ctx context.Context, filters []map[string]interface{}, opts map[string]string) (customers.GetCustomersResponseBulk, error)
	DeleteCustomerBulk(ctx context.Context, bulk []map[string]interface{}, opts map[string]string) (customers.DeleteCustomersResponseBulk, error)
	SaveCustomerBulk(ctx context.Context, bulk []map[string]interface{}, opts map[string]string) (customers.SaveCustomerResponseBulk, error)
}
