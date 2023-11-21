package integration_rds_metric_collector_test

import (
	"fmt"
	"os"
	"testing"

	"code.cloudfoundry.org/lager/v3"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	uuid "github.com/satori/go.uuid"

	rdsconfig "github.com/alphagov/paas-rds-broker/config"
	collectorconfig "github.com/alphagov/paas-rds-metric-collector/pkg/config"

	"path"

	"code.cloudfoundry.org/locket"
	fakeLoggregator "github.com/alphagov/paas-rds-metric-collector/testhelpers/loggregator"
	. "github.com/alphagov/paas-rds-broker/ci/helpers"
	"github.com/alphagov/paas-rds-metric-collector/testhelpers"
	"github.com/onsi/gomega/gbytes"
)

var (
	rdsSubnetGroupName *string
	ec2SecurityGroupID *string

	rdsBrokerConfig  *rdsconfig.Config
	rdsBrokerSession *gexec.Session
	brokerAPIClient  *BrokerAPIClient
	rdsClient        *RDSClient

	testSuiteLogger lager.Logger

	mockLocketServerSession *gexec.Session

	rdsMetricCollectorPath     string
	rdsMetricCollectorConfig   *collectorconfig.Config
	rdsMetricsCollectorSession *gexec.Session

	fakeLoggregatorServer *fakeLoggregator.FakeLoggregatorIngressServer
)

func TestSuite(t *testing.T) {
	BeforeSuite(func() {
		const fixturesPath = "../../fixtures"
		var err error

		testSuiteLogger = lager.NewLogger("test-suite")
		testSuiteLogger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.INFO))

		// Compile test Locket server
		mockLocketServer := testhelpers.MockLocketServer{}
		mockLocketServer.Build()
		mockLocketServerSession = mockLocketServer.Run(fixturesPath, "alwaysGrantLock")
		Eventually(mockLocketServerSession.Buffer).Should(gbytes.Say("grpc.grpc-server.started"))

		// Update config
		rdsBrokerConfig, err = rdsconfig.LoadConfig(path.Join(fixturesPath, "broker_config.json"))
		Expect(err).ToNot(HaveOccurred())
		err = rdsBrokerConfig.Validate()
		Expect(err).ToNot(HaveOccurred())

		rdsBrokerConfig.RDSConfig.BrokerName = fmt.Sprintf("%s-%s",
			rdsBrokerConfig.RDSConfig.BrokerName,
			uuid.NewV4().String(),
		)

		awsSession := session.Must(session.NewSession(&aws.Config{
			Region: aws.String(rdsBrokerConfig.RDSConfig.Region)},
		))
		rdsSubnetGroupName, err = CreateSubnetGroup(rdsBrokerConfig.RDSConfig.DBPrefix, awsSession)
		Expect(err).ToNot(HaveOccurred())
		ec2SecurityGroupID, err = CreateSecurityGroup(rdsBrokerConfig.RDSConfig.DBPrefix, awsSession)
		Expect(err).ToNot(HaveOccurred())

		for serviceIndex := range rdsBrokerConfig.RDSConfig.Catalog.Services {
			for planIndex := range rdsBrokerConfig.RDSConfig.Catalog.Services[serviceIndex].Plans {
				plan := &rdsBrokerConfig.RDSConfig.Catalog.Services[serviceIndex].Plans[planIndex]
				plan.RDSProperties.DBSubnetGroupName = rdsSubnetGroupName
				plan.RDSProperties.VpcSecurityGroupIds = []*string{ec2SecurityGroupID}
			}
		}

		// Start a fake server for loggregator
		fakeLoggregatorServer, err = fakeLoggregator.NewFakeLoggregatorIngressServer(
			path.Join(fixturesPath, "loggregator-server.cert.pem"),
			path.Join(fixturesPath, "loggregator-server.key.pem"),
			path.Join(fixturesPath, "ca.cert.pem"),
		)
		Expect(err).ShouldNot(HaveOccurred())
		err = fakeLoggregatorServer.Start()
		Expect(err).ShouldNot(HaveOccurred())

		// Compile the rds collector
		rdsMetricCollectorPath, err = gexec.Build("github.com/alphagov/paas-rds-metric-collector")
		Expect(err).ShouldNot(HaveOccurred())

		// Update config
		rdsMetricCollectorConfig := collectorconfig.Config{
			LogLevel: "info",
			AWS: collectorconfig.AWSConfig{
				Region:       "eu-west-1",
				AWSPartition: "aws",
			},
			RDSBrokerInfo: collectorconfig.RDSBrokerInfoConfig{
				BrokerName:         rdsBrokerConfig.RDSConfig.BrokerName,
				DBPrefix:           "build-test",
				MasterPasswordSeed: "something-secret",
			},
			Scheduler: collectorconfig.SchedulerConfig{
				InstanceRefreshInterval:    30,
				SQLMetricCollectorInterval: 5,
				CWMetricCollectorInterval:  5,
			},
			LoggregatorEmitter: collectorconfig.LoggregatorEmitterConfig{
				MetronURL:  fakeLoggregatorServer.Addr,
				CACertPath: path.Join(fixturesPath, "ca.cert.pem"),
				CertPath:   path.Join(fixturesPath, "client.cert.pem"),
				KeyPath:    path.Join(fixturesPath, "client.key.pem"),
			},
			ClientLocketConfig: locket.ClientLocketConfig{
				LocketCACertFile:     path.Join(fixturesPath, "ca.cert.pem"),
				LocketClientCertFile: path.Join(fixturesPath, "client.cert.pem"),
				LocketClientKeyFile:  path.Join(fixturesPath, "client.key.pem"),
				LocketAddress:        mockLocketServer.ListenAddress,
			},
		}
		Expect(err).ToNot(HaveOccurred())

		// Start the services
		rdsBrokerSession, brokerAPIClient, rdsClient = startNewBroker(rdsBrokerConfig)
		rdsMetricsCollectorSession = startNewCollector(&rdsMetricCollectorConfig)
	})

	AfterSuite(func() {
		if fakeLoggregatorServer != nil {
			fakeLoggregatorServer.Stop()
		}
		if rdsBrokerSession != nil {
			rdsBrokerSession.Kill()
		}
		if rdsMetricsCollectorSession != nil {
			rdsMetricsCollectorSession.Kill()
		}

		awsSession := session.New(&aws.Config{
			Region: aws.String(rdsBrokerConfig.RDSConfig.Region)},
		)
		if ec2SecurityGroupID != nil {
			Expect(DestroySecurityGroup(ec2SecurityGroupID, awsSession, testSuiteLogger)).To(Succeed())
		}
		if rdsSubnetGroupName != nil {
			Expect(DestroySubnetGroup(rdsSubnetGroupName, awsSession, testSuiteLogger)).To(Succeed())
		}
		mockLocketServerSession.Kill()
	})

	RegisterFailHandler(Fail)
	RunSpecs(t, "RDS Metric Collector Suite")
}
