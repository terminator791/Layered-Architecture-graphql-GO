package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/domain/models"
)

// MockMessageRepository is a mock implementation of repository.MessageRepository
type MockMessageRepository struct {
	mock.Mock
}

func (m *MockMessageRepository) CreateMessage(ctx context.Context, message *models.Message) (*models.Message, error) {
	args := m.Called(ctx, message)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Message), args.Error(1)
}

func (m *MockMessageRepository) GetMessagesByRoom(ctx context.Context, room string) ([]*models.Message, error) {
	args := m.Called(ctx, room)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Message), args.Error(1)
}

func (m *MockMessageRepository) GetMessageByID(ctx context.Context, id string) (*models.Message, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Message), args.Error(1)
}

func (m *MockMessageRepository) GetMessagesByRoomID(ctx context.Context, roomID string, limit, offset int) ([]*models.Message, error) {
	args := m.Called(ctx, roomID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Message), args.Error(1)
}

func (m *MockMessageRepository) UpdateMessage(ctx context.Context, message *models.Message) (*models.Message, error) {
	args := m.Called(ctx, message)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Message), args.Error(1)
}

func (m *MockMessageRepository) DeleteMessage(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMessageRepository) SearchMessages(ctx context.Context, query string, roomID *string, limit, offset int) ([]*models.Message, error) {
	args := m.Called(ctx, query, roomID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Message), args.Error(1)
}

func (m *MockMessageRepository) GetMessageReactions(ctx context.Context, messageID string) ([]*models.MessageReaction, error) {
	args := m.Called(ctx, messageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.MessageReaction), args.Error(1)
}

func (m *MockMessageRepository) AddMessageReaction(ctx context.Context, reaction *models.MessageReaction) (*models.MessageReaction, error) {
	args := m.Called(ctx, reaction)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.MessageReaction), args.Error(1)
}

func (m *MockMessageRepository) RemoveMessageReaction(ctx context.Context, messageID, userID, emoji string) error {
	args := m.Called(ctx, messageID, userID, emoji)
	return args.Error(0)
}

func (m *MockMessageRepository) GetReactionCounts(ctx context.Context, messageID string) (map[string]int, error) {
	args := m.Called(ctx, messageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int), args.Error(1)
}

// MockUserRepository is a mock implementation of repository.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, user *models.User) (*models.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUserStatus(ctx context.Context, userID string, status models.UserStatus) error {
	args := m.Called(ctx, userID, status)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateUserLastSeen(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteUser(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUsersByIDs(ctx context.Context, ids []string) ([]*models.User, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

// MockRoomRepository is a mock implementation of repository.RoomRepository
type MockRoomRepository struct {
	mock.Mock
}

func (m *MockRoomRepository) CreateRoom(ctx context.Context, room *models.Room) (*models.Room, error) {
	args := m.Called(ctx, room)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Room), args.Error(1)
}

func (m *MockRoomRepository) GetRoomByID(ctx context.Context, id string) (*models.Room, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Room), args.Error(1)
}

func (m *MockRoomRepository) UpdateRoom(ctx context.Context, room *models.Room) (*models.Room, error) {
	args := m.Called(ctx, room)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Room), args.Error(1)
}

func (m *MockRoomRepository) DeleteRoom(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRoomRepository) ListRooms(ctx context.Context, limit, offset int) ([]*models.Room, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Room), args.Error(1)
}

func (m *MockRoomRepository) GetRoomsByUserID(ctx context.Context, userID string) ([]*models.Room, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Room), args.Error(1)
}

func (m *MockRoomRepository) GetRoomMembers(ctx context.Context, roomID string) ([]*models.RoomMember, error) {
	args := m.Called(ctx, roomID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.RoomMember), args.Error(1)
}

func (m *MockRoomRepository) GetRoomMember(ctx context.Context, roomID, userID string) (*models.RoomMember, error) {
	args := m.Called(ctx, roomID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RoomMember), args.Error(1)
}

func (m *MockRoomRepository) AddRoomMember(ctx context.Context, member *models.RoomMember) (*models.RoomMember, error) {
	args := m.Called(ctx, member)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RoomMember), args.Error(1)
}

func (m *MockRoomRepository) UpdateRoomMember(ctx context.Context, member *models.RoomMember) (*models.RoomMember, error) {
	args := m.Called(ctx, member)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RoomMember), args.Error(1)
}

func (m *MockRoomRepository) RemoveRoomMember(ctx context.Context, roomID, userID string) error {
	args := m.Called(ctx, roomID, userID)
	return args.Error(0)
}

func (m *MockRoomRepository) GetRoomMemberCount(ctx context.Context, roomID string) (int, error) {
	args := m.Called(ctx, roomID)
	return args.Int(0), args.Error(1)
}

func (m *MockRoomRepository) GetOnlineMemberCount(ctx context.Context, roomID string) (int, error) {
	args := m.Called(ctx, roomID)
	return args.Int(0), args.Error(1)
}

func (m *MockRoomRepository) UpdateMemberLastRead(ctx context.Context, roomID, userID string) error {
	args := m.Called(ctx, roomID, userID)
	return args.Error(0)
}

// MockTypingRepository is a mock implementation of repository.TypingRepository
type MockTypingRepository struct {
	mock.Mock
}

func (m *MockTypingRepository) StartTyping(ctx context.Context, roomID, userID string) error {
	args := m.Called(ctx, roomID, userID)
	return args.Error(0)
}

func (m *MockTypingRepository) StopTyping(ctx context.Context, roomID, userID string) error {
	args := m.Called(ctx, roomID, userID)
	return args.Error(0)
}

func (m *MockTypingRepository) GetTypingUsers(ctx context.Context, roomID string) ([]*models.TypingIndicator, error) {
	args := m.Called(ctx, roomID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TypingIndicator), args.Error(1)
}

func (m *MockTypingRepository) CleanExpiredTypingIndicators(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}