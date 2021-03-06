package ui

import (
	"time"

	biuifmt "github.com/cloudfoundry/bosh-init/ui/fmt"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/pivotal-golang/clock"
)

type Stage interface {
	Perform(name string, closure func() error) error
	PerformComplex(name string, closure func(Stage) error) error
}

type stage struct {
	ui          UI
	timeService clock.Clock
	logger      boshlog.Logger
	logTag      string

	simpleMode bool
}

func NewStage(ui UI, timeService clock.Clock, logger boshlog.Logger) Stage {
	return &stage{
		ui:          ui,
		timeService: timeService,
		logger:      logger,
		logTag:      "stage",
		simpleMode:  true,
	}
}

func (s *stage) Perform(name string, closure func() error) error {
	if !s.simpleMode {
		// enter simple mode (only line break if exiting complex mode)
		s.ui.PrintLinef("")
		s.simpleMode = true
	}

	s.ui.BeginLinef("%s...", name)
	startTime := s.timeService.Now()
	err := closure()
	if err != nil {
		if skipErr, ok := err.(SkipStageError); ok {
			s.ui.EndLinef(" Skipped [%s] (%s)", skipErr.SkipMessage(), s.elapsedSince(startTime))
			s.logger.Info("Skipped stage '%s': %s", name, skipErr.Error())
			return nil
		}
		s.ui.EndLinef(" Failed (%s)", s.elapsedSince(startTime))
		return err
	}
	s.ui.EndLinef(" Finished (%s)", s.elapsedSince(startTime))
	return nil
}

func (s *stage) PerformComplex(name string, closure func(Stage) error) error {
	// exit simple mode (always line break when entering a new complex stage)
	s.ui.PrintLinef("")
	s.simpleMode = false

	s.ui.PrintLinef("Started %s", name)
	startTime := s.timeService.Now()
	err := closure(s.newSubStage())
	if err != nil {
		s.ui.PrintLinef("Failed %s (%s)", name, s.elapsedSince(startTime))
		return err
	}
	s.ui.PrintLinef("Finished %s (%s)", name, s.elapsedSince(startTime))
	return nil
}

func (s *stage) elapsedSince(startTime time.Time) string {
	stopTime := s.timeService.Now()
	duration := stopTime.Sub(startTime)
	return biuifmt.Duration(duration)
}

func (s *stage) newSubStage() Stage {
	return NewStage(NewIndentingUI(s.ui), s.timeService, s.logger)
}
