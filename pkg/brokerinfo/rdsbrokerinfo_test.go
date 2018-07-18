package brokerinfo_test

import (
	"fmt"

	"github.com/alphagov/paas-rds-broker/awsrds"
	rdsfake "github.com/alphagov/paas-rds-broker/awsrds/fakes"

	"github.com/alphagov/paas-rds-metric-collector/pkg/brokerinfo"
	"github.com/alphagov/paas-rds-metric-collector/pkg/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RDSBrokerInfo", func() {
	var (
		brokerInfo     *brokerinfo.RDSBrokerInfo
		fakeDBInstance *rdsfake.FakeDBInstance
	)

	BeforeEach(func() {
		fakeDBInstance = &rdsfake.FakeDBInstance{}
		brokerInfo = brokerinfo.NewRDSBrokerInfo(
			config.RDSBrokerInfoConfig{
				BrokerName:         "broker_name",
				DBPrefix:           "dbprefix",
				MasterPasswordSeed: "12345",
			},
			fakeDBInstance,
			logger,
		)
	})

	Context("ListInstances()", func() {
		BeforeEach(func() {
			fakeDBInstance.DescribeByTagDBInstanceDetails = []*awsrds.DBInstanceDetails{
				&awsrds.DBInstanceDetails{
					Identifier:     "dbprefix-instance-id-1",
					Engine:         "postgres",
					Address:        "endpoint-address-1.example.com",
					Port:           5432,
					DBName:         "dbprefix-db",
					MasterUsername: "master-username",
				},
				&awsrds.DBInstanceDetails{
					Identifier:     "dbprefix-instance-id-2",
					Engine:         "postgres",
					Address:        "endpoint-address-2.example.com",
					Port:           5432,
					DBName:         "dbprefix-db",
					MasterUsername: "master-username",
				},
				&awsrds.DBInstanceDetails{
					Identifier:     "dbprefix-instance-id-3",
					Engine:         "mysql",
					Address:        "endpoint-address-3.example.com",
					Port:           3306,
					DBName:         "dbprefix-db",
					MasterUsername: "master-username",
				},
			}
		})

		It("returns error if it fails retrieving existing instances in AWS", func() {
			fakeDBInstance.DescribeByTagError = fmt.Errorf("Error calling rds.DescribeByTag(...)")

			_, err := brokerInfo.ListInstances()
			Expect(err).To(HaveOccurred())
		})
		It("lists the instances for the right tag", func() {
			_, err := brokerInfo.ListInstances()
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeDBInstance.DescribeByTagKey).To(Equal("Broker Name"))
			Expect(fakeDBInstance.DescribeByTagValue).To(Equal("broker_name"))
		})
		It("returns the list of instances", func() {
			instanceGUIDs, err := brokerInfo.ListInstances()
			Expect(err).NotTo(HaveOccurred())
			Expect(instanceGUIDs).To(ConsistOf(
				brokerinfo.InstanceInfo{GUID: "instance-id-1", Type: "postgres"},
				brokerinfo.InstanceInfo{GUID: "instance-id-2", Type: "postgres"},
				brokerinfo.InstanceInfo{GUID: "instance-id-3", Type: "mysql"},
			))
		})
	})

	Context("ConnectionString()", func() {
		BeforeEach(func() {
			fakeDBInstance.DescribeDBInstanceDetails = awsrds.DBInstanceDetails{
				Identifier:     "dbprefix-instance-id",
				Address:        "endpoint-address.example.com",
				Port:           5432,
				DBName:         "dbprefix-db",
				MasterUsername: "master-username",
			}
		})

		It("returns error if it fails retrieving existing instances in AWS", func() {
			fakeDBInstance.DescribeError = fmt.Errorf("Error calling rds.Describe(...)")

			_, err := brokerInfo.ConnectionString(brokerinfo.InstanceInfo{GUID: "instance-id", Type: "postgres"})
			Expect(err).To(HaveOccurred())
		})
		It("returns error if we query the wrong instance type", func() {
			_, err := brokerInfo.ConnectionString(brokerinfo.InstanceInfo{GUID: "instance-id", Type: "foo"})
			Expect(err).To(HaveOccurred())
		})
		It("retrieves information of the right AWS RDS instance", func() {
			_, err := brokerInfo.ConnectionString(brokerinfo.InstanceInfo{GUID: "instance-id", Type: "postgres"})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeDBInstance.DescribeCalled).To(BeTrue())
			Expect(fakeDBInstance.DescribeID).To(Equal("dbprefix-instance-id"))
		})
		It("returns the proper connection string", func() {
			connectionString, err := brokerInfo.ConnectionString(brokerinfo.InstanceInfo{GUID: "instance-id", Type: "postgres"})
			Expect(err).ToNot(HaveOccurred())
			Expect(connectionString).To(Equal("postgresql://master-username:9Fs6CWnuwf0BAY3rDFAels3OXANSo0-M@endpoint-address.example.com:5432/dbprefix-db?sslmode=require"))
		})
		It("fails if the type is invalid", func() {
			_, err := brokerInfo.ConnectionString(brokerinfo.InstanceInfo{GUID: "instance-id", Type: "foo"})
			Expect(err).To(HaveOccurred())
		})
	})
})
