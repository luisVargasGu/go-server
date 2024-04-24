-- Insert fake data into Channels table
INSERT INTO Channels (Name)
VALUES 
    ('Channel 1'),
    ('Channel 2'),
    ('Channel 3');

-- Insert fake data into Rooms table
INSERT INTO Rooms (Name, ChannelID)
VALUES 
    ('Room 1', 1),
    ('Room 2', 1),
    ('Room 3', 2);

-- Insert fake data into Messages table
INSERT INTO Messages (RoomID, SenderID, Content, Timestamp, IsRead)
VALUES 
    (1, 1, 'Message 1', NOW(), true),
    (1, 2, 'Message 2', NOW(), true),
    (2, 1, 'Message 3', NOW(), false);

-- Insert fake data into ChannelsToUsers table
INSERT INTO ChannelsToUsers (user_id, channel_id)
VALUES 
    (1, 1),
    (2, 1),
    (3, 2);

-- Insert fake data into RoomsToChannels table
INSERT INTO RoomsToChannels (room_id, channel_id)
VALUES 
    (1, 1),
    (2, 1),
    (3, 2);

