package installation_test

import (
	. "github.com/cloudfoundry/bosh-init/installation"
	. "github.com/cloudfoundry/bosh-init/internal/github.com/onsi/ginkgo"
	. "github.com/cloudfoundry/bosh-init/internal/github.com/onsi/gomega"

	biconfig "github.com/cloudfoundry/bosh-init/config"
	boshlog "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/logger"
	fakesys "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/uuid/fakes"
)

var _ = Describe("TargetProvider", func() {
	var (
		fakeFS                 *fakesys.FakeFileSystem
		fakeUUIDGenerator      *fakeuuid.FakeGenerator
		logger                 boshlog.Logger
		deploymentStateService biconfig.DeploymentStateService

		targetProvider TargetProvider

		configPath            = "/deployment.json"
		installationsRootPath = "/.bosh_init/installations"
	)

	BeforeEach(func() {
		fakeFS = fakesys.NewFakeFileSystem()
		fakeUUIDGenerator = fakeuuid.NewFakeGenerator()
		logger = boshlog.NewLogger(boshlog.LevelNone)
		deploymentStateService = biconfig.NewFileSystemDeploymentStateService(
			fakeFS,
			fakeUUIDGenerator,
			logger,
			configPath,
		)
		targetProvider = NewTargetProvider(deploymentStateService, fakeUUIDGenerator, installationsRootPath)
	})

	Context("when the installation_id exists in the deployment state", func() {
		BeforeEach(func() {
			err := fakeFS.WriteFileString(configPath, `{"installation_id":"12345"}`)
			Expect(err).ToNot(HaveOccurred())
		})

		It("uses the existing installation_id & returns a target based on it", func() {
			target, err := targetProvider.NewTarget()
			Expect(err).ToNot(HaveOccurred())
			Expect(target.Path()).To(Equal("/.bosh_init/installations/12345"))
		})

		It("does not change the saved installation_id", func() {
			_, err := targetProvider.NewTarget()
			Expect(err).ToNot(HaveOccurred())

			deploymentState, err := deploymentStateService.Load()
			Expect(err).ToNot(HaveOccurred())
			Expect(deploymentState.InstallationID).To(Equal("12345"))
		})
	})

	Context("when the installation_id does not exist in the deployment state", func() {
		BeforeEach(func() {
			err := fakeFS.WriteFileString(configPath, `{}`)
			Expect(err).ToNot(HaveOccurred())
		})

		It("generates a new installation_id & returns a target based on it", func() {
			target, err := targetProvider.NewTarget()
			Expect(err).ToNot(HaveOccurred())
			Expect(target.Path()).To(Equal("/.bosh_init/installations/fake-uuid-1"))
		})

		It("saves the new installation_id", func() {
			_, err := targetProvider.NewTarget()
			Expect(err).ToNot(HaveOccurred())

			deploymentState, err := deploymentStateService.Load()
			Expect(err).ToNot(HaveOccurred())
			Expect(deploymentState.InstallationID).To(Equal("fake-uuid-1"))
		})
	})

	Context("when the deployment state does not exist", func() {
		BeforeEach(func() {
			err := fakeFS.RemoveAll(configPath)
			Expect(err).ToNot(HaveOccurred())
		})

		It("generates a new installation_id & returns a target based on it", func() {
			target, err := targetProvider.NewTarget()
			Expect(err).ToNot(HaveOccurred())
			Expect(target.Path()).To(Equal("/.bosh_init/installations/fake-uuid-1"))
		})

		It("saves the new installation_id", func() {
			_, err := targetProvider.NewTarget()
			Expect(err).ToNot(HaveOccurred())

			deploymentState, err := deploymentStateService.Load()
			Expect(err).ToNot(HaveOccurred())
			Expect(deploymentState.InstallationID).To(Equal("fake-uuid-1"))
		})
	})
})