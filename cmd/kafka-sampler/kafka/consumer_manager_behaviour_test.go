package kafka

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	mock_kafka "github.com/neblic/platform/cmd/kafka-sampler/kafka/mock"
	"github.com/neblic/platform/logging"

	. "github.com/onsi/ginkgo/v2"

	. "github.com/onsi/gomega"
)

func TestConsumerManager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Manager behaviour")
}

var _ = Describe("Manager behaviour", func() {
	var (
		logger   logging.Logger
		mockCtrl *gomock.Controller
		client   *mock_kafka.MockClient
		group    *mock_kafka.MockConsumerGroup
		manager  *ConsumerManager
	)
	//initialization
	BeforeEach(func() {
		defer GinkgoRecover()

		logger, _ = logging.NewZapDev()
		mockCtrl = gomock.NewController(GinkgoT())
		client = mock_kafka.NewMockClient(mockCtrl)
		group = mock_kafka.NewMockConsumerGroup(mockCtrl)
		manager = &ConsumerManager{
			ctx:    context.Background(),
			logger: logger,
			config: NewConfig(),
			client: client,
			groupProvider: func(topic string) (ConsumerGroup, error) {
				group := mock_kafka.NewMockConsumerGroup(mockCtrl)
				group.EXPECT().Consume(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
				group.EXPECT().Close().AnyTimes().Return(nil)
				return group, nil
			},
			consumers: map[string]*consumerInstance{},
		}
	})

	When("reconcile with a map of topics is called", func() {
		BeforeEach(func(ctx SpecContext) {
			// Prepare group
			group.EXPECT().Consume(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			// Add some topics
			err := manager.reconcile([]string{"topic1"})
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("must add all new topics", func() {
			err := manager.reconcile([]string{"topic1", "topic2"})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(manager.Topics()).Should(ContainElements([]string{"topic1", "topic2"}))
		})

		It("must remove topics that no longer exist", func() {
			err := manager.reconcile([]string{"topic2"})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(manager.Topics()).Should(ContainElements([]string{"topic2"}))
		})
	})

	//tear down
	AfterEach(func() {
		// We add this in order to check that all the registered mocks ware really called
		mockCtrl.Finish()
	})
})
