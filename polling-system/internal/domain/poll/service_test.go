package poll_test

import (
	"errors"
	"testing"

	"polling-system/internal/domain/poll"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)


	mock.Mock
}


func (m *MockPollRepo) Create(p *poll.Poll, options []poll.Option) (int64, error) {
	args := m.Called(p, options)
	
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPollRepo) GetByID(id int64) (*poll.Poll, []poll.Option, error) {
	
	args := m.Called(id)
	return args.Get(0).(*poll.Poll), args.Get(1).([]poll.Option), args.Error(2)
}

func (m *MockPollRepo) List(status *string) ([]poll.Poll, error) {
	args := m.Called(status)
	return args.Get(0).([]poll.Poll), args.Error(1)
}

func (m *MockPollRepo) UpdateStatus(id int64, status string) error {
	args := m.Called(id, status)
	return args.Error(0)
}


func TestService_Create(t *testing.T) {
	mockRepo := new(MockPollRepo)
	service := poll.NewService(mockRepo)
	creatorID := int64(1)


	t.Run("Success_ValidPoll", func(t *testing.T) {
		newPoll := &poll.Poll{Title: "Favorite Color", CreatorID: creatorID}
		options := []poll.Option{{Text: "Red"}, {Text: "Blue"}}
		expectedID := int64(10)

		
		mockRepo.On("Create", newPoll, options).Return(expectedID, nil).Once()

		id, err := service.Create(newPoll, options)

		
		assert.NoError(t, err)
		assert.Equal(t, expectedID, id)
		
		assert.Equal(t, "draft", newPoll.Status)

		
		mockRepo.AssertCalled(t, "Create", newPoll, options)
	})

	
	t.Run("Fail_MissingTitle", func(t *testing.T) {
		p := &poll.Poll{Title: "", CreatorID: creatorID} 
		options := []poll.Option{{Text: "R"}, {Text: "B"}}

		id, err := service.Create(p, options)

		
		assert.Error(t, err)
		assert.Equal(t, "title required", err.Error())
		assert.Equal(t, int64(0), id)

		
		mockRepo.AssertNotCalled(t, "Create")
	})

	
	t.Run("Fail_TooFewOptions", func(t *testing.T) {
		p := &poll.Poll{Title: "Test", CreatorID: creatorID}
		options := []poll.Option{{Text: "Only one"}}

		id, err := service.Create(p, options)

	
		assert.Error(t, err)
		assert.Equal(t, "poll must have at least 2 options", err.Error())
		assert.Equal(t, int64(0), id)

		mockRepo.AssertNotCalled(t, "Create")
	})
}



func TestService_UpdateStatus(t *testing.T) {
	mockRepo := new(MockPollRepo)
	service := poll.NewService(mockRepo)
	pollID := int64(5)

	
	t.Run("Success_ToActive", func(t *testing.T) {
		newStatus := "active"
	
		mockRepo.On("UpdateStatus", pollID, newStatus).Return(nil).Once()

		err := service.UpdateStatus(pollID, newStatus)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	
	t.Run("Success_ToClosed", func(t *testing.T) {
		newStatus := "closed"
		mockRepo.On("UpdateStatus", pollID, newStatus).Return(nil).Once()

		err := service.UpdateStatus(pollID, newStatus)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	
	t.Run("Fail_InvalidStatus", func(t *testing.T) {
		invalidStatus := "published" 

		err := service.UpdateStatus(pollID, invalidStatus)

		
		assert.ErrorIs(t, err, poll.ErrInvalidStatus)

		
		mockRepo.AssertNotCalled(t, "UpdateStatus", pollID, invalidStatus)
	})


	t.Run("Fail_RepositoryError", func(t *testing.T) {
		dbError := errors.New("db connection failed")
		newStatus := "draft"
	
		mockRepo.On("UpdateStatus", pollID, newStatus).Return(dbError).Once()

		err := service.UpdateStatus(pollID, newStatus)

		
		assert.ErrorIs(t, err, dbError)
		mockRepo.AssertExpectations(t)
	})
}
