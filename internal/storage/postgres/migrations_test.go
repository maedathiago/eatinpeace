package postgres_test

import (
	"strings"
	"testing"

	"github.com/maedathiago/eatinpeace/internal/storage/postgres"
)

func TestOperationalFoundationMigrationCoversMinimumEntities(t *testing.T) {
	sql, err := postgres.ReadOperationalFoundationMigration("../../..")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	for _, table := range []string{
		"restaurants",
		"service_shifts",
		"tables",
		"table_sessions",
		"orders",
		"order_items",
		"operational_events",
		"floor_tasks",
		"complaints",
		"bill_handoffs",
		"sla_policies",
		"staff_members",
	} {
		if !strings.Contains(sql, "create table if not exists "+table) {
			t.Fatalf("migration does not create %s", table)
		}
	}
	for _, field := range []string{"card_number", "payment_intent", "tax_amount", "invoice_number", "stock_quantity"} {
		if strings.Contains(sql, field) {
			t.Fatalf("migration introduced forbidden field %s", field)
		}
	}
}
