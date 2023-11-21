package collector

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/alphagov/paas-rds-metric-collector/pkg/metrics"
)

// openMultipleDBConns opens as many connections as specified by
// count using the given driver and url.
func openMultipleDBConns(ctx context.Context, count int, driver, url string) (err error, execQuery func(string)) {
	var dbConns []*sql.DB

	go func() {
		select {
		case <-ctx.Done():
			for _, c := range dbConns {
				c.Close()
			}
		}
	}()

	execQuery = func(q string) {
		for _, c := range dbConns {
			go c.ExecContext(ctx, q)
		}
	}

	for i := 0; i < count; i++ {
		dbConn, err := sql.Open(driver, url)
		if err != nil {
			break
		}
		err = dbConn.Ping()
		if err != nil {
			break
		}
		dbConns = append(dbConns, dbConn)
	}
	return err, execQuery
}

func getMetricByKey(collectedMetrics []metrics.Metric, key string) *metrics.Metric {
	for _, metric := range collectedMetrics {
		if metric.Key == key {
			return &metric
		}
	}
	return nil
}

// Replaces the DB name in a postgres DB connection string
func injectDBName(connectionString, newDBName string) string {
	re := regexp.MustCompile("(.*:[0-9()]+)[^?]*([?].*)?$")
	return re.ReplaceAllString(connectionString, fmt.Sprintf("$1/%s$2", newDBName))
}

var _ = Describe("injectDBName", func() {
	It("replaces the db name", func() {
		Expect(
			injectDBName("postgresql://postgres@localhost:5432/foo?sslmode=disable", "mydb"),
		).To(Equal(
			"postgresql://postgres@localhost:5432/mydb?sslmode=disable",
		))
		Expect(
			injectDBName("postgresql://postgres@localhost:5432?sslmode=disable", "mydb"),
		).To(Equal(
			"postgresql://postgres@localhost:5432/mydb?sslmode=disable",
		))
		Expect(
			injectDBName("user:pass@tcp(localhost:3306)?something=false", "mydb"),
		).To(Equal(
			"user:pass@tcp(localhost:3306)/mydb?something=false",
		))
		Expect(
			injectDBName("user:pass@tcp(localhost:3306)/foo?something=false", "mydb"),
		).To(Equal(
			"user:pass@tcp(localhost:3306)/mydb?something=false",
		))
	})
})
