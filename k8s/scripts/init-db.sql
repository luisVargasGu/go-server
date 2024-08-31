CREATE DATABASE chat_app;

\c chat_app;

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
    CreatedAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    Avatar BYTEA DEFAULT decode('iVBORw0KGgoAAAANSUhEUgAAAEsAAABNCAMAAADZyWnFAAAAOVBMVEXo6epXWFrs7e37+/zz8/Pv7/D+/v5ZWlz29vb5+fm/wMFiYmTh4uPV1teoqaptbnCFhoh5eXuVlpdyR7wwAAAD/ElEQVRYw61Yi5KDIAxETBQQ8fH/H3sJ0Ko14erNMXWcUdkmYVmSmMFbGMD2Pd2c9QMiPXGnJz1i7/gJ5idAT+h2n9Wj8R6cd0A38A7d6cYXWERjrKWftQAWXp/eZ3kwmjkDWrDmcxAg2+ktDmzreRb2BpxzfNUb1BuAUYa1+P70OsuI8WLHWoMtE+IFHwbR7xegbBzcZoE74tX7sljefDUQLrNyvHrPy4DeI68I3awx36LhxzoyVl+wenqK5sF4zSIOAuGQj8zEvNC0PvYJFC8pzxqyKcwJpl+hKAiEMvQoxpjvomV1Ml/mRFEnfBzDvm7btqYQlQU9ccK9hxSqsE5dGdOWZtGyA8Dk3cDAQqxiIqQxD4ZbgmjZaw8NJjtMlxCreR0rUsWbksyNGu4aL/Q3KDuv3QWKxhSkJci6xPEqtPL3YMX1E4nGkqIU/0L1iiUs4S5A0dglsIpVfBRWcBKxulGKGTOVdZVj/6WHGWyZpfAXXWVOfG0Wg+2CYS7rBJNMwEoalGyYzVxlFRX+aFWxxjEJ/110FbzwKm46luik5b1tRax5aWBt0jYvuioJTRNLChipDxpZ3+NjLMu6KopcXJ9iGdZVWTD3p/EiJ80gnhY26FDiOuakQzkuWqQI8hRQzgSTGsHfxXgZb+QcpMWJsUuiAahgtVxUndR8bHFimhUkBavBifEfsRSu6j62gr+bhyPptJ/NMx/VHdlNQU2G1Bey4ms7qI1FeYkU9zTrSDqWFSzrdv17jfdlLW9YnbJ9ql0trDsv2litDPW2lLSI+tde09X8R7cDV5HUenYruqqIa7e2kmoD+GAhW+RCp5xD5e0nVpOodA559z922VzDNJh/o9eqxh4856samJ23OyeSahf5iODlEzIGSfMn5RCyNV/FO4PjzBWHqBMLlzNWzL+EvDCGtC5jpyWsVM7sYb7C2aOGubiWNq6BGkcav10Y7pqvDh959JzWX4AucNG+82j3rmEq0q67Jsr1WipByr1qDeNr3THv0wOgYt24hVPdUWoYAiU2PUN6q5B91chDLZgxjH8ZvBVK/8jjq370PnV/A9sM3HpDzTSp5ePRG3pXy6CXQC2oBEe9feoNPQcjKDz3hk79ifDQTXawluuu1DCn3tBDXizh6Chh7Q29+zkQ9/FbsG5c59zIefWG8KM35IES6O4rXpEswrs3BLU39GpWVmvj+sVWYqMus269oWorJTi/aA5tQ/THLHf0hm6dvxgatrEWBuilfqGThsc5LeNdyKpsWS/OOnpD1/5qP2SBHbvzGFmdSVkGL8zKZ20hm78GLoeABTu3v7ZlW9c9BYr3JUyXOINp96M9kcVENJHqORwGlie1H21/AGSVWyPVNNOeAAAAAElFTkSuQmCC', 'base64'),
);

CREATE TABLE Channels (
    ID SERIAL PRIMARY KEY,
    Name VARCHAR(255) NOT NULL,
    Avatar BYTEA DEFAULT decode('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', 'base64')
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

CREATE TABLE Invites (
	ID SERIAL PRIMARY KEY,
	ChannelID INT NOT NULL,
	InviterID INT NOT NULL,
	InviteeID INT NOT NULL DEFAULT -1,
	InviteCode VARCHAR(255),
	Expiration TIMESTAMP NOT NULL,
	CreatedAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	CONSTRAINT fk_channel
		FOREIGN KEY (ChannelID) REFERENCES Channels(ID),
	CONSTRAINT fk_inviter
		FOREIGN KEY (InviterID) REFERENCES Users(ID),
	CONSTRAINT fk_invitee
		FOREIGN KEY (InviteeID) REFERENCES Users(ID),
	UNIQUE (ChannelID, InviteeID)
);

CREATE TABLE SeenMessages (
	user_id INT NOT NULL,
	message_id INT NOT NULL,
	seen_time_stamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (user_id, message_id),
	FOREIGN KEY (user_id) REFERENCES Users(ID),
	FOREIGN KEY (message_id) REFERENCES Messages(ID)
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

CREATE INDEX idx_channel ON Invites(ChannelID);
CREATE INDEX idx_inviter ON Invites(InviterID);
CREATE INDEX idx_invitee ON Invites(InviteeID);
CREATE INDEX idx_expiration ON Invites(Expiration);

CREATE USER admin WITH PASSWORD 'password';

GRANT ALL PRIVILEGES ON DATABASE chat_app TO admin;
