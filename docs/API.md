# 🚀 GraphQL API Documentation

A comprehensive guide to the GraphQL API for the real-time chat application. This document serves as both a reference and tutorial for developers integrating with the API.

## 📋 Table of Contents

1. [API Overview](#api-overview)
2. [Authentication](#authentication)
3. [User Management](#user-management)
4. [Room Management](#room-management)
5. [Messaging System](#messaging-system)
6. [Real-time Subscriptions](#real-time-subscriptions)
7. [Error Handling](#error-handling)
8. [Best Practices](#best-practices)
9. [Rate Limiting](#rate-limiting)
10. [Examples & Tutorials](#examples--tutorials)

## 🎯 API Overview

### GraphQL Endpoint

```
POST /graphql
```

### Key Features

- **Type Safety**: Strong typing with GraphQL schema
- **Real-time Subscriptions**: WebSocket-based live updates
- **Flexible Queries**: Request exactly the data you need
- **Introspective**: Self-documenting API
- **Authentication**: JWT-based secure access

### Schema Highlights

- **Users**: Registration, profiles, status management
- **Rooms**: Public/private rooms, member management, roles
- **Messages**: Rich messaging with reactions, replies, search
- **Real-time**: Live typing indicators, presence, notifications

## 🔐 Authentication

### Registration

Create a new user account with email verification.

```graphql
mutation Register {
  register(input: {
    username: "john_doe"
    email: "john@example.com"
    password: "securepassword123"
    displayName: "John Doe"
    bio: "Software developer passionate about clean code"
    avatarUrl: "https://example.com/avatar.jpg"
  }) {
    token
    user {
      id
      username
      email
      displayName
      bio
      avatarUrl
      status
      createdAt
      updatedAt
    }
  }
}
```

**Response**:
```json
{
  "data": {
    "register": {
      "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "user": {
        "id": "user_12345",
        "username": "john_doe",
        "email": "john@example.com",
        "displayName": "John Doe",
        "bio": "Software developer passionate about clean code",
        "avatarUrl": "https://example.com/avatar.jpg",
        "status": "ONLINE",
        "createdAt": "2024-01-15T10:30:00Z",
        "updatedAt": "2024-01-15T10:30:00Z"
      }
    }
  }
}
```

### Login

Authenticate existing user and receive JWT token.

```graphql
mutation Login {
  login(input: {
    username: "john_doe"
    password: "securepassword123"
  }) {
    token
    user {
      id
      username
      displayName
      status
      lastSeenAt
    }
  }
}
```

### Authentication Headers

Include JWT token in all subsequent requests:

```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

## 👥 User Management

### Get Current User

Retrieve authenticated user's profile information.

```graphql
query Me {
  me {
    id
    username
    email
    displayName
    bio
    avatarUrl
    status
    lastSeenAt
    createdAt
    updatedAt
  }
}
```

### Update Profile

Modify user profile information.

```graphql
mutation UpdateProfile {
  updateUser(input: {
    displayName: "John Smith"
    bio: "Senior Software Engineer"
    avatarUrl: "https://example.com/new-avatar.jpg"
  }) {
    id
    displayName
    bio
    avatarUrl
    updatedAt
  }
}
```

### Update Status

Change online status for presence indication.

```graphql
mutation UpdateStatus {
  updateUserStatus(status: AWAY) {
    id
    status
    lastSeenAt
  }
}
```

**Status Options**:
- `ONLINE`: User is actively using the application
- `AWAY`: User is idle or minimized the application
- `BUSY`: User is in do-not-disturb mode
- `OFFLINE`: User has logged out

### Get User by ID

Retrieve public information about any user.

```graphql
query GetUser($userId: ID!) {
  user(id: $userId) {
    id
    username
    displayName
    avatarUrl
    bio
    status
    lastSeenAt
  }
}
```

### List Users

Get paginated list of users with search capabilities.

```graphql
query ListUsers($limit: Int, $offset: Int) {
  users(limit: $limit, offset: $offset) {
    id
    username
    displayName
    avatarUrl
    status
    lastSeenAt
  }
}
```

## 🏠 Room Management

### Create Room

Create a new chat room with specified configuration.

```graphql
mutation CreateRoom {
  createRoom(input: {
    name: "Product Team"
    description: "Daily discussions about product development"
    roomType: PRIVATE
    password: "team2024"
    maxMembers: 50
    avatarUrl: "https://example.com/room-avatar.jpg"
  }) {
    id
    name
    description
    roomType
    avatarUrl
    maxMembers
    createdAt
    creator {
      id
      username
      displayName
    }
    memberCount
    onlineCount
  }
}
```

**Room Types**:
- `PUBLIC`: Anyone can join
- `PRIVATE`: Invitation or password required
- `DIRECT`: One-on-one conversation

### Join Room

Join an existing room with optional password.

```graphql
mutation JoinRoom {
  joinRoom(input: {
    roomId: "room_67890"
    password: "team2024"
  }) {
    id
    role
    joinedAt
    user {
      id
      username
      displayName
    }
  }
}
```

### Get Room Details

Retrieve comprehensive room information.

```graphql
query GetRoom($roomId: ID!) {
  room(id: $roomId) {
    id
    name
    description
    roomType
    avatarUrl
    maxMembers
    createdAt
    updatedAt
    creator {
      id
      username
      displayName
    }
    members {
      id
      role
      joinedAt
      lastReadAt
      user {
        id
        username
        displayName
        avatarUrl
        status
      }
    }
    memberCount
    onlineCount
  }
}
```

### List My Rooms

Get all rooms the authenticated user is a member of.

```graphql
query MyRooms {
  myRooms {
    id
    name
    description
    roomType
    avatarUrl
    memberCount
    onlineCount
    updatedAt
  }
}
```

### Update Room

Modify room settings (admin/moderator only).

```graphql
mutation UpdateRoom($roomId: ID!) {
  updateRoom(roomId: $roomId, input: {
    name: "Product Development Team"
    description: "Enhanced team collaboration space"
    maxMembers: 75
  }) {
    id
    name
    description
    maxMembers
    updatedAt
  }
}
```

### Leave Room

Leave a room you're currently a member of.

```graphql
mutation LeaveRoom($roomId: ID!) {
  leaveRoom(roomId: $roomId)
}
```

### Manage Members

Update member roles or remove members (admin/moderator only).

```graphql
mutation UpdateMemberRole {
  updateRoomMember(input: {
    roomId: "room_67890"
    userId: "user_12345"
    role: MODERATOR
  }) {
    id
    role
    user {
      id
      username
      displayName
    }
  }
}

mutation KickMember {
  kickRoomMember(roomId: "room_67890", userId: "user_54321")
}
```

**Member Roles**:
- `ADMIN`: Full control over room settings and members
- `MODERATOR`: Can manage messages and moderate members
- `MEMBER`: Standard participant with message permissions

## 💬 Messaging System

### Send Message

Send a text message to a room.

```graphql
mutation SendMessage {
  sendMessageToRoom(
    roomId: "room_67890"
    text: "Hello everyone! 👋"
    messageType: TEXT
  ) {
    id
    text
    messageType
    createdAt
    userInfo {
      id
      username
      displayName
      avatarUrl
    }
    reactions {
      id
      emoji
      user {
        username
      }
    }
    reactionCount {
      emoji
      count
    }
  }
}
```

### Send Rich Message

Send a message with metadata (images, files).

```graphql
mutation SendImageMessage {
  sendMessageToRoom(
    roomId: "room_67890"
    text: "Check out this screenshot!"
    messageType: IMAGE
    metadata: {
      imageUrl: "https://example.com/screenshot.png"
      imageWidth: 1920
      imageHeight: 1080
      fileName: "screenshot.png"
      fileSize: 245760
      mimeType: "image/png"
    }
  ) {
    id
    text
    messageType
    metadata {
      imageUrl
      imageWidth
      imageHeight
      fileName
      fileSize
      mimeType
    }
    createdAt
    userInfo {
      username
      displayName
    }
  }
}
```

### Reply to Message

Send a message as a reply to another message.

```graphql
mutation ReplyToMessage {
  sendMessageToRoom(
    roomId: "room_67890"
    text: "That's a great idea!"
    replyToId: "message_98765"
  ) {
    id
    text
    replyToId
    replyTo {
      id
      text
      userInfo {
        username
      }
    }
    createdAt
    userInfo {
      username
      displayName
    }
  }
}
```

### Edit Message

Update message content (own messages only).

```graphql
mutation EditMessage {
  updateMessage(input: {
    messageId: "message_12345"
    text: "Updated message content"
  }) {
    id
    text
    editedAt
    createdAt
    userInfo {
      username
    }
  }
}
```

### Delete Message

Remove a message (soft delete).

```graphql
mutation DeleteMessage {
  deleteMessage(messageId: "message_12345")
}
```

### Get Messages

Retrieve paginated messages for a room.

```graphql
query GetMessages($roomId: ID!, $limit: Int, $offset: Int) {
  messagesByRoom(roomId: $roomId, limit: $limit, offset: $offset) {
    id
    text
    messageType
    editedAt
    deletedAt
    createdAt
    userInfo {
      id
      username
      displayName
      avatarUrl
    }
    replyTo {
      id
      text
      userInfo {
        username
      }
    }
    reactions {
      id
      emoji
      user {
        username
      }
    }
    reactionCount {
      emoji
      count
    }
    metadata {
      imageUrl
      fileName
      fileSize
      mimeType
    }
  }
}
```

### Search Messages

Full-text search across messages.

```graphql
query SearchMessages($query: String!, $roomId: ID, $limit: Int, $offset: Int) {
  searchMessages(query: $query, roomId: $roomId, limit: $limit, offset: $offset) {
    id
    text
    messageType
    createdAt
    userInfo {
      username
      displayName
    }
    # Search results include highlighted matches
  }
}
```

### Message Reactions

Add emoji reactions to messages.

```graphql
mutation AddReaction {
  addReaction(input: {
    messageId: "message_12345"
    emoji: "👍"
  }) {
    id
    emoji
    createdAt
    user {
      id
      username
    }
  }
}

mutation RemoveReaction {
  removeReaction(input: {
    messageId: "message_12345"
    emoji: "👍"
  })
}
```

### Typing Indicators

Show when users are typing in a room.

```graphql
mutation StartTyping {
  startTyping(input: {
    roomId: "room_67890"
  })
}

mutation StopTyping {
  stopTyping(input: {
    roomId: "room_67890"
  })
}
```

## 🔄 Real-time Subscriptions

### Message Subscriptions

Listen for new messages in a room.

```graphql
subscription MessageAdded($roomId: ID!) {
  messageAddedToRoom(roomId: $roomId) {
    id
    text
    messageType
    createdAt
    userInfo {
      id
      username
      displayName
      avatarUrl
    }
    replyTo {
      id
      text
      userInfo {
        username
      }
    }
  }
}
```

### Message Updates

Listen for message edits and deletions.

```graphql
subscription MessageUpdated($roomId: ID!) {
  messageUpdated(roomId: $roomId) {
    id
    text
    editedAt
    deletedAt
    userInfo {
      username
    }
  }
}
```

### Reaction Events

Listen for reaction additions and removals.

```graphql
subscription ReactionAdded($roomId: ID!) {
  reactionAdded(roomId: $roomId) {
    id
    emoji
    user {
      username
      displayName
    }
    createdAt
  }
}

subscription ReactionRemoved($roomId: ID!) {
  reactionRemoved(roomId: $roomId) {
    id
    emoji
    user {
      username
    }
  }
}
```

### User Presence

Listen for user status changes in a room.

```graphql
subscription UserStatusChanged($roomId: ID!) {
  userStatusChanged(roomId: $roomId) {
    id
    username
    displayName
    status
    lastSeenAt
  }
}
```

### Typing Indicators

Listen for typing status in a room.

```graphql
subscription TypingStarted($roomId: ID!) {
  typingStarted(roomId: $roomId) {
    user {
      id
      username
      displayName
    }
    roomId
    startedAt
  }
}

subscription TypingStopped($roomId: ID!) {
  typingStopped(roomId: $roomId) {
    user {
      id
      username
    }
    roomId
  }
}
```

### Room Events

Listen for room membership changes.

```graphql
subscription MemberJoined($roomId: ID!) {
  roomMemberJoined(roomId: $roomId) {
    id
    role
    joinedAt
    user {
      id
      username
      displayName
      avatarUrl
    }
  }
}

subscription MemberLeft($roomId: ID!) {
  roomMemberLeft(roomId: $roomId) {
    user {
      id
      username
      displayName
    }
  }
}
```

## ⚠️ Error Handling

### Error Response Format

```json
{
  "errors": [
    {
      "message": "User is not a member of this room",
      "locations": [
        {
          "line": 2,
          "column": 3
        }
      ],
      "path": [
        "sendMessageToRoom"
      ],
      "extensions": {
        "code": "FORBIDDEN",
        "details": {
          "userId": "user_12345",
          "roomId": "room_67890"
        }
      }
    }
  ],
  "data": null
}
```

### Common Error Codes

- `UNAUTHENTICATED`: Missing or invalid JWT token
- `FORBIDDEN`: Insufficient permissions for operation
- `NOT_FOUND`: Requested resource doesn't exist
- `VALIDATION_ERROR`: Invalid input data
- `RATE_LIMITED`: Too many requests in time window
- `INTERNAL_ERROR`: Server-side error

### Error Handling Best Practices

```javascript
// Client-side error handling example
const handleGraphQLErrors = (errors) => {
  errors.forEach(error => {
    switch (error.extensions?.code) {
      case 'UNAUTHENTICATED':
        // Redirect to login
        window.location.href = '/login';
        break;
      case 'FORBIDDEN':
        // Show permission denied message
        showNotification('Permission denied', 'error');
        break;
      case 'VALIDATION_ERROR':
        // Show field-specific validation errors
        showValidationErrors(error.extensions.details);
        break;
      case 'RATE_LIMITED':
        // Show rate limit warning
        showNotification('Please slow down', 'warning');
        break;
      default:
        // Generic error handling
        showNotification(error.message, 'error');
    }
  });
};
```

## 📊 Rate Limiting

### Current Limits

- **Authentication**: 10 requests per minute per IP
- **Message Creation**: 60 messages per minute per user
- **Room Operations**: 30 requests per minute per user
- **Search**: 20 requests per minute per user

### Rate Limit Headers

```http
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1642234567
```

## 🎯 Best Practices

### Query Optimization

1. **Request Only Needed Fields**:
```graphql
# Good: Specific fields
query GetMessages {
  messagesByRoom(roomId: "room_123", limit: 20) {
    id
    text
    createdAt
    userInfo {
      username
      displayName
    }
  }
}

# Avoid: All fields when not needed
query GetMessages {
  messagesByRoom(roomId: "room_123", limit: 20) {
    # ... all fields
  }
}
```

2. **Use Pagination**:
```graphql
query GetMessages($offset: Int = 0) {
  messagesByRoom(roomId: "room_123", limit: 20, offset: $offset) {
    # ... fields
  }
}
```

3. **Batch Operations**:
```graphql
# Good: Single query for multiple rooms
query GetMultipleRooms($roomIds: [ID!]!) {
  rooms: roomIds @include(if: true) {
    id
    name
    memberCount
  }
}
```

### Subscription Management

1. **Clean Up Subscriptions**:
```javascript
// Unsubscribe when component unmounts
useEffect(() => {
  const subscription = subscribeToMessages(roomId);
  
  return () => {
    subscription.unsubscribe();
  };
}, [roomId]);
```

2. **Handle Connection States**:
```javascript
const subscription = subscribeToMessages(roomId, {
  onData: (data) => {
    // Handle new message
  },
  onError: (error) => {
    // Handle subscription errors
    console.error('Subscription error:', error);
  },
  onComplete: () => {
    // Handle subscription completion
  }
});
```

### Security Considerations

1. **Validate Permissions**: Always check user permissions on the client
2. **Sanitize Input**: Clean user input before display
3. **Handle Sensitive Data**: Never log or expose sensitive information
4. **Use HTTPS**: Always use secure connections in production

## 📚 Examples & Tutorials

### Complete Chat Implementation

```javascript
// React component example
import { useQuery, useMutation, useSubscription } from '@apollo/client';

const ChatRoom = ({ roomId }) => {
  // Get messages
  const { data: messagesData, loading } = useQuery(GET_MESSAGES, {
    variables: { roomId, limit: 50 }
  });

  // Send message mutation
  const [sendMessage] = useMutation(SEND_MESSAGE, {
    update: (cache, { data: { sendMessageToRoom } }) => {
      // Optimistically update cache
      const existingMessages = cache.readQuery({
        query: GET_MESSAGES,
        variables: { roomId, limit: 50 }
      });
      
      cache.writeQuery({
        query: GET_MESSAGES,
        variables: { roomId, limit: 50 },
        data: {
          messagesByRoom: [...existingMessages.messagesByRoom, sendMessageToRoom]
        }
      });
    }
  });

  // Subscribe to new messages
  useSubscription(MESSAGE_ADDED, {
    variables: { roomId },
    onData: ({ data }) => {
      // Handle real-time message
      console.log('New message:', data.messageAddedToRoom);
    }
  });

  // Subscribe to typing indicators
  useSubscription(TYPING_STARTED, {
    variables: { roomId },
    onData: ({ data }) => {
      setTypingUsers(prev => [...prev, data.typingStarted.user]);
    }
  });

  const handleSendMessage = async (text) => {
    try {
      await sendMessage({
        variables: { roomId, text }
      });
    } catch (error) {
      handleGraphQLErrors(error.graphQLErrors);
    }
  };

  if (loading) return <div>Loading messages...</div>;

  return (
    <div className="chat-room">
      <MessageList messages={messagesData?.messagesByRoom || []} />
      <MessageInput onSend={handleSendMessage} />
    </div>
  );
};
```

### GraphQL Queries & Mutations

```javascript
// GraphQL operations
const GET_MESSAGES = gql`
  query GetMessages($roomId: ID!, $limit: Int, $offset: Int) {
    messagesByRoom(roomId: $roomId, limit: $limit, offset: $offset) {
      id
      text
      messageType
      createdAt
      editedAt
      userInfo {
        id
        username
        displayName
        avatarUrl
      }
      replyTo {
        id
        text
        userInfo {
          username
        }
      }
      reactions {
        id
        emoji
        user {
          username
        }
      }
      reactionCount {
        emoji
        count
      }
    }
  }
`;

const SEND_MESSAGE = gql`
  mutation SendMessage($roomId: ID!, $text: String!) {
    sendMessageToRoom(roomId: $roomId, text: $text) {
      id
      text
      messageType
      createdAt
      userInfo {
        id
        username
        displayName
        avatarUrl
      }
    }
  }
`;

const MESSAGE_ADDED = gql`
  subscription MessageAdded($roomId: ID!) {
    messageAddedToRoom(roomId: $roomId) {
      id
      text
      messageType
      createdAt
      userInfo {
        id
        username
        displayName
        avatarUrl
      }
    }
  }
`;
```

This comprehensive API documentation provides everything needed to build a full-featured chat application client. The examples demonstrate real-world usage patterns and best practices for GraphQL development.