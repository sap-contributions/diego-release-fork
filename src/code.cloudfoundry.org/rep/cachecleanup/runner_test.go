package cachecleanup_test

import (
	"os"
	"time"

	"code.cloudfoundry.org/clock/fakeclock"
	fakeexecutor "code.cloudfoundry.org/executor/fakes"
	"code.cloudfoundry.org/lager/v3/lagertest"
	"code.cloudfoundry.org/rep/cachecleanup"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/ifrit"
	ginkgomon "github.com/tedsuo/ifrit/ginkgomon_v2"
)

var _ = Describe("Runner", func() {
	var (
		runner          *cachecleanup.Runner
		process         ifrit.Process
		logger          *lagertest.TestLogger
		executorClient  *fakeexecutor.FakeClient
		fakeClock       *fakeclock.FakeClock
		cleanupInterval time.Duration
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("cachecleanup-test")
		executorClient = &fakeexecutor.FakeClient{}
		fakeClock = fakeclock.NewFakeClock(time.Now())
		cleanupInterval = 30 * time.Minute
	})

	JustBeforeEach(func() {
		runner = cachecleanup.NewRunner(logger, fakeClock, cleanupInterval, executorClient)
		process = ifrit.Background(runner)
	})

	AfterEach(func() {
		ginkgomon.Interrupt(process)
	})

	It("becomes ready immediately", func() {
		Eventually(process.Ready()).Should(BeClosed())
	})

	It("does not call ReclaimCacheSpace before the first tick", func() {
		Eventually(process.Ready()).Should(BeClosed())
		Expect(executorClient.ReclaimCacheSpaceCallCount()).To(Equal(0))
	})

	Context("when the ticker ticks", func() {
		JustBeforeEach(func() {
			Eventually(process.Ready()).Should(BeClosed())
		})

		It("calls ReclaimCacheSpace on each tick", func() {
			fakeClock.WaitForWatcherAndIncrement(cleanupInterval)
			Eventually(executorClient.ReclaimCacheSpaceCallCount).Should(Equal(1))

			fakeClock.WaitForWatcherAndIncrement(cleanupInterval)
			Eventually(executorClient.ReclaimCacheSpaceCallCount).Should(Equal(2))
		})

		It("passes the logger to ReclaimCacheSpace", func() {
			fakeClock.WaitForWatcherAndIncrement(cleanupInterval)
			Eventually(executorClient.ReclaimCacheSpaceCallCount).Should(Equal(1))
			passedLogger := executorClient.ReclaimCacheSpaceArgsForCall(0)
			Expect(passedLogger).NotTo(BeNil())
		})
	})

	Context("when signalled", func() {
		JustBeforeEach(func() {
			Eventually(process.Ready()).Should(BeClosed())
		})

		It("exits with nil", func() {
			process.Signal(os.Interrupt)
			Eventually(process.Wait()).Should(Receive(BeNil()))
		})
	})
})
