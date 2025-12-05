package vote_test

import (
	"errors"
	"testing"

	"polling-system/internal/domain/poll"
	"polling-system/internal/domain/vote"
	"polling-system/internal/worker"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)


var ErrAlreadyVoted = errors.New("user already voted in this poll")
var ErrPollNotActive = errors.New("poll is not active or not found")
var ErrPollNotFound = errors.New("poll not found in repo") 


type MockVoteRepo struct {
	mock.Mock
}

func (m *MockVoteRepo) Create(v *vote.Vote) error {
	args := m.Called(v)
	return args.Error(0)
}

func (m *MockVoteRepo) HasUserVoted(pollID, userID int64) (bool, error) {
	args := m.Called(pollID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockVoteRepo) GetResults(pollID int64) ([]vote.Result, error) {
	args := m.Called(pollID)
	return args.Get(0).([]vote.Result), args.Error(1)
}
func (m *MockVoteRepo) CountByPoll(pollID int64) (map[int64]int64, int64, error) {
	args := m.Called(pollID)
	return args.Get(0).(map[int64]int64), args.Get(1).(int64), args.Error(2)
}

type MockPollService struct {
	mock.Mock
}

func (m *MockPollService) GetByID(id int64) (*poll.Poll, []poll.Option, error) {
	args := m.Called(id)
	return args.Get(0).(*poll.Poll), args.Get(1).([]poll.Option), args.Error(2)
}

func (m *MockPollService) CheckStatus(pollID int64) (string, error) {
	args := m.Called(pollID)
	return args.String(0), args.Error(1)
}



func TestService_Vote(t *testing.T) {
	voteRepo := new(MockVoteRepo)
	pollSvc := new(MockPollService)

	voteCh := make(chan worker.VoteEvent, 1)

	service := vote.NewService(voteRepo, pollSvc)

	const userID = int64(1)
	const pollID = int64(10)
	const optionID = int64(101)


	t.Run("Success_FirstVote", func(t *testing.T) {
		pollSvc.On("CheckStatus", pollID).Return("active", nil).Once()
		voteRepo.On("HasUserVoted", pollID, userID).Return(false, nil).Once()
		voteRepo.On("Create", mock.AnythingOfType("*vote.Vote")).Return(nil).Once()

		err := service.Vote(pollID, optionID, userID)

		assert.NoError(t, err)

		voteRepo.AssertCalled(t, "Create", mock.AnythingOfType("*vote.Vote"))

		select {
		case event := <-voteCh:
			assert.Equal(t, pollID, event.PollID)
			assert.Equal(t, optionID, event.OptionID)
		default:
		}

		pollSvc.AssertExpectations(t)
		voteRepo.AssertExpectations(t)
	})

	t.Run("Fail_AlreadyVoted", func(t *testing.T) {
		pollSvc.On("CheckStatus", pollID).Return("active", nil).Once()
		voteRepo.On("HasUserVoted", pollID, userID).Return(true, nil).Once()

		err := service.Vote(pollID, optionID, userID)

		assert.Equal(t, ErrAlreadyVoted.Error(), err.Error())

	})

	
	t.Run("Fail_InactivePoll", func(t *testing.T) {
		pollSvc.On("CheckStatus", pollID).Return("closed", nil).Once()

		err := service.Vote(pollID, optionID, userID)

		assert.Equal(t, ErrPollNotActive.Error(), err.Error())

		voteRepo.AssertNotCalled(t, "HasUserVoted")
	})

	t.Run("Fail_PollNotFound", func(t *testing.T) {
		
		pollSvc.On("CheckStatus", pollID).Return("", ErrPollNotFound).Once()

		err := service.Vote(pollID, optionID, userID)

		assert.Equal(t, ErrPollNotActive.Error(), err.Error())

		voteRepo.AssertNotCalled(t, "HasUserVoted")
	})
}
