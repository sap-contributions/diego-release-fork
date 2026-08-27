package containerstore

import (
	"code.cloudfoundry.org/executor"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("injectWorkloadIdentity", func() {
	var (
		baseConfig map[string]interface{}
		container  executor.Container
	)

	BeforeEach(func() {
		baseConfig = map[string]interface{}{
			"source":         "nfs://10.0.0.1/share",
			"mount-endpoint": "https://fss.example.com",
		}
	})

	Context("when the container is an LRP", func() {
		BeforeEach(func() {
			container = executor.Container{
				Guid: "instance-guid-abc",
				Tags: executor.Tags{
					"lifecycle":    "lrp",
					"process-guid": "app-guid-123",
				},
			}
		})

		It("injects _workload_guid with the process-guid and _workload_type as lrp", func() {
			result := injectWorkloadIdentity(baseConfig, container)
			Expect(result["_workload_guid"]).To(Equal("app-guid-123"))
			Expect(result["_workload_type"]).To(Equal("lrp"))
		})

		It("preserves existing config keys", func() {
			result := injectWorkloadIdentity(baseConfig, container)
			Expect(result["source"]).To(Equal("nfs://10.0.0.1/share"))
			Expect(result["mount-endpoint"]).To(Equal("https://fss.example.com"))
		})

		It("does not mutate the original config map", func() {
			injectWorkloadIdentity(baseConfig, container)
			Expect(baseConfig).NotTo(HaveKey("_workload_guid"))
			Expect(baseConfig).NotTo(HaveKey("_workload_type"))
		})
	})

	Context("when the container is a Task", func() {
		BeforeEach(func() {
			container = executor.Container{
				Guid: "task-guid-xyz",
				Tags: executor.Tags{
					"lifecycle": "task",
				},
			}
		})

		It("injects _workload_guid with the task guid and _workload_type as task", func() {
			result := injectWorkloadIdentity(baseConfig, container)
			Expect(result["_workload_guid"]).To(Equal("task-guid-xyz"))
			Expect(result["_workload_type"]).To(Equal("task"))
		})
	})

	Context("when the container has no lifecycle tag", func() {
		BeforeEach(func() {
			container = executor.Container{
				Guid: "some-guid",
				Tags: executor.Tags{},
			}
		})

		It("returns the original config unchanged", func() {
			result := injectWorkloadIdentity(baseConfig, container)
			Expect(result).NotTo(HaveKey("_workload_guid"))
			Expect(result).NotTo(HaveKey("_workload_type"))
			Expect(result).To(Equal(baseConfig))
		})
	})

	Context("when the lifecycle tag has an unrecognized value", func() {
		BeforeEach(func() {
			container = executor.Container{
				Guid: "some-guid",
				Tags: executor.Tags{
					"lifecycle": "other",
				},
			}
		})

		It("injects no identity keys and preserves the existing config", func() {
			result := injectWorkloadIdentity(baseConfig, container)
			Expect(result).NotTo(HaveKey("_workload_guid"))
			Expect(result).NotTo(HaveKey("_workload_type"))
			Expect(result["source"]).To(Equal("nfs://10.0.0.1/share"))
			Expect(result["mount-endpoint"]).To(Equal("https://fss.example.com"))
		})

		It("does not mutate the original config map", func() {
			injectWorkloadIdentity(baseConfig, container)
			Expect(baseConfig).NotTo(HaveKey("_workload_guid"))
			Expect(baseConfig).NotTo(HaveKey("_workload_type"))
		})
	})

	Context("when config is nil", func() {
		It("returns a new map with identity keys for LRPs", func() {
			container = executor.Container{
				Guid: "instance-guid",
				Tags: executor.Tags{
					"lifecycle":    "lrp",
					"process-guid": "pg-1",
				},
			}
			result := injectWorkloadIdentity(nil, container)
			Expect(result["_workload_guid"]).To(Equal("pg-1"))
			Expect(result["_workload_type"]).To(Equal("lrp"))
		})
	})
})
