package collector

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/mock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/alphagov/paas-rds-metric-collector/pkg/brokerinfo"
	"github.com/alphagov/paas-rds-metric-collector/pkg/brokerinfo/fakebrokerinfo"
	"github.com/alphagov/paas-rds-metric-collector/pkg/metrics"
	"github.com/alphagov/paas-rds-metric-collector/pkg/utils"
)

var _ = Describe("NewMysqlMetricsCollectorDriver", func() {

	var (
		brokerInfo             *fakebrokerinfo.FakeBrokerInfo
		metricsCollectorDriver MetricsCollectorDriver
		metricsCollector       MetricsCollector
		collectedMetrics       []metrics.Metric
		testDBName             string
		testDBConnectionString string
	)

	BeforeEach(func() {
		testDBName = utils.RandomString(10)
		testDBConnectionString = injectDBName(mysqlTestDatabaseConnectionURL, testDBName)
		mainDBConn, err := sql.Open("mysql", mysqlTestDatabaseConnectionURL)
		Expect(err).NotTo(HaveOccurred())
		defer mainDBConn.Close()
		_, err = mainDBConn.Exec(fmt.Sprintf("CREATE DATABASE %s", testDBName))
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		dbConn, err := sql.Open("mysql", mysqlTestDatabaseConnectionURL)
		Expect(err).NotTo(HaveOccurred())
		defer dbConn.Close()

		_, err = dbConn.Query(fmt.Sprintf("DROP DATABASE %s", testDBName))
		Expect(err).NotTo(HaveOccurred())
	})

	BeforeEach(func() {
		var err error
		brokerInfo = &fakebrokerinfo.FakeBrokerInfo{}
		metricsCollectorDriver = NewMysqlMetricsCollectorDriver(5, brokerInfo, logger)

		brokerInfo.On(
			"ConnectionString", mock.Anything,
		).Return(
			testDBConnectionString, nil,
		)
		By("Creating a new collector")
		metricsCollector, err = metricsCollectorDriver.NewCollector(
			brokerinfo.InstanceInfo{
				GUID: "instance-guid1",
				Type: "mysql",
			},
		)
		Expect(err).NotTo(HaveOccurred())

		By("Retrieving initial metrics")
		collectedMetrics, err = metricsCollector.Collect()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := metricsCollector.Close()
		Expect(err).NotTo(HaveOccurred())
	})

	It("should return the CollectInterval", func() {
		Expect(metricsCollectorDriver.GetCollectInterval()).To(Equal(5))
	})

	It("returns the right metricsCollectorDriver name", func() {
		Expect(metricsCollectorDriver.GetName()).To(Equal("mysql"))
	})

	It("can collect the number of connection_errors", func() {
		metric := getMetricByKey(collectedMetrics, "connection_errors")
		Expect(metric).ToNot(BeNil())
		Expect(metric.Value).To(BeNumerically(">=", 0))
		Expect(metric.Unit).To(Equal("err"))
	})

	It("can collect connection-related metrics", func() {
		metrics := []string{
			"threads_running",
			"threads_connected",
			"threads_created",
			"max_connections",
			"queries",
			"questions",
			"aborted_connects",
			"aborted_clients",
		}
		for _, v := range metrics {
			By(fmt.Sprintf("Checking %s", v))
			metric := getMetricByKey(collectedMetrics, v)
			Expect(metric).ToNot(BeNil())
			Expect(metric.Value).To(BeNumerically(">=", 0))
			Expect(metric.Unit).To(Equal("conn"))
		}
	})

	It("can collect InnoDB metrics", func() {
		metrics := []string{
			"innodb_row_lock_time",
			"innodb_row_lock_waits",
			"innodb_num_open_files",
			"innodb_log_waits",
			"innodb_buffer_pool_bytes_data",
			"innodb_buffer_pool_bytes_dirty",
			"innodb_buffer_pool_pages_data",
			"innodb_buffer_pool_pages_dirty",
			"innodb_buffer_pool_pages_flushed",
			"innodb_buffer_pool_pages_free",
			"innodb_buffer_pool_pages_misc",
			"innodb_buffer_pool_pages_total",
			"innodb_buffer_pool_read_ahead",
			"innodb_buffer_pool_read_ahead_evicted",
			"innodb_buffer_pool_read_ahead_rnd",
			"innodb_buffer_pool_read_requests",
			"innodb_buffer_pool_reads",
			"innodb_buffer_pool_wait_free",
			"innodb_buffer_pool_write_requests",
		}
		for _, v := range metrics {
			By(fmt.Sprintf("Checking %s", v))
			metric := getMetricByKey(collectedMetrics, v)
			Expect(metric).ToNot(BeNil())
			Expect(metric.Value).To(BeNumerically(">=", 0))
		}
	})

	It("can collect the number of threads connected and threads running", func() {
		var err error

		By("Checking initial number of threads connected")
		metric := getMetricByKey(collectedMetrics, "threads_connected")
		Expect(metric).ToNot(BeNil())
		Expect(metric.Value).To(BeNumerically(">=", 1))
		Expect(metric.Unit).To(Equal("conn"))

		initialConnections := metric.Value

		By("Creating multiple new threads_connected")
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		err, execQueryFunc := openMultipleDBConns(ctx, 20, "mysql", mysqlTestDatabaseConnectionURL)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() float64 {
			collectedMetrics, err = metricsCollector.Collect()
			Expect(err).NotTo(HaveOccurred())
			metric = getMetricByKey(collectedMetrics, "threads_connected")
			Expect(metric).ToNot(BeNil())
			Expect(metric.Unit).To(Equal("conn"))
			return metric.Value
		}, 2*time.Second).Should(
			BeNumerically(">=", 20),
		)

		By("Having multiple queries active")
		execQueryFunc("select sleep(1);")

		Eventually(func() float64 {
			collectedMetrics, err = metricsCollector.Collect()
			Expect(err).NotTo(HaveOccurred())
			metric = getMetricByKey(collectedMetrics, "threads_running")
			Expect(metric).ToNot(BeNil())
			Expect(metric.Unit).To(Equal("conn"))
			return metric.Value
		}, 2*time.Second).Should(
			BeNumerically(">=", 20),
		)

		By("Closing again the connections")
		cancel()

		Eventually(func() float64 {
			collectedMetrics, err = metricsCollector.Collect()
			Expect(err).NotTo(HaveOccurred())
			metric = getMetricByKey(collectedMetrics, "threads_connected")
			Expect(metric).ToNot(BeNil())
			Expect(metric.Unit).To(Equal("conn"))
			return metric.Value
		}, 2*time.Second).Should(
			BeNumerically("<=", initialConnections+5),
		)
	})
})
