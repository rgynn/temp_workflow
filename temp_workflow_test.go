package temp_workflow

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"go.temporal.io/sdk/testsuite"
)

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

func (s *UnitTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

func (s *UnitTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *UnitTestSuite) Test_Workflow_Success() {
	raised_event := &Event{Status: "To Do"}
	update_event := &Event{Status: "In Progress"}
	done_event := &Event{Status: "Done"}
	s.env.RegisterWorkflow(Workflow)
	s.env.RegisterActivity(ActivityOne)
	s.env.RegisterActivity(ActivityTwo)
	s.env.OnActivity(ActivityOne, mock.Anything, raised_event).Return("created_id", nil)
	s.env.OnActivity(ActivityTwo, mock.Anything, update_event).Return(nil)
	s.env.OnActivity(ActivityTwo, mock.Anything, done_event).Return(nil)
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(UPDATE_CHANNEL, update_event)
	}, time.Minute)
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(UPDATE_CHANNEL, done_event)
	}, 2*time.Minute)
	s.env.ExecuteWorkflow(Workflow, raised_event)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}
