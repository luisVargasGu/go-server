DO $$DECLARE
    r RECORD;
BEGIN
    FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = current_schema()) LOOP
        EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(r.tablename) || ' CASCADE';
    END LOOP;
END$$;

CREATE TABLE Users (
    ID INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    Username varchar(255) UNIQUE NOT NULL,
    Password varchar(255) NOT NULL,
    CreatedAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE Channels (
    ID SERIAL PRIMARY KEY,
    Name VARCHAR(255) NOT NULL
);

CREATE TABLE Rooms (
    ID SERIAL PRIMARY KEY,
    Name VARCHAR(255) NOT NULL,
    ChannelID INT NOT NULL,
    FOREIGN KEY (ChannelID) REFERENCES Channels(ID) ON DELETE CASCADE
);

CREATE TABLE Messages (
    ID SERIAL PRIMARY KEY,
    RoomID INT NOT NULL,
    SenderID INT NOT NULL, -- Assuming SenderID is a reference to Users.ID
    Content TEXT,
    Timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    IsRead BOOLEAN,
    FOREIGN KEY (RoomID) REFERENCES Rooms(ID) ON DELETE CASCADE,
    FOREIGN KEY (SenderID) REFERENCES Users(ID)
);

CREATE TABLE ChannelsToUsers (
    user_id INT,
    channel_id INT,
    PRIMARY KEY (user_id, channel_id),
    FOREIGN KEY (user_id) REFERENCES Users(ID) ON DELETE CASCADE,
    FOREIGN KEY (channel_id) REFERENCES Channels(ID) ON DELETE CASCADE
);

CREATE TABLE RoomsToChannels (
    room_id INT,
    channel_id INT,
    PRIMARY KEY (room_id, channel_id),
    FOREIGN KEY (room_id) REFERENCES Rooms(ID) ON DELETE CASCADE,
    FOREIGN KEY (channel_id) REFERENCES Channels(ID) ON DELETE CASCADE
);

CREATE INDEX idx_username_users ON Users (Username);
CREATE INDEX idx_room_id_rooms ON Rooms (ID);
CREATE INDEX idx_room_id_messages ON Messages (RoomID);

CREATE INDEX idx_user_id_users_to_channels ON ChannelsToUsers (user_id);
CREATE INDEX idx_room_id_rooms_to_channels ON RoomsToChannels (room_id);
CREATE INDEX idx_channel_id_rooms_to_channels ON RoomsToChannels (channel_id);
